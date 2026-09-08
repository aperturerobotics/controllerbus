package inmem

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"
)

// TestCloseCancelsBeforeJoining checks dependency cancellation and cleanup errors.
func TestCloseCancelsBeforeJoining(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		b, _ := newTrackingBus()
		firstClosing := make(chan struct{})
		finishFirst := make(chan struct{})
		var secondCanceled atomic.Bool
		closeFailure := errors.New("close failed")
		first := &lifecycleController{
			executeFn: func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			},
			closeFn: func() error {
				close(firstClosing)
				<-finishFirst
				return closeFailure
			},
		}
		second := &lifecycleController{executeFn: func(ctx context.Context) error {
			<-ctx.Done()
			secondCanceled.Store(true)
			return nil
		}}
		for _, controller := range []*lifecycleController{first, second} {
			if _, err := b.AddController(t.Context(), controller, nil); err != nil {
				t.Fatal(err)
			}
		}
		closed := make(chan error, 1)
		go func() { closed <- b.Close() }()
		<-firstClosing
		synctest.Wait()
		if !secondCanceled.Load() {
			t.Fatal("shutdown joined the first controller before canceling its dependency")
		}
		close(finishFirst)
		if err := <-closed; !errors.Is(err, closeFailure) {
			t.Fatalf("shutdown lost the cleanup error: %v", err)
		}
		if err := b.Close(); !errors.Is(err, closeFailure) {
			t.Fatalf("repeated shutdown changed its result: %v", err)
		}
		if first.closeCalls.Load() != 1 || second.closeCalls.Load() != 1 || len(b.GetControllers()) != 0 {
			t.Fatal("shutdown did not close each controller exactly once")
		}
		rejected := &lifecycleController{}
		if _, err := b.AddController(t.Context(), rejected, nil); !errors.Is(err, bus.ErrClosed) || rejected.closeCalls.Load() != 1 {
			t.Fatalf("closed bus admitted or leaked a controller: %v", err)
		}
	})
}

// TestCloseJoinsDetachedController checks a release that is already closing storage.
func TestCloseJoinsDetachedController(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		b, tracking := newTrackingBus()
		closing := make(chan struct{})
		finish := make(chan struct{})
		controller := &lifecycleController{closeFn: func() error {
			close(closing)
			<-finish
			return nil
		}}
		release, err := b.AddController(t.Context(), controller, nil)
		if err != nil {
			t.Fatal(err)
		}
		released := make(chan struct{})
		go func() { release(); close(released) }()
		<-tracking.detached
		<-closing
		if len(b.GetControllers()) != 0 {
			t.Fatal("closing controller remains publicly attached")
		}
		var complete atomic.Bool
		closed := make(chan error, 1)
		go func() {
			closed <- b.Close()
			complete.Store(true)
		}()
		synctest.Wait()
		if complete.Load() {
			t.Fatal("shutdown returned while detached controller cleanup was running")
		}
		close(finish)
		<-released
		if err := <-closed; err != nil {
			t.Fatal(err)
		}
	})
}

// TestCloseJoinsAdmission checks handler registration racing with shutdown.
func TestCloseJoinsAdmission(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		b, tracking := newTrackingBus()
		entered := make(chan struct{})
		proceed := make(chan struct{})
		tracking.beforeAdd = func(directive.Handler) {
			close(entered)
			<-proceed
		}
		controller := &lifecycleController{}
		admitted := make(chan error, 1)
		go func() {
			_, err := b.AddController(t.Context(), controller, nil)
			admitted <- err
		}()
		<-entered
		closed := make(chan error, 1)
		go func() { closed <- b.Close() }()
		synctest.Wait()
		select {
		case <-closed:
			t.Fatal("shutdown returned before an admitted handler finished registration")
		default:
		}
		close(proceed)
		if err := <-admitted; err != nil {
			t.Fatal(err)
		}
		if err := <-closed; err != nil {
			t.Fatal(err)
		}
		if controller.closeCalls.Load() != 1 {
			t.Fatal("shutdown lost the concurrently admitted controller")
		}
	})
}
