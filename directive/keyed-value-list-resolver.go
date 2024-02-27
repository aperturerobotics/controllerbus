package directive

import (
	"github.com/aperturerobotics/util/keyedlist"
)

// KeyedListResolver accepts and de-duplicates a list of keyed objects.
// The objects are emitted to the ValueHandler.
// If an object changes, the old value is removed & replaced.
//
// cmp checks if two values are equal.
// if equal, the old version of the value is used.
//
// K is the key type
// V is the value type
type KeyedListResolver[K, V comparable] struct {
	*keyedlist.KeyedList[K, V]
}

// NewKeyedListResolver constructs a new KeyedListResolver.
func NewKeyedListResolver[K, V comparable](
	getKey func(v V) K,
	cmp func(a V, b V) bool,
	hnd ValueHandler,
) *KeyedListResolver[K, V] {
	valueIds := make(map[K]uint32)
	return &KeyedListResolver[K, V]{
		KeyedList: keyedlist.NewKeyedList[K, V](
			getKey,
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
