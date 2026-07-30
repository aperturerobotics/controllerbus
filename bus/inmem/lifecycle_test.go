package inmem

import (
	"context"
	stderrors "errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/aperturerobotics/controllerbus/bus"
	cbcontroller "github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	directivecontroller "github.com/aperturerobotics/controllerbus/directive/controller"
	"github.com/sirupsen/logrus"
)

type lifecycleController struct {
	executeFn   func(context.Context) error
	closeFn     func() error
	handleFn    func(context.Context, directive.Instance) ([]directive.Resolver, error)
	closeCalls  atomic.Int32
	handleCalls atomic.Int32
}

func (c *lifecycleController) Execute(ctx context.Context) error {
	if c.executeFn == nil {
		return nil
	}
	return c.executeFn(ctx)
}

func (c *lifecycleController) Close() error {
	c.closeCalls.Add(1)
	if c.closeFn == nil {
		return nil
	}
	return c.closeFn()
}

func (c *lifecycleController) HandleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	c.handleCalls.Add(1)
	if c.handleFn == nil {
		return nil, nil
	}
	return c.handleFn(ctx, di)
}

func (c *lifecycleController) GetControllerInfo() *cbcontroller.Info { return nil }

type trackingDirectiveController struct {
	directive.Controller
	addErr       error
	beforeAdd    func(directive.Handler)
	detachCalls  atomic.Int32
	detached     chan struct{}
	detachedOnce sync.Once
}

func (c *trackingDirectiveController) AddHandler(handler directive.Handler) (func(), error) {
	if c.beforeAdd != nil {
		c.beforeAdd(handler)
	}
	if c.addErr != nil {
		return nil, c.addErr
	}

	var release func()
	if c.Controller != nil {
		var err error
		release, err = c.Controller.AddHandler(handler)
		if err != nil {
			return nil, err
		}
	} else {
		release = func() {}
	}
	return func() {
		c.detachCalls.Add(1)
		c.detachedOnce.Do(func() { close(c.detached) })
		release()
	}, nil
}

func newTrackingBus() (*Bus, *trackingDirectiveController) {
	logger := logrus.New()
	dc := directivecontroller.NewController(context.Background(), logrus.NewEntry(logger))
	tracking := &trackingDirectiveController{
		Controller: dc,
		detached:   make(chan struct{}),
	}
	return NewBus(tracking), tracking
}

func TestAddControllerReleaseWaitsForExecute(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		b, tracking := newTrackingBus()
		executionStarted := make(chan struct{})
		allowExecuteReturn := make(chan struct{})
		releaseReturned := make(chan struct{})
		var executeReturned atomic.Bool
		var closeBeforeExecute atomic.Bool
		var callbackCalls atomic.Int32
		var callbackErr error

		ctrl := &lifecycleController{
			executeFn: func(ctx context.Context) error {
				close(executionStarted)
				<-ctx.Done()
				<-allowExecuteReturn
				executeReturned.Store(true)
				return nil
			},
			closeFn: func() error {
				if !executeReturned.Load() {
					closeBeforeExecute.Store(true)
				}
				// A Close implementation may call back into the bus.
				_ = b.GetControllers()
				return nil
			},
		}
		release, err := b.AddController(t.Context(), ctrl, func(err error) {
			callbackErr = err
			callbackCalls.Add(1)
		})
		if err != nil {
			t.Fatalf("AddController failed: %v", err)
		}
		<-executionStarted

		go func() {
			release()
			close(releaseReturned)
		}()
		<-tracking.detached
		synctest.Wait()

		releasedEarly := false
		select {
		case <-releaseReturned:
			releasedEarly = true
		default:
		}
		closeCallsBeforeReturn := ctrl.closeCalls.Load()
		controllersAfterDetach := len(b.GetControllers())

		close(allowExecuteReturn)
		synctest.Wait()
		if releasedEarly {
			t.Fatal("release returned before Execute returned")
		}
		if closeCallsBeforeReturn != 0 {
			t.Fatalf("Close called before Execute returned: calls = %d", closeCallsBeforeReturn)
		}
		if controllersAfterDetach != 0 {
			t.Fatalf("detached controller count = %d, want 0", controllersAfterDetach)
		}
		if !executeReturned.Load() {
			t.Fatal("Execute return marker was not set")
		}
		if closeBeforeExecute.Load() {
			t.Fatal("Close observed Execute before its return marker")
		}
		select {
		case <-releaseReturned:
		default:
			t.Fatal("release did not return after Execute returned")
		}
		if got := ctrl.closeCalls.Load(); got != 1 {
			t.Fatalf("Close calls = %d, want 1", got)
		}
		if got := tracking.detachCalls.Load(); got != 1 {
			t.Fatalf("handler detach calls = %d, want 1", got)
		}
		if got := callbackCalls.Load(); got != 1 {
			t.Fatalf("callback calls = %d, want 1", got)
		}
		if callbackErr != nil {
			t.Fatalf("callback error = %v, want nil", callbackErr)
		}
	})
}

