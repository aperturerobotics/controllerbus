package controller_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

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
