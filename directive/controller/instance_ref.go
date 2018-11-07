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
		r.di.refsMtx.Lock()
		found := false
		nonWeakRefCount := 0
		for i, ref := range r.di.refs {
			if !found && ref == r {
				found = true
				r.di.refs[i] = r.di.refs[len(r.di.refs)-1]
				r.di.refs[len(r.di.refs)-1] = nil
				r.di.refs = r.di.refs[:len(r.di.refs)-1]
			} else if ref != nil && !ref.weakRef {
				nonWeakRefCount++
			}

			if found && nonWeakRefCount != 0 {
				break
			}
		}

		if r.di != nil {
			if nonWeakRefCount == 0 {
				defer r.di.Close()
			}
			r.di.refsMtx.Unlock()
		}
	})
}

// _ is a type assertion
var _ directive.Reference = ((*directiveInstanceReference)(nil))
