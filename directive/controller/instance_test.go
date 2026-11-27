package controller_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/controllerbus/directive/controller"
	directive_mock "github.com/aperturerobotics/controllerbus/directive/mock"
	"github.com/sirupsen/logrus"
)

func TestDisposeSkipsWeakRefReleasedByValueRemovedCallback(t *testing.T) {
	var releaseWeak atomic.Pointer[func()]
	readyCh := make(chan struct{})
	res := &releaseWeakRefResolver{
		releaseWeak: &releaseWeak,
		readyCh:     readyCh,
	}
	ctrl := controller.NewController(context.Background(), logrus.NewEntry(logrus.New()))
	removeHandler, err := ctrl.AddHandler(directive.NewFuncHandler(func(context.Context, directive.Instance) ([]directive.Resolver, error) {
		return []directive.Resolver{res}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer removeHandler()

	di, strongRef, err := ctrl.AddDirective(context.Background(), &directive_mock.MockDirective{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	<-readyCh

	weakHandler := &countingRefHandler{}
	weakRef := di.AddReference(context.Background(), weakHandler, true)
	releaseWeakFunc := weakRef.Release
	releaseWeak.Store(&releaseWeakFunc)

	strongRef.Release()

	if got := weakHandler.removed.Load(); got != 0 {
		t.Fatalf("released weak ref received %d value removal callbacks", got)
	}
}

type releaseWeakRefResolver struct {
	releaseWeak *atomic.Pointer[func()]
	readyOnce   sync.Once
	readyCh     chan struct{}
}

func (r *releaseWeakRefResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	valueID, accepted := handler.AddValue("value")
	if accepted {
		handler.AddValueRemovedCallback(valueID, func() {
			releaseWeak := r.releaseWeak.Load()
			if releaseWeak != nil && *releaseWeak != nil {
				(*releaseWeak)()
			}
		})
	}
	handler.MarkIdle(true)
	r.readyOnce.Do(func() {
		close(r.readyCh)
	})
	<-ctx.Done()
	return context.Canceled
}

type countingRefHandler struct {
	removed atomic.Int32
}

func (h *countingRefHandler) HandleValueAdded(directive.Instance, directive.AttachedValue) {}

func (h *countingRefHandler) HandleValueRemoved(directive.Instance, directive.AttachedValue) {
	h.removed.Add(1)
}

func (h *countingRefHandler) HandleInstanceDisposed(directive.Instance) {}

// _ is a type assertion.
var _ directive.Resolver = ((*releaseWeakRefResolver)(nil))

// _ is a type assertion.
var _ directive.ReferenceHandler = ((*countingRefHandler)(nil))

// TestRemovedResolverClearsErrors proves both watcher APIs retire a failed
// resolver even when its removal does not change the directive's idle state.
func TestRemovedResolverClearsErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	ctrl := controller.NewController(ctx, logrus.NewEntry(logrus.New()))
	di, ref, err := ctrl.AddDirective(ctx, &directive_mock.MockDirective{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ref.Release()
	failure := errors.New("host unavailable")
	idleErrors := make(chan bool, 16)
	stateErrors := make(chan bool, 16)
	defer di.AddIdleCallback(func(_ bool, errs []error) { idleErrors <- len(errs) != 0 })()
	defer di.AddStateCallback(func(_ bool, errs []error, _ []directive.AttachedValue) { stateErrors <- len(errs) != 0 })()

	remove, err := ctrl.AddHandler(directive.NewFuncHandler(func(context.Context, directive.Instance) ([]directive.Resolver, error) {
		return []directive.Resolver{directive.NewFuncResolver(func(context.Context, directive.ResolverHandler) error { return failure })}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer remove()
	await := func(ch <-chan bool, failed bool) {
		t.Helper()
		for {
			select {
			case got := <-ch:
				if got == failed {
					return
				}
			case <-ctx.Done():
				t.Fatalf("watcher did not publish failed=%v", failed)
			}
		}
	}
	await(idleErrors, true)
	await(stateErrors, true)
	remove()
	await(idleErrors, false)
	await(stateErrors, false)
}

// TestResolverAttachedWhileFullStartsAfterRemoval proves a resolver attached
// while the directive holds its maximum values starts once a value goes, so a
// handler replacing another never leaves the directive without one.
func TestResolverAttachedWhileFullStartsAfterRemoval(t *testing.T) {
	// Fill a directive capped at one value with the first handler's value.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	ctrl := controller.NewController(ctx, logrus.NewEntry(logrus.New()))
	valueHandler := func(val string) directive.Handler {
		return directive.NewFuncHandler(func(context.Context, directive.Instance) ([]directive.Resolver, error) {
			return directive.R(directive.NewValueResolver([]string{val}), nil)
		})
	}
	removeFirst, err := ctrl.AddHandler(valueHandler("first"))
	if err != nil {
		t.Fatal(err)
	}
	dir := &directive_mock.MockDirective{ValueOpts: directive.ValueOptions{MaxValueCount: 1, MaxValueHardCap: true}}
	di, ref, err := ctrl.AddDirective(ctx, dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ref.Release()
	values := make(chan []directive.AttachedValue, 16)
	defer di.AddStateCallback(func(_ bool, _ []error, vals []directive.AttachedValue) { values <- vals })()
	awaitValue := func(want string) {
		t.Helper()
		for {
			select {
			case vals := <-values:
				if len(vals) == 1 && vals[0].GetValue() == want {
					return
				}
			case <-ctx.Done():
				t.Fatalf("directive did not hold %q", want)
			}
		}
	}
	awaitValue("first")

	// Attach the second handler while full, then remove the first.
	removeSecond, err := ctrl.AddHandler(valueHandler("second"))
	if err != nil {
		t.Fatal(err)
	}
	defer removeSecond()
	removeFirst()
	awaitValue("second")
}

// TestAccessResolverReleasesRemovedValueOnce removes an acquisition before its
// removal callback is registered and requires cancellation to release it once.
func TestAccessResolverReleasesRemovedValueOnce(t *testing.T) {
	// Hold value publication while its directive is removed.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	ctrl := controller.NewController(ctx, logrus.NewEntry(logrus.New()))
	added := make(chan struct{})
	resume := make(chan struct{})
	done := make(chan error, 1)
	var releases atomic.Int32
	access := directive.NewAccessResolver(func(context.Context, func()) (string, func(), error) {
		return "value", func() { releases.Add(1) }, nil
	})
	remove, err := ctrl.AddHandler(directive.NewFuncHandler(func(context.Context, directive.Instance) ([]directive.Resolver, error) {
		return directive.R(directive.NewFuncResolver(func(ctx context.Context, handler directive.ResolverHandler) error {
			err := access.Resolve(ctx, handler)
			done <- err
			return err
		}), nil)
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer remove()

	// Pause AddValue before AccessResolver can register its removal callback.
	handler := directive.NewCallbackHandler(func(directive.AttachedValue) {
		close(added)
		select {
		case <-resume:
		case <-ctx.Done():
		}
	}, nil, nil)
	_, ref, err := ctrl.AddDirective(ctx, &directive_mock.MockDirective{}, handler)
	if err != nil {
		t.Fatal(err)
	}
	defer ref.Release()
	select {
	case <-added:
	case <-ctx.Done():
		t.Fatal("access value was not published")
	}

	// Remove the value, then let late callback registration observe removal.
	ref.Release()
	close(resume)
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("resolver returned %v, want cancellation", err)
		}
	case <-ctx.Done():
		t.Fatal("access resolver did not finish")
	}
	if got := releases.Load(); got != 1 {
		t.Fatalf("acquisition released %d times, want once", got)
	}
}

// TestAccessResolverReplacesInvalidatedValue retires an invalidated acquisition
// before publishing its replacement, then releases that replacement on removal.
func TestAccessResolverReplacesInvalidatedValue(t *testing.T) {
	// Record acquisitions and releases from the real directive controller.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	ctrl := controller.NewController(ctx, logrus.NewEntry(logrus.New()))
	invalidators := make(chan func(), 2)
	values := make(chan int32, 2)
	released := make(chan int32, 2)
	done := make(chan error, 1)
	var acquired, releaseCount atomic.Int32
	var acquiredBeforeRelease atomic.Bool
	access := directive.NewAccessResolver(func(_ context.Context, invalidate func()) (int32, func(), error) {
		id := acquired.Add(1)
		if releaseCount.Load() != id-1 {
			acquiredBeforeRelease.Store(true)
		}
		invalidators <- invalidate
		return id, func() {
			releaseCount.Add(1)
			released <- id
		}, nil
	})
	remove, err := ctrl.AddHandler(directive.NewFuncHandler(func(context.Context, directive.Instance) ([]directive.Resolver, error) {
		return directive.R(directive.NewFuncResolver(func(ctx context.Context, handler directive.ResolverHandler) error {
			err := access.Resolve(ctx, handler)
			done <- err
			return err
		}), nil)
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer remove()

	// Observe the first publication before invalidating its acquisition twice.
	handler := directive.NewCallbackHandler(func(value directive.AttachedValue) {
		values <- value.GetValue().(int32)
	}, nil, nil)
	_, ref, err := ctrl.AddDirective(ctx, &directive_mock.MockDirective{}, handler)
	if err != nil {
		t.Fatal(err)
	}
	defer ref.Release()
	select {
	case id := <-values:
		if id != 1 {
			t.Fatalf("first acquisition = %d, want 1", id)
		}
	case <-ctx.Done():
		t.Fatal("first acquisition was not published")
	}
	var invalidate func()
	select {
	case invalidate = <-invalidators:
	case <-ctx.Done():
		t.Fatal("getter did not provide its invalidator")
	}
	invalidate()
	invalidate()

	// A replacement is published only after the old acquisition is released.
	select {
	case id := <-values:
		if id != 2 {
			t.Fatalf("replacement acquisition = %d, want 2", id)
		}
	case <-ctx.Done():
		t.Fatal("replacement was not published")
	}
	if acquiredBeforeRelease.Load() {
		t.Fatal("replacement was acquired before the prior release")
	}
	select {
	case id := <-released:
		if id != 1 {
			t.Fatalf("retired acquisition = %d, want 1", id)
		}
	case <-ctx.Done():
		t.Fatal("prior acquisition was not released")
	}

	// Exercise notifications retained after their replacement was published.
	invalidate()
	invalidate()

	// Join cancellation and require exactly one release of the replacement.
	ref.Release()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("resolver returned %v, want cancellation", err)
		}
	case <-ctx.Done():
		t.Fatal("replacement resolver did not finish")
	}
	select {
	case id := <-released:
		if id != 2 {
			t.Fatalf("last released acquisition = %d, want 2", id)
		}
	case <-ctx.Done():
		t.Fatal("replacement acquisition was not released")
	}
	if got := acquired.Load(); got != 2 {
		t.Fatalf("late invalidation created acquisition %d", got)
	}
	select {
	case id := <-released:
		t.Fatalf("acquisition %d released again", id)
	default:
	}
}

// BenchmarkAccessResolverLifecycle measures a real controller's acquisition,
// publication, removal and joined resolver cleanup with a no-op getter backend.
func BenchmarkAccessResolverLifecycle(b *testing.B) {
	// Keep one controller and join every resolver before the next acquisition.
	ctx := b.Context()
	ctrl := controller.NewController(ctx, logrus.NewEntry(logrus.New()))
	published := make(chan struct{}, 1)
	done := make(chan error, 1)
	access := directive.NewAccessResolver(func(context.Context, func()) (string, func(), error) {
		return "value", func() {}, nil
	})
	remove, err := ctrl.AddHandler(directive.NewFuncHandler(func(context.Context, directive.Instance) ([]directive.Resolver, error) {
		return directive.R(directive.NewFuncResolver(func(ctx context.Context, handler directive.ResolverHandler) error {
			err := access.Resolve(ctx, handler)
			done <- err
			return err
		}), nil)
	}))
	if err != nil {
		b.Fatal(err)
	}
	defer remove()
	handler := directive.NewCallbackHandler(func(directive.AttachedValue) { published <- struct{}{} }, nil, nil)
	b.ReportAllocs()

	// Include the complete real-controller cycle, excluding external backend work.
	for b.Loop() {
		_, ref, err := ctrl.AddDirective(ctx, &directive_mock.MockDirective{}, handler)
		if err != nil {
			b.Fatal(err)
		}
		<-published
		ref.Release()
		if err := <-done; !errors.Is(err, context.Canceled) {
			b.Fatal(err)
		}
	}
}
