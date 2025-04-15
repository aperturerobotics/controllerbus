package directive

import (
	"context"
	"sync/atomic"

	"github.com/aperturerobotics/util/broadcast"
)

// KeyedGetterFunc is a keyed getter function to resolve a keyed resolver.
//
// release is a function to call if the value is released.
// return nil, nil, nil for not found.
// return a release function if necessary
type KeyedGetterFunc[K, V comparable] func(ctx context.Context, key K, release func()) (V, func(), error)

// KeyedGetterResolver resolves a directive with a keyed getter function.
type KeyedGetterResolver[K, V comparable] struct {
	getter KeyedGetterFunc[K, V]
	key    K
}

// NewKeyedGetterResolver constructs a new KeyedGetterResolver.
func NewKeyedGetterResolver[K, V comparable](getter KeyedGetterFunc[K, V], key K) *KeyedGetterResolver[K, V] {
	return &KeyedGetterResolver[K, V]{getter: getter, key: key}
}

// Resolve resolves the values, emitting them to the handler.
func (r *KeyedGetterResolver[K, V]) Resolve(ctx context.Context, handler ResolverHandler) error {
	var bcast broadcast.Broadcast
	for {
		_ = handler.ClearValues()

		var waitCh <-chan struct{}
		bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
			waitCh = getWaitCh()
		})

		var valID atomic.Uint32
		val, relVal, err := r.getter(ctx, r.key, func() {
			old := valID.Swap(0)
			if old != 0 {
				_, _ = handler.RemoveValue(old)
				bcast.HoldLock(func(broadcast func(), getWaitCh func() <-chan struct{}) {
					broadcast()
				})
			}
		})
		if err != nil {
			if relVal != nil {
				relVal()
			}
			return err
		}
		var empty V
		if val == empty {
			// no values
			return nil
		}

		var result V = val //nolint:staticcheck
		addedValID, accepted := handler.AddValue(result)
		if !accepted {
			// value not needed
			valID.Store(0)
			if relVal != nil {
				relVal()
			}
			return nil
		}
		if !valID.CompareAndSwap(0, addedValID) {
			// already released
			valID.Store(0)
			if relVal != nil {
				relVal()
			}
			_, _ = handler.RemoveValue(addedValID)

			// try again
			continue
		}

		// wait for the broadcast or context to be canceled
		handler.MarkIdle(true)
		select {
		case <-ctx.Done():
		case <-waitCh:
		}

		valID.Store(0)
		if relVal != nil {
			relVal()
		}
		if ctx.Err() != nil {
			return context.Canceled
		}
	}
}

// _ is a type assertion
var _ Resolver = ((*KeyedGetterResolver[string, int])(nil))