func TestAddControllerReleaseRacesWithExecuteError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		b, tracking := newTrackingBus()
		executeErr := stderrors.New("terminal execution failure")
		closeErr := stderrors.New("terminal close failure")
		executionStarted := make(chan struct{})
		raceGate := make(chan struct{})
		releaseReady := make(chan struct{})
		releaseReturned := make(chan struct{})
		callbackResults := make(chan error, 2)
		var callbackCalls atomic.Int32

		ctrl := &lifecycleController{
			executeFn: func(context.Context) error {
				close(executionStarted)
				<-raceGate
				return executeErr
			},
			closeFn: func() error { return closeErr },
		}
		release, err := b.AddController(t.Context(), ctrl, func(err error) {
			callbackCalls.Add(1)
			callbackResults <- err
		})
		if err != nil {
			t.Fatalf("AddController failed: %v", err)
		}
		<-executionStarted
		go func() {
			close(releaseReady)
			<-raceGate
			release()
			close(releaseReturned)
		}()
		<-releaseReady

		close(raceGate)
		synctest.Wait()
		select {
		case <-releaseReturned:
		default:
			t.Fatal("racing release did not return")
		}
		if got := ctrl.closeCalls.Load(); got != 1 {
			t.Fatalf("Close calls after racing finalizers = %d, want 1", got)
		}
		if got := tracking.detachCalls.Load(); got != 1 {
			t.Fatalf("handler detach calls after racing finalizers = %d, want 1", got)
		}
		if got := callbackCalls.Load(); got != 1 {
			t.Fatalf("callback calls after racing finalizers = %d, want 1", got)
		}
		select {
		case got := <-callbackResults:
			if !stderrors.Is(got, executeErr) {
				t.Fatalf("callback error = %v, want execution error", got)
			}
			var typedCloseErr *bus.ControllerCloseError
			if !stderrors.As(got, &typedCloseErr) || !stderrors.Is(typedCloseErr, closeErr) {
				t.Fatalf("callback error = %v, want typed close error", got)
			}
		default:
			t.Fatal("callback did not receive execution error")
		}

		release()
		if got := ctrl.closeCalls.Load(); got != 1 {
			t.Fatalf("Close calls after repeated release = %d, want 1", got)
		}
	})
}

func TestAddControllerReleaseWaitsForContextIgnoringExecute(t *testing.T) {
	const observerTimeout = 75 * time.Millisecond
	const completionTimeout = time.Second

	b, tracking := newTrackingBus()
	executionStarted := make(chan struct{})
	allowExecuteReturn := make(chan struct{})
	releaseReturned := make(chan struct{})
	var executeReturned atomic.Bool
	var closeBeforeExecute atomic.Bool

	ctrl := &lifecycleController{
		executeFn: func(context.Context) error {
			close(executionStarted)
			<-allowExecuteReturn
			executeReturned.Store(true)
			return nil
		},
		closeFn: func() error {
			if !executeReturned.Load() {
				closeBeforeExecute.Store(true)
			}
			return nil
		},
	}
	release, err := b.AddController(t.Context(), ctrl, nil)
	if err != nil {
		t.Fatalf("AddController failed: %v", err)
	}
	<-executionStarted
	go func() {
		release()
		close(releaseReturned)
	}()

	select {
	case <-tracking.detached:
	case <-time.After(completionTimeout):
		t.Fatal("handler was not detached after release request")
	}
	select {
	case <-releaseReturned:
		t.Fatal("release returned while Execute ignored cancellation")
	case <-time.After(observerTimeout):
	}
	if got := ctrl.closeCalls.Load(); got != 0 {
		t.Fatalf("Close calls while Execute was blocked = %d, want 0", got)
	}

	close(allowExecuteReturn)
	select {
	case <-releaseReturned:
	case <-time.After(completionTimeout):
		t.Fatal("release did not return after Execute gate opened")
	}
	if !executeReturned.Load() {
		t.Fatal("Execute return marker was not set")
	}
	if closeBeforeExecute.Load() {
		t.Fatal("Close ran before context-ignoring Execute returned")
	}
	if got := ctrl.closeCalls.Load(); got != 1 {
		t.Fatalf("Close calls = %d, want 1", got)
	}
}

