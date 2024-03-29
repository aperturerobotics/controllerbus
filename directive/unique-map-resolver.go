package directive

import (
	"github.com/aperturerobotics/util/unique"
)

// UniqueMapResolver accepts and de-duplicates a list of keyed objects.
// The objects are emitted to the ValueHandler.
// If an object changes, the old value is removed & replaced.
//
// cmp checks if two values are equal.
// if equal, the old version of the value is used.
//
// K is the key type
// V is the value type
type UniqueMapResolver[K, V comparable] struct {
	*unique.KeyedMap[K, V]
}

// NewUniqueMapResolver constructs a new UniqueMapResolver.
func NewUniqueMapResolver[K, V comparable](
	cmp func(k K, a V, b V) bool,
	hnd ValueHandler,
) *UniqueMapResolver[K, V] {
	valueIds := make(map[K]uint32)
	return &UniqueMapResolver[K, V]{
		KeyedMap: unique.NewKeyedMap[K, V](
			cmp,
			func(k K, v V, added bool, removed bool) {
				if added {
					valID, accepted := hnd.AddValue(v)
					if accepted {
						valueIds[k] = valID
					}
					return
				}

				// removed or changed
				valID, ok := valueIds[k]
				if ok {
					_, _ = hnd.RemoveValue(valID)
					delete(valueIds, k)
				}
				if removed {
					return
				}

				// changed
				valID, accepted := hnd.AddValue(v)
				if accepted {
					valueIds[k] = valID
				}
			},
			// set initial to nil so we generate the AddValue calls.
			nil,
		),
	}
}

// UniqueMapXfrmResolver accepts and de-duplicates a list of keyed objects.
// The objects are transformed and emitted to the ValueHandler.
// If an object changes, the old value is removed & replaced.
//
// cmp checks if two values are equal.
// if equal, the old version of the value is used.
//
// xfrm transforms the values before emitting them.
// if xfrm returns false the value is dropped.
//
// K is the key type
// V is the value type
// O is the type to emit to the resolver
type UniqueMapXfrmResolver[K, V comparable, O any] struct {
	*unique.KeyedMap[K, V]
}

// NewUniqueMapXfrmResolver constructs a new UniqueMapXfrmResolver.
func NewUniqueMapXfrmResolver[K, V comparable, O any](
	cmp func(k K, a V, b V) bool,
	xfrm func(k K, v V) (O, bool),
	hnd ValueHandler,
) *UniqueMapXfrmResolver[K, V, O] {
	valueIds := make(map[K]uint32)
	return &UniqueMapXfrmResolver[K, V, O]{
		KeyedMap: unique.NewKeyedMap[K, V](
			cmp,
			func(k K, v V, added bool, removed bool) {
				if added {
					xfrm, xfrmOk := xfrm(k, v)
					if xfrmOk {
						valID, accepted := hnd.AddValue(xfrm)
						if accepted {
							valueIds[k] = valID
						}
					}
					return
				}

				// removed or changed
				valID, ok := valueIds[k]
				if ok {
					_, _ = hnd.RemoveValue(valID)
					delete(valueIds, k)
				}
				if removed {
					return
				}

				// changed
				xfrm, xfrmOk := xfrm(k, v)
				if xfrmOk {
					valID, accepted := hnd.AddValue(xfrm)
					if accepted {
						valueIds[k] = valID
					}
				}
			},
			// set initial to nil so we generate the AddValue calls.
			nil,
		),
	}
}
