package bus

import (
	"context"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/ccontainer"
	"github.com/aperturerobotics/util/refcount"
)

// NewOneOffRefCountResolver builds a RefCount resolver which executes a one-off directive.
//
// When at least one reference to the container is present, and ctx is set to an
// active context, the directive will be added to the bus, and the first valid
// value returned matching the type parameter and optional filter will be stored
// in the refcount container.
//
// idleCb is called when idle with the list of resolver errors.
// idleCb should return (wait, error): if wait=true, continues to wait.
// if idleCb is nil: continues to wait when the directive becomes idle
// errs is the list of errors from the resolvers (if any)
//
// If err != nil, ref == nil.
func NewOneOffRefCountResolver[T comparable](
	bus Bus,
	dir directive.Directive,
	idleCb ExecIdleCallback,
	filterCb func(val directive.TypedAttachedValue[T]) (bool, error),
) refcount.RefCountResolver[T] {
	return func(ctx context.Context, released func()) (T, func(), error) {
		av, _, ref, err := ExecOneOffWithFilterTyped[T](
			ctx,
			bus,
			dir,
			idleCb,
			released,
			filterCb,
		)
		if err != nil || ref == nil {
			var empty T
			return empty, nil, err
		}

		return av.GetValue(), ref.Release, nil
	}
}

// NewOneOffRefCount builds a RefCount which executes a one-off directive.
//
// When at least one reference to the container is present, and ctx is set to an
// active context, the directive will be added to the bus, and the first valid
// value returned matching the type parameter and optional filter will be stored
// in the refcount container.
//
// ctx, target and targetErr can be empty
// keepUnref sets if the value should be kept if there are zero references.
//
// idleCb is called when idle with the list of resolver errors.
// idleCb should return (wait, error): if wait=true, continues to wait.
// if idleCb is nil: continues to wait when the directive becomes idle
// errs is the list of errors from the resolvers (if any)
//
// If err != nil, ref == nil.
func NewOneOffRefCount[T comparable](
	ctx context.Context,
	bus Bus,
	dir directive.Directive,
	idleCb ExecIdleCallback,
	filterCb func(val directive.TypedAttachedValue[T]) (bool, error),
	keepUnref bool,
	target *ccontainer.CContainer[T],
	targetErr *ccontainer.CContainer[*error],
) *refcount.RefCount[T] {
	return refcount.NewRefCount[T](
		ctx,
		keepUnref,
		target,
		targetErr,
		NewOneOffRefCountResolver[T](bus, dir, idleCb, filterCb),
	)
}