func TestAddControllerLifecycleOutcomes(t *testing.T) {
	t.Run("add failure closes and reports close error", func(t *testing.T) {
		addErr := stderrors.New("handler registration failed")
		closeErr := stderrors.New("failed attachment close failed")
		tracking := &trackingDirectiveController{addErr: addErr, detached: make(chan struct{})}
		b := NewBus(tracking)
		ctrl := &lifecycleController{closeFn: func() error { return closeErr }}

		release, err := b.AddController(t.Context(), ctrl, nil)
		if release != nil {
			t.Fatal("release function is non-nil after add failure")
		}
		if !stderrors.Is(err, addErr) {
			t.Fatalf("AddController error = %v, want add error", err)
		}
		var typedCloseErr *bus.ControllerCloseError
		if !stderrors.As(err, &typedCloseErr) || !stderrors.Is(typedCloseErr, closeErr) {
			t.Fatalf("AddController error = %v, want typed close error", err)
		}
		if got := ctrl.closeCalls.Load(); got != 1 {
			t.Fatalf("Close calls after add failure = %d, want 1", got)
		}
	})

	t.Run("nil return remains attached until release", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			b, tracking := newTrackingBus()
			executed := make(chan struct{})
			callbackResults := make(chan error, 1)
			ctrl := &lifecycleController{executeFn: func(context.Context) error {
				close(executed)
				return nil
			}}
			release, err := b.AddController(t.Context(), ctrl, func(err error) { callbackResults <- err })
			if err != nil {
				t.Fatalf("AddController failed: %v", err)
			}
			<-executed
			synctest.Wait()
			if got := len(b.GetControllers()); got != 1 {
				t.Fatalf("controller count after nil Execute = %d, want 1", got)
			}
			release()
			if got := ctrl.closeCalls.Load(); got != 1 {
				t.Fatalf("Close calls = %d, want 1", got)
			}
			if got := tracking.detachCalls.Load(); got != 1 {
				t.Fatalf("handler detach calls = %d, want 1", got)
			}
			if got := <-callbackResults; got != nil {
				t.Fatalf("callback error = %v, want nil", got)
			}
		})
	})

	t.Run("execution error finalizes and reaches callback", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			b, tracking := newTrackingBus()
			executeErr := stderrors.New("execute failed")
			callbackResults := make(chan error, 1)
			ctrl := &lifecycleController{executeFn: func(context.Context) error { return executeErr }}
			_, err := b.AddController(t.Context(), ctrl, func(err error) { callbackResults <- err })
			if err != nil {
				t.Fatalf("AddController failed: %v", err)
			}
			synctest.Wait()
			if got := <-callbackResults; !stderrors.Is(got, executeErr) {
				t.Fatalf("callback error = %v, want execution error", got)
			}
			if got := ctrl.closeCalls.Load(); got != 1 {
				t.Fatalf("Close calls = %d, want 1", got)
			}
			if got := tracking.detachCalls.Load(); got != 1 {
				t.Fatalf("handler detach calls = %d, want 1", got)
			}
		})
	})

	t.Run("panic finalizes and reaches callback", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			b, tracking := newTrackingBus()
			callbackResults := make(chan error, 1)
			ctrl := &lifecycleController{executeFn: func(context.Context) error { panic("execute panic") }}
			_, err := b.AddController(t.Context(), ctrl, func(err error) { callbackResults <- err })
			if err != nil {
				t.Fatalf("AddController failed: %v", err)
			}
			synctest.Wait()
			got := <-callbackResults
			if got == nil || !strings.Contains(got.Error(), "controller panicked") {
				t.Fatalf("callback error = %v, want recovered panic", got)
			}
			if got := ctrl.closeCalls.Load(); got != 1 {
				t.Fatalf("Close calls = %d, want 1", got)
			}
			if got := tracking.detachCalls.Load(); got != 1 {
				t.Fatalf("handler detach calls = %d, want 1", got)
			}
		})
	})

	t.Run("close error is typed in callback", func(t *testing.T) {
		b, _ := newTrackingBus()
		closeErr := stderrors.New("close failed")
		executed := make(chan struct{})
		callbackResults := make(chan error, 1)
		ctrl := &lifecycleController{
			executeFn: func(context.Context) error { close(executed); return nil },
			closeFn:   func() error { return closeErr },
		}
		release, err := b.AddController(t.Context(), ctrl, func(err error) { callbackResults <- err })
		if err != nil {
			t.Fatalf("AddController failed: %v", err)
		}
		<-executed
		release()
		got := <-callbackResults
		var typedCloseErr *bus.ControllerCloseError
		if !stderrors.As(got, &typedCloseErr) || !stderrors.Is(typedCloseErr, closeErr) {
			t.Fatalf("callback error = %v, want typed close error", got)
		}
	})

	t.Run("repeated release finalizes once", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			b, tracking := newTrackingBus()
			executed := make(chan struct{})
			var callbackCalls atomic.Int32
			ctrl := &lifecycleController{executeFn: func(context.Context) error { close(executed); return nil }}
			release, err := b.AddController(t.Context(), ctrl, func(error) { callbackCalls.Add(1) })
			if err != nil {
				t.Fatalf("AddController failed: %v", err)
			}
			<-executed
			synctest.Wait()
			go release()
			go release()
			synctest.Wait()
			if got := ctrl.closeCalls.Load(); got != 1 {
				t.Fatalf("Close calls after repeated release = %d, want 1", got)
			}
			if got := tracking.detachCalls.Load(); got != 1 {
				t.Fatalf("handler detach calls after repeated release = %d, want 1", got)
			}
			if got := callbackCalls.Load(); got != 1 {
				t.Fatalf("callback calls after repeated release = %d, want 1", got)
			}
		})
	})
}

