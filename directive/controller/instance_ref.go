package controller

import (
	"sync"

	"github.com/aperturerobotics/controllerbus/directive"
)

// directiveInstanceReference implements directive reference
type directiveInstanceReference struct {
	relOnce sync.Once
	di      *DirectiveInstance
	valCb   directive.ReferenceHandler
	weakRef bool
}

// Release releases the reference.
func (r *directiveInstanceReference) Release() {
	r.relOnce.Do(func() {
		r.di.releaseReference(r)
	})
}

// _ is a type assertion
var _ directive.Reference = ((*directiveInstanceReference)(nil))
