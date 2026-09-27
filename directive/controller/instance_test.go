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

	di, strongRef, err := ctrl.AddDirective(&directive_mock.MockDirective{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	<-readyCh

	weakHandler := &countingRefHandler{}
	weakRef := di.AddReference(weakHandler, true)
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
	di, ref, err := ctrl.AddDirective(&directive_mock.MockDirective{}, nil)
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
