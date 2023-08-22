package directive

import (
	"context"

	"github.com/aperturerobotics/util/keyed"
)

// KeyedRefCountResolver resolves a directive with a Keyed RefCount container.
//
// Adds a reference, waits for a value, and adds it to the handler.
// Removes the value and waits for a new one when the value is released.
// Calls MarkIdle when the value has been added.
// buildValue can be nil
// if useCtx is set, uses the refcount resolver context to start the refcount container.
// If buildValue is set, will be called with the values.
// if buildValue returns nil, nil, ignores the value.
type KeyedRefCountResolver[K, V comparable] struct {
	rc         *keyed.KeyedRefCount[K, V]
	key        K
	useCtx     bool
	buildValue func(ctx context.Context, val V) (Value, error)
}

// NewKeyedRefCountResolver constructs a new KeyedRefCountResolver.
//
// if useCtx is set, uses the refcount resolver context to start the refcount container.
func NewKeyedRefCountResolver[K, V comparable](
	rc *keyed.KeyedRefCount[K, V],
	key K,
	useCtx bool,
	buildValue func(ctx context.Context, val V) (Value, error),
) *KeyedRefCountResolver[K, V] {
	return &KeyedRefCountResolver[K, V]{
		rc:         rc,
		key:        key,
		useCtx:     useCtx,
		buildValue: buildValue,
	}
}

// Resolve resolves the values, emitting them to the handler.
func (r *KeyedRefCountResolver[K, V]) Resolve(ctx context.Context, handler ResolverHandler) error {
	if r.useCtx {
		_ = r.rc.SetContextIfCanceled(ctx, false)
	}

	ref, data, _ := r.rc.AddKeyRef(r.key)
	defer ref.Release()

	var val Value
	var empty V
	if r.buildValue != nil {
		var err error
		val, err = r.buildValue(ctx, data)
		if err != nil {
			return err
		}
	} else if data != empty {
		val = data
	}

	if val == nil {
		return nil
	}

	_, accepted := handler.AddValue(val)
	if !accepted {
		return nil
	}

	// mark idle & wait for ctx to be canceled
	handler.MarkIdle()
	<-ctx.Done()
	handler.ClearValues()
	return context.Canceled
}

// _ is a type assertion
var _ Resolver = ((*KeyedRefCountResolver[string, int])(nil))
