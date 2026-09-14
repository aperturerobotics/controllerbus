package bus_bridge_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/bus"
	bus_bridge "github.com/aperturerobotics/controllerbus/bus/bridge"
	"github.com/aperturerobotics/controllerbus/bus/inmem"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/callback"
	"github.com/aperturerobotics/controllerbus/directive"
	cdc "github.com/aperturerobotics/controllerbus/directive/controller"
	boilerplate "github.com/aperturerobotics/controllerbus/example/boilerplate/v1"
	"github.com/sirupsen/logrus"
)

// TestBusBridgeOptionalLookup preserves empty and failed results across the bridge.
func TestBusBridgeOptionalLookup(t *testing.T) {
	for _, fail := range []bool{false, true} {
		name := "empty"
		if fail {
			name = "failed"
		}
		t.Run(name, func(t *testing.T) {
			// Create separate buses so optional lookups must cross the real bridge.
			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			t.Cleanup(cancel)
			le := logrus.NewEntry(logrus.New())
			parent := inmem.NewBus(cdc.NewController(ctx, le))
			child := inmem.NewBus(cdc.NewController(ctx, le))
			t.Cleanup(func() { _ = parent.Close() })
			t.Cleanup(func() { _ = child.Close() })
			release, err := child.AddController(ctx, bus_bridge.NewBusBridge(parent, nil), nil)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(release)

			// A matching resolver failure must reach the caller instead of hanging.
			refused := errors.New("provider refused lookup")
			if fail {
				release, err := parent.AddController(ctx, callback.NewCallbackController(
					controller.NewInfo("test", controller.MustParseVersion("0.0.1"), "test"), nil,
					func(context.Context, directive.Instance) ([]directive.Resolver, error) {
						return directive.Resolvers(directive.NewFuncResolver(func(context.Context, directive.ResolverHandler) error {
							return refused
						})), nil
					}, nil,
				), nil)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(release)
			}

			// An idle empty lookup completes without a value; errors retain their identity.
			value, _, ref, err := bus.ExecOneOff(ctx, child, &boilerplate.Boilerplate{MessageText: "optional"}, bus.ReturnIfIdle(true), nil)
			if ref != nil {
				ref.Release()
			}
			if fail && !errors.Is(err, refused) {
				t.Fatalf("lookup error = %v, want provider refusal", err)
			}
			if !fail && err != nil {
				t.Fatalf("empty lookup did not settle: %v", err)
			}
			if value != nil {
				t.Fatal("empty lookup produced a value")
			}
		})
	}
}
