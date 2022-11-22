package controller

import (
	"sync/atomic"

	"github.com/aperturerobotics/controllerbus/directive"
)

// dirRef is a directive reference
type dirRef struct {
	// released indicates this ref is released
	released atomic.Bool
	// di is the directive instance
	di *directiveInstance
	// weak indicates this is a weak ref
	weak bool
	// h is the reference handler
	h directive.ReferenceHandler
}

// Release releases the reference.
func (r *dirRef) Release() {
	if !r.released.Swap(true) {
		r.di.c.mtx.Lock()
		r.di.removeReferenceLocked(r)
		r.di.c.mtx.Unlock()
	}
}

// _ is a type assertion
var _ directive.Reference = ((*dirRef)(nil))