func TestExecuteControllerLifecycle(t *testing.T) {
	t.Run("nil return requires removal", func(t *testing.T) {
		b, tracking := newTrackingBus()
		ctrl := &lifecycleController{}
		if err := b.ExecuteController(t.Context(), ctrl); err != nil {
			t.Fatalf("ExecuteController failed: %v", err)
		}
		if got := len(b.GetControllers()); got != 1 {
			t.Fatalf("controller count after nil Execute = %d, want 1", got)
		}
		if got := ctrl.closeCalls.Load(); got != 0 {
			t.Fatalf("Close calls before removal = %d, want 0", got)
		}
		b.RemoveController(ctrl)
		if got := ctrl.closeCalls.Load(); got != 1 {
			t.Fatalf("Close calls after removal = %d, want 1", got)
		}
		if got := tracking.detachCalls.Load(); got != 1 {
			t.Fatalf("handler detach calls = %d, want 1", got)
		}
	})

	t.Run("execution and close errors are returned", func(t *testing.T) {
		b, tracking := newTrackingBus()
		executeErr := stderrors.New("sync execute failed")
		closeErr := stderrors.New("sync close failed")
		ctrl := &lifecycleController{
			executeFn: func(context.Context) error { return executeErr },
			closeFn:   func() error { return closeErr },
		}
		err := b.ExecuteController(t.Context(), ctrl)
		if !stderrors.Is(err, executeErr) {
			t.Fatalf("ExecuteController error = %v, want execution error", err)
		}
		var typedCloseErr *bus.ControllerCloseError
		if !stderrors.As(err, &typedCloseErr) || !stderrors.Is(typedCloseErr, closeErr) {
			t.Fatalf("ExecuteController error = %v, want typed close error", err)
		}
		if got := ctrl.closeCalls.Load(); got != 1 {
			t.Fatalf("Close calls = %d, want 1", got)
		}
		if got := tracking.detachCalls.Load(); got != 1 {
			t.Fatalf("handler detach calls = %d, want 1", got)
		}
	})

	t.Run("concurrent removal cancels and waits", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			b, tracking := newTrackingBus()
			executionStarted := make(chan struct{})
			cancellationObserved := make(chan struct{})
			allowExecuteReturn := make(chan struct{})
			executeReturned := make(chan error, 1)
			removeReturned := make(chan struct{})
			ctrl := &lifecycleController{executeFn: func(ctx context.Context) error {
				close(executionStarted)
				<-ctx.Done()
				close(cancellationObserved)
				<-allowExecuteReturn
				return nil
			}}

			go func() {
				executeReturned <- b.ExecuteController(t.Context(), ctrl)
			}()
			<-executionStarted
			go func() {
				b.RemoveController(ctrl)
				close(removeReturned)
			}()
			<-cancellationObserved
			<-tracking.detached
			synctest.Wait()

			executeReturnedEarly := false
			select {
			case <-executeReturned:
				executeReturnedEarly = true
			default:
			}
			removeReturnedEarly := false
			select {
			case <-removeReturned:
				removeReturnedEarly = true
			default:
			}
			closeCallsBeforeReturn := ctrl.closeCalls.Load()

			close(allowExecuteReturn)
			synctest.Wait()
			if executeReturnedEarly {
				t.Fatal("ExecuteController returned before controller Execute returned")
			}
			if removeReturnedEarly {
				t.Fatal("RemoveController returned before controller Execute returned")
			}
			if closeCallsBeforeReturn != 0 {
				t.Fatalf("Close calls before Execute returned = %d, want 0", closeCallsBeforeReturn)
			}
			select {
			case err := <-executeReturned:
				if err != nil {
					t.Fatalf("ExecuteController failed: %v", err)
				}
			default:
				t.Fatal("ExecuteController did not return after controller Execute returned")
			}
			select {
			case <-removeReturned:
			default:
				t.Fatal("RemoveController did not return after controller Execute returned")
			}
			if got := ctrl.closeCalls.Load(); got != 1 {
				t.Fatalf("Close calls = %d, want 1", got)
			}
			if got := tracking.detachCalls.Load(); got != 1 {
				t.Fatalf("handler detach calls = %d, want 1", got)
			}
		})
	})
}

