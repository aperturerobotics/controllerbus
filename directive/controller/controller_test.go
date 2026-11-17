package controller

import (
	"context"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/directive"
	directive_mock "github.com/aperturerobotics/controllerbus/directive/mock"
)

// TestRemovedHandlerRejectsPendingResolvers pauses a handler until removal
// finishes, then requires its late result to remain outside the directive.
func TestRemovedHandlerRejectsPendingResolvers(t *testing.T) {
	// Hold the handler call while another goroutine removes its registration.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	ctrl := NewController(ctx, nil)
	entered := make(chan struct{})
	resume := make(chan struct{})
	remove, err := ctrl.AddHandler(directive.NewFuncHandler(func(context.Context, directive.Instance) ([]directive.Resolver, error) {
		close(entered)
		select {
		case <-resume:
			return directive.R(directive.NewValueResolver([]string{"removed host"}), nil)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(remove)

	// Begin a directive and remove the handler before its resolver returns.
	done := make(chan *directiveInstance, 1)
	go func() {
		di, ref, err := ctrl.AddDirective(ctx, &directive_mock.MockDirective{}, nil)
		if err != nil {
			t.Error(err)
			done <- nil
			return
		}
		t.Cleanup(ref.Release)
		done <- di.(*directiveInstance)
	}()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("handler did not enter")
	}
	remove()
	close(resume)

	// Inspect attachment synchronously, without waiting for a stale value.
	select {
	case di := <-done:
		if di == nil {
			t.Fatal("directive creation failed")
		}
		ctrl.mtx.Lock()
		defer ctrl.mtx.Unlock()
		if len(di.res) != 0 {
			t.Fatal("removed handler attached its late resolver")
		}
	case <-ctx.Done():
		t.Fatal("directive did not return")
	}
}

// TestRemovedDirectiveRejectsPendingResolvers covers the AddHandler path when
// its unlocked handler call outlives the directive it was called for.
func TestRemovedDirectiveRejectsPendingResolvers(t *testing.T) {
	// Create a directive before attaching the paused handler.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	ctrl := NewController(ctx, nil)
	instance, ref, err := ctrl.AddDirective(ctx, &directive_mock.MockDirective{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(ref.Release)
	entered := make(chan struct{})
	resume := make(chan struct{})
	done := make(chan func(), 1)

	// Close the directive while AddHandler waits for HandleDirective to return.
	go func() {
		remove, err := ctrl.AddHandler(directive.NewFuncHandler(func(context.Context, directive.Instance) ([]directive.Resolver, error) {
			close(entered)
			select {
			case <-resume:
				return directive.R(directive.NewValueResolver([]string{"removed host"}), nil)
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}))
		if err != nil {
			t.Error(err)
		}
		done <- remove
	}()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("handler did not enter")
	}
	instance.Close()
	close(resume)

	// A closed directive cannot acquire resolvers from the late handler result.
	select {
	case remove := <-done:
		if remove == nil {
			t.Fatal("handler registration failed")
		}
		t.Cleanup(remove)
		ctrl.mtx.Lock()
		defer ctrl.mtx.Unlock()
		if len(instance.(*directiveInstance).res) != 0 {
			t.Fatal("removed directive attached a late resolver")
		}
	case <-ctx.Done():
		t.Fatal("handler registration did not return")
	}
}

// TestIdleCallbackRemovalCancelsStartingResolver requires a reentrant removal
// to cancel the context that startup publishes before idle notification.
func TestIdleCallbackRemovalCancelsStartingResolver(t *testing.T) {
	// Make an empty directive idle before attaching a resolver.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	ctrl := NewController(ctx, nil)
	instance, ref, err := ctrl.AddDirective(ctx, &directive_mock.MockDirective{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(ref.Release)
	defer instance.AddIdleCallback(func(idle bool, _ []error) {
		if !idle {
			ctrl.mtx.Lock()
			ctrl.removeHandlerLocked(ctrl.hnd[0])
			ctrl.mtx.Unlock()
		}
	})()

	// Observe the resolver's own cancellation rather than its parent's deadline.
	exited := make(chan error, 1)
	remove, err := ctrl.AddHandler(directive.NewFuncHandler(func(context.Context, directive.Instance) ([]directive.Resolver, error) {
		return directive.R(directive.NewFuncResolver(func(ctx context.Context, _ directive.ResolverHandler) error {
			<-ctx.Done()
			exited <- ctx.Err()
			return ctx.Err()
		}), nil)
	}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(remove)

	// The callback must leave no running context behind on the removed resolver.
	select {
	case <-exited:
		if ctx.Err() != nil {
			t.Fatal("resolver only stopped at the parent deadline")
		}
	case <-ctx.Done():
		t.Fatal("idle callback removal did not cancel the starting resolver")
	}
}
