package keyed

import (
	"context"
	"sync"
)

// Routine is a function called as a goroutine.
// If nil is returned, exits cleanly permanently.
// If an error is returned, can be restarted later.
type Routine func(ctx context.Context) error

// Constructor returns a function to start for the given key.
// If nil is returned, skips starting that routine.
type Constructor func(key string) Routine

// Keyed manages a set of goroutines with associated Keys.
type Keyed struct {
	// ctx is the keyed root context
	ctx context.Context
	// ctorCb is the constructor callback
	ctorCb Constructor

	// mtx guards below fields
	mtx sync.Mutex
	// routines is the set of running routines
	routines map[string]*runningRoutine
}

// NewKeyed constructs a new Keyed execution manager.
func NewKeyed(ctx context.Context, ctorCb Constructor) *Keyed {
	if ctorCb == nil {
		ctorCb = func(key string) Routine {
			return nil
		}
	}
	return &Keyed{
		ctx:      ctx,
		ctorCb:   ctorCb,
		routines: make(map[string]*runningRoutine, 1),
	}
}

// SyncKeys synchronizes the list of running routines with the given list.
// If restart=true, restarts any failed routines in the list.
func (k *Keyed) SyncKeys(keys []string, restart bool) {
	k.mtx.Lock()
	defer k.mtx.Unlock()

	routines := make(map[string]*runningRoutine, len(keys))
	for _, key := range keys {
		v := routines[key]
		if v != nil {
			continue
		}
		v, existed := k.routines[key]
		if !existed {
			v = newRunningRoutine(k, k.ctorCb(key))
			k.routines[key] = v
		}
		routines[key] = v
		if !existed || restart {
			v.start(k.ctx)
		}
	}
	for key, rr := range k.routines {
		if _, ok := routines[key]; ok {
			continue
		}
		if rr.ctxCancel != nil {
			rr.ctxCancel()
		}
		delete(k.routines, key)
	}
}
