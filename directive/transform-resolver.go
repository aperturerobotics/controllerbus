package directive

import (
	"context"
)

// TransformAttachedValueFunc transforms an AttachedValue.
// Returns the transformed value to emit to the handler, bool true/false to emit the value, and any error.
type TransformAttachedValueFunc[T Value] func(ctx context.Context, val AttachedValue) (T, bool, error)

// TransformResolver resolves a directive by adding another directive and transforming the results.
// T is the type of the value that we will yield to the handler.
// xfrm returns the transformed value to emit to the handler, bool true/false to emit the value, and any error.
// if xfrm returns an error the entire resolver unwinds with the error.
// NOTE: xfrm must not block as it will block the CallbackHandler.
type TransformResolver[T Value] struct {
	adder DirectiveAdder
	dir   Directive
	xfrm  TransformAttachedValueFunc[T]
}

// NewTransformResolver constructs a new TransformResolver.
func NewTransformResolver[T Value](adder DirectiveAdder, dir Directive, xfrm TransformAttachedValueFunc[T]) *TransformResolver[T] {
	return &TransformResolver[T]{
		adder: adder,
		dir:   dir,
		xfrm:  xfrm,
	}
}

// Resolve resolves the values, emitting them to the handler.
func (r *TransformResolver[T]) Resolve(ctx context.Context, handler ResolverHandler) error {
	addedVals := make(map[uint32]uint32)
	errCh := make(chan error, 1)
	pushErr := func(err error) {
		select {
		case errCh <- err:
		default:
		}
	}
	di, ref, err := r.adder.AddDirective(
		r.dir,
		NewCallbackHandler(func(av AttachedValue) {
			var val T
			if xfrm := r.xfrm; xfrm != nil {
				result, ok, err := xfrm(ctx, av)
				if err != nil {
					pushErr(err)
					return
				}
				if !ok {
					return
				}
				val = result
			} else {
				var ok bool
				val, ok = av.GetValue().(T)
				if !ok {
					return
				}
			}

			id, accepted := handler.AddValue(val)
			if accepted {
				addedVals[id] = av.GetValueID()
			}
		}, func(av AttachedValue) {
			valueID := av.GetValueID()
			if addedValueID, ok := addedVals[valueID]; ok {
				delete(addedVals, valueID)
				_, _ = handler.RemoveValue(addedValueID)
			}
		},
			nil,
		),
	)
	if err != nil {
		return err
	}
	defer handler.ClearValues()
	defer ref.Release()

	defer di.AddIdleCallback(func(errs []error) {
		for _, err := range errs {
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
		}

		errCh <- nil
	})()

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case err := <-errCh:
			if err != nil {
				return err
			}
			handler.MarkIdle()
		}
	}
}

// _ is a type assertion
var _ Resolver = ((*TransformResolver[string])(nil))
