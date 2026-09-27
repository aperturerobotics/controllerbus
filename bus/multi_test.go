package bus_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/bus/inmem"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/controllerbus/directive/controller"
	directive_mock "github.com/aperturerobotics/controllerbus/directive/mock"
	"github.com/sirupsen/logrus"
)

// TestCollectValuesWatchRetainsHealthyValues keeps live values visible while
// another resolver fails and after that failed resolver is removed.
func TestCollectValuesWatchRetainsHealthyValues(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	ctrl := controller.NewController(ctx, logrus.NewEntry(logrus.New()))
	b := inmem.NewBus(ctrl)
	t.Cleanup(func() { b.Close() })
	removeHealthy, err := ctrl.AddHandler(directive.NewFuncHandler(func(context.Context, directive.Instance) ([]directive.Resolver, error) {
		return []directive.Resolver{directive.NewFuncResolver(func(_ context.Context, handler directive.ResolverHandler) error {
			handler.AddValue("healthy")
			return nil
		})}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer removeHealthy()
	type snapshot struct {
		failed bool
		values []string
	}
	states := make(chan snapshot, 16)
	_, release, err := bus.ExecCollectValuesWatch(ctx, b, &directive_mock.MockDirective{}, true,
		func(errs []error, values []string) error {
			select {
			case states <- snapshot{len(errs) != 0, values}:
			case <-ctx.Done():
			}
			return nil
		}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	await := func(failed bool) {
		t.Helper()
		for {
			select {
			case state := <-states:
				if state.failed != failed {
					continue
				}
				if !slices.Equal(state.values, []string{"healthy"}) {
					t.Fatalf("failed=%v: healthy values disappeared: %v", failed, state.values)
				}
				return
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
		}
	}
	await(false)
	removeFailed, err := ctrl.AddHandler(directive.NewFuncHandler(func(context.Context, directive.Instance) ([]directive.Resolver, error) {
		return []directive.Resolver{directive.NewFuncResolver(func(context.Context, directive.ResolverHandler) error {
			return errors.New("unrelated resolver failed")
		})}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer removeFailed()
	await(true)
	removeFailed()
	await(false)
}
