package keyed

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

// KeyedRefCount manages a list of running routines with reference counts.
type KeyedRefCount[T comparable] struct {
	// keyed is the underlying keyed controller
	keyed *Keyed[T]

	// mtx guards below fields
	mtx sync.Mutex
	// refs is the list of keyed refs.
	refs map[string][]*KeyedRef[T]
}

// KeyedRef is a reference to a key.
type KeyedRef[T comparable] struct {
	rc  *KeyedRefCount[T]
	key string
	rel atomic.Bool
}

// Release releases the reference.
func (k *KeyedRef[T]) Release() {
	if k.rel.Swap(true) {
		return
	}
	k.rc.mtx.Lock()
	refs := k.rc.refs[k.key]
	for i := 0; i < len(refs); i++ {
		if refs[i] == k {
			refs[i] = refs[len(refs)-1]
			refs[len(refs)-1] = nil
			refs = refs[:len(refs)-1]

			if len(refs) == 0 {
				delete(k.rc.refs, k.key)
				_ = k.rc.keyed.RemoveKey(k.key)
			} else {
				k.rc.refs[k.key] = refs
			}
			break
		}
	}
	k.rc.mtx.Unlock()
}

// NewKeyedRefCount constructs a new Keyed execution manager with reference counting.
// Note: routines won't start until SetContext is called.
func NewKeyedRefCount[T comparable](
	ctorCb func(key string) (Routine, T),
	opts ...Option[T],
) *KeyedRefCount[T] {
	return &KeyedRefCount[T]{
		keyed: NewKeyed(ctorCb, opts...),
		refs:  make(map[string][]*KeyedRef[T]),
	}
}

// NewKeyedRefCountWithLogger constructs a new Keyed execution manager with reference counting.
// Logs when a controller exits without being removed from the Keys set.
// Note: routines won't start until SetContext is called.
func NewKeyedRefCountWithLogger[T comparable](
	ctorCb func(key string) (Routine, T),
	le *logrus.Entry,
) *KeyedRefCount[T] {
	return &KeyedRefCount[T]{
		keyed: NewKeyedWithLogger(ctorCb, le),
		refs:  make(map[string][]*KeyedRef[T]),
	}
}

// SetContext updates the root context, restarting all running routines.
// if restart is true, all errored routines also restart
func (k *KeyedRefCount[T]) SetContext(ctx context.Context, restart bool) {
	k.keyed.SetContext(ctx, restart)
}

// GetKeys returns the list of keys registered with the Keyed instance.
// Note: this is an instantaneous snapshot.
func (k *KeyedRefCount[T]) GetKeys() []string {
	return k.keyed.GetKeys()
}

// GetKeysWithData returns the keys and the data for the keys.
// Note: this is an instantaneous snapshot.
func (k *KeyedRefCount[T]) GetKeysWithData() []KeyWithData[T] {
	return k.keyed.GetKeysWithData()
}

// GetKey returns the routine for the given key.
// Note: this is an instantaneous snapshot.
func (k *KeyedRefCount[T]) GetKey(key string) (Routine, T) {
	return k.keyed.GetKey(key)
}

// ResetRoutine resets the given routine after checking the condition functions.
// If any return true, resets the instance.
//
// If len(conds) == 0, always resets the given key.
func (k *KeyedRefCount[T]) ResetRoutine(key string, conds ...func(T) bool) (existed bool, reset bool) {
	return k.keyed.ResetRoutine(key, conds...)
}

// AddKeyRef adds a reference to the given key.
// Returns if the key already existed or not.
func (k *KeyedRefCount[T]) AddKeyRef(key string) (ref *KeyedRef[T], existed bool) {
	k.mtx.Lock()
	refs := k.refs[key]
	existed = len(refs) != 0
	nref := &KeyedRef[T]{rc: k, key: key}
	if !existed {
		_ = k.keyed.SetKey(key, true)
	}
	refs = append(refs, nref)
	k.refs[key] = refs
	k.mtx.Unlock()
	return nref, existed
}
