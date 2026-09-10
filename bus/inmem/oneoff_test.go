package inmem

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"
	directive_controller "github.com/aperturerobotics/controllerbus/directive/controller"
	boilerplate_v1 "github.com/aperturerobotics/controllerbus/example/boilerplate/v1"
)

// TestOneOffTransformCancellation allows cancellation to win while the accepted
// value's transformation is still returning. Both APIs must return an empty
// result on cancellation without accessing the callback's unpublished value.
func TestOneOffTransformCancellation(t *testing.T) {
	for _, typed := range []bool{false, true} {
		for range 32 {
			ctx, cancel := context.WithCancel(t.Context())
			b := NewBus(directive_controller.NewController(t.Context(), nil))
			ctrl := &lifecycleController{handleFn: func(context.Context, directive.Instance) ([]directive.Resolver, error) {
				return directive.Resolvers(directive.NewValueResolver([]int{1})), nil
			}}
			release, err := b.AddController(t.Context(), ctrl, nil)
			if err != nil {
				t.Fatal(err)
			}
			transform := func() {
				cancel()
				runtime.Gosched()
			}
			var value any
			var ref directive.Reference
			if typed {
				var result *int
				result, _, _, ref, err = bus.ExecOneOffWithXfrmTyped(ctx, b, &boilerplate_v1.Boilerplate{}, nil, nil, func(directive.TypedAttachedValue[int]) (*int, bool, error) {
					transform()
					return new(2), true, nil
				})
				if result != nil {
					value = *result
				}
			} else {
				value, _, _, ref, err = bus.ExecOneOffWithXfrm(ctx, b, &boilerplate_v1.Boilerplate{}, nil, nil, func(directive.AttachedValue) (directive.Value, bool, error) {
					transform()
					return 2, true, nil
				})
			}
			if ref != nil {
				ref.Release()
			}
			release()
			cancel()
			if errors.Is(err, context.Canceled) {
				if value != nil {
					t.Fatalf("cancellation returned an unpublished result: %v", value)
				}
			} else if err != nil || value != 2 {
				t.Fatalf("successful transform returned value=%v err=%v", value, err)
			}
		}
	}
}
