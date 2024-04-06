package directive

import (
	"context"
)

// TransformAttachedValueFunc transforms an AttachedValue.
// rval is the returned transformed value to emit to the handler (if any).
// rel is the release function to call when the value is no longer valid (if any).
// rval is ignored if ok = false.
// err, if any, will unwind the resolver with the error.
type TransformAttachedValueFunc[T Value] func(ctx context.Context, val AttachedValue) (rval T, rel func(), ok bool, err error)

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
	type addedVal struct {
		id  uint32
		rel func()
	}
	addedVals := make(map[uint32]addedVal)
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
			var valRel func()
			if xfrm := r.xfrm; xfrm != nil {
				result, rel, ok, err := xfrm(ctx, av)
				if err != nil {
					pushErr(err)
					if rel != nil {
						rel()
					}
					return
				}
				if !ok {
					if rel != nil {
						rel()
					}
					return
				}
				val = result
				valRel = rel
			} else {
				var ok bool
				val, ok = av.GetValue().(T)
				if !ok {
					return
				}
			}

			id, accepted := handler.AddValue(val)
			if accepted {
				addedVals[id] = addedVal{id: av.GetValueID(), rel: valRel}
			} else {
				if valRel != nil {
					valRel()
				}
			}
		}, func(av AttachedValue) {
			valueID := av.GetValueID()
			if addedValue, ok := addedVals[valueID]; ok {
				delete(addedVals, valueID)
				_, _ = handler.RemoveValue(addedValue.id)
				if addedValue.rel != nil {
					addedValue.rel()
				}
			}
		},
			func() {
				for k, addedValue := range addedVals {
					delete(addedVals, k)
					_, _ = handler.RemoveValue(addedValue.id)
					if addedValue.rel != nil {
						addedValue.rel()
					}
				}
			},
		),
	)
	if err != nil {
		return err
	}
	defer handler.ClearValues()
	defer ref.Release()

	defer di.AddIdleCallback(func(isIdle bool, errs []error) {
		for _, err := range errs {
			if err != nil {
				pushErr(err)
				return
			}
		}

		handler.MarkIdle(isIdle)
	})()

	select {
	case <-ctx.Done():
		return context.Canceled
	case err := <-errCh:
		return err
	}
}

// _ is a type assertion
var _ Resolver = ((*TransformResolver[string])(nil))
