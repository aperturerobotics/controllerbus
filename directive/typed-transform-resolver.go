package directive

import (
	"context"
	"errors"
)

// TransformTypedAttachedValueFunc transforms a TypedAttachedValue.
// rval is the returned transformed value to emit to the handler (if any).
// rel is the release function to call when the value is no longer valid (if any).
// rval is ignored if ok = false.
// err, if any, will unwind the resolver with the error.
type TransformTypedAttachedValueFunc[V ComparableValue, T Value] func(ctx context.Context, val TypedAttachedValue[V]) (rval T, rel func(), ok bool, err error)

// TypedTransformResolver resolves a directive by adding another directive and transforming the results.
// V is the type of the input TypedAttachedValue.
// T is the type of the value that we will yield to the handler.
// xfrm returns the transformed value to emit to the handler, bool true/false to emit the value, and any error.
// if xfrm returns an error the entire resolver unwinds with the error.
// NOTE: xfrm must not block as it will block the TypedCallbackHandler.
type TypedTransformResolver[V ComparableValue, T Value] struct {
	adder              DirectiveAdder
	dir                Directive
	xfrm               TransformTypedAttachedValueFunc[V, T]
	unknownTypeHandler ReferenceHandler
}

// NewTypedTransformResolver constructs a new TypedTransformResolver.
func NewTypedTransformResolver[V ComparableValue, T Value](
	adder DirectiveAdder,
	dir Directive,
	xfrm TransformTypedAttachedValueFunc[V, T],
	unknownTypeHandler ReferenceHandler,
) *TypedTransformResolver[V, T] {
	return &TypedTransformResolver[V, T]{
		adder:              adder,
		dir:                dir,
		xfrm:               xfrm,
		unknownTypeHandler: unknownTypeHandler,
	}
}

// Resolve resolves the values, emitting them to the handler.
func (r *TypedTransformResolver[V, T]) Resolve(ctx context.Context, handler ResolverHandler) error {
	if r.xfrm == nil {
		return errors.New("xfrm cannot be nil for TypedTransformResolver")
	}

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
		NewTypedCallbackHandler(
			func(tav TypedAttachedValue[V]) {
				var val T
				result, rel, ok, err := r.xfrm(ctx, tav)
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

				id, accepted := handler.AddValue(val)
				if accepted {
					addedVals[id] = addedVal{id: tav.GetValueID(), rel: rel}
				} else {
					if rel != nil {
						rel()
					}
				}
			},
			func(tav TypedAttachedValue[V]) {
				valueID := tav.GetValueID()
				if addedValue, ok := addedVals[valueID]; ok {
					delete(addedVals, valueID)
					_, _ = handler.RemoveValue(addedValue.id)
					if addedValue.rel != nil {
						addedValue.rel()
					}
				}
			},
			func() {
				pushErr(ErrDirectiveDisposed)
				for id, val := range addedVals {
					if val.rel != nil {
						val.rel()
					}
					delete(addedVals, id)
				}
			},
			r.unknownTypeHandler,
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
var _ Resolver = ((*TypedTransformResolver[int, string])(nil))