func TestRemoveControllerWaitsForExecute(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		b, tracking := newTrackingBus()
		executionStarted := make(chan struct{})
		cancellationObserved := make(chan struct{})
		allowExecuteReturn := make(chan struct{})
		removeReturned := make(chan struct{})
		var executeReturned atomic.Bool
		ctrl := &lifecycleController{
			executeFn: func(ctx context.Context) error {
				close(executionStarted)
				<-ctx.Done()
				close(cancellationObserved)
				<-allowExecuteReturn
				executeReturned.Store(true)
				return nil
			},
			closeFn: func() error {
				if !executeReturned.Load() {
					t.Error("Close called before Execute returned")
				}
				return nil
			},
		}
		release, err := b.AddController(t.Context(), ctrl, nil)
		if err != nil {
			t.Fatalf("AddController failed: %v", err)
		}
		<-executionStarted
		go func() {
			b.RemoveController(ctrl)
			close(removeReturned)
		}()
		<-cancellationObserved
		<-tracking.detached
		synctest.Wait()
		removeReturnedEarly := false
		select {
		case <-removeReturned:
			removeReturnedEarly = true
		default:
		}
		closeCallsBeforeReturn := ctrl.closeCalls.Load()

		close(allowExecuteReturn)
		synctest.Wait()
		if removeReturnedEarly {
			t.Fatal("RemoveController returned before Execute returned")
		}
		if closeCallsBeforeReturn != 0 {
			t.Fatalf("Close calls before Execute returned = %d, want 0", closeCallsBeforeReturn)
		}
		select {
		case <-removeReturned:
		default:
			t.Fatal("RemoveController did not return after Execute returned")
		}
		if got := ctrl.closeCalls.Load(); got != 1 {
			t.Fatalf("Close calls = %d, want 1", got)
		}
		release()
		if got := ctrl.closeCalls.Load(); got != 1 {
			t.Fatalf("Close calls after release following removal = %d, want 1", got)
		}
	})
}

func TestAddControllerCallsControllerOutsideBusMutex(t *testing.T) {
	const timeout = time.Second

	tracking := &trackingDirectiveController{detached: make(chan struct{})}
	b := NewBus(tracking)
	handlerReturned := make(chan struct{})
	tracking.beforeAdd = func(handler directive.Handler) {
		_, _ = handler.HandleDirective(t.Context(), nil)
		close(handlerReturned)
	}
	ctrl := &lifecycleController{handleFn: func(context.Context, directive.Instance) ([]directive.Resolver, error) {
		_ = b.GetControllers()
		return nil, nil
	}}
	type addResult struct {
		release func()
		err     error
	}
	result := make(chan addResult, 1)
	go func() {
		release, err := b.AddController(t.Context(), ctrl, nil)
		result <- addResult{release: release, err: err}
	}()

	var got addResult
	select {
	case got = <-result:
	case <-time.After(timeout):
		t.Fatal("AddController held the bus mutex while calling HandleDirective")
	}
	if got.err != nil {
		t.Fatalf("AddController failed: %v", got.err)
	}
	select {
	case <-handlerReturned:
	default:
		t.Fatal("HandleDirective was not called during registration")
	}
	got.release()
	if calls := ctrl.handleCalls.Load(); calls != 1 {
		t.Fatalf("HandleDirective calls = %d, want 1", calls)
	}
}
