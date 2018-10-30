package controller

import (
	"github.com/aperturerobotics/controllerbus/directive"
	"sync"
)

// directiveInstanceReference implements directive reference
type directiveInstanceReference struct {
	relOnce sync.Once
	di      *DirectiveInstance
	valCb   directive.ReferenceHandler
}

// Release releases the reference.
func (r *directiveInstanceReference) Release() {
	r.relOnce.Do(func() {
		r.di.refsMtx.Lock()
		for i, ref := range r.di.refs {
			if ref == r {
				r.di.refs[i] = r.di.refs[len(r.di.refs)-1]
				r.di.refs[len(r.di.refs)-1] = nil
				r.di.refs = r.di.refs[:len(r.di.refs)-1]
				break
			}
		}

		if len(r.di.refs) == 0 {
			defer r.di.Close()
		}
		r.di.refsMtx.Unlock()
	})
}

// _ is a type assertion
var _ directive.Reference = ((*directiveInstanceReference)(nil))
