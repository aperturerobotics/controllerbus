package controller

import (
	"github.com/aperturerobotics/controllerbus/directive"
	"sync"
)

// DirectiveInstance implements the directive instance interface.
type DirectiveInstance struct {
	// dir is the underlying directive.
	dir directive.Directive

	// refsMtx guards refs
	refsMtx sync.Mutex
	// refs is the list of references
	refs []*directiveInstanceReference

	// valsMtx guards vals
	valsMtx sync.Mutex
	// vals is the list of emitted values
	vals []directive.Value

	// rel is the released cb
	rel func()
	// relOnce ensures rel is called once
	relOnce sync.Once
}

// directiveInstanceReference implements directive reference
type directiveInstanceReference struct {
	relOnce sync.Once
	di      *DirectiveInstance
	valCb   func(directive.Value)
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

// NewDirectiveInstance constructs a new directive instance with an initial reference.
func NewDirectiveInstance(dir directive.Directive, cb func(directive.Value), released func()) (*DirectiveInstance, directive.Reference) {
	i := &DirectiveInstance{dir: dir, rel: released}
	ref := &directiveInstanceReference{di: i, valCb: cb}
	i.refs = append(i.refs, ref)
	return i, ref
}

// AddReference adds a reference to the directive.
// Will return nil if the directive is already expired.
func (r *DirectiveInstance) AddReference(cb func(directive.Value)) directive.Reference {
	ref := &directiveInstanceReference{
		di:    r,
		valCb: cb,
	}

	r.refsMtx.Lock()
	if len(r.refs) == 0 {
		ref = nil
	} else {
		r.valsMtx.Lock()
		// avoid calling cb() before returning addreference
		defer func() {
			go func() {
				// cb should not block
				for _, v := range r.vals {
					cb(v)
				}
				r.valsMtx.Unlock()
			}()
		}()

		r.refs = append(r.refs, ref)
	}
	r.refsMtx.Unlock()

	return ref
}

// emitValue emits a value to all listeners
func (r *DirectiveInstance) emitValue(v directive.Value) {
	r.valsMtx.Lock()
	r.vals = append(r.vals, v)
	r.valsMtx.Unlock()

	r.refsMtx.Lock()
	for _, ref := range r.refs {
		ref.valCb(v)
	}
	r.refsMtx.Unlock()
}

// GetDirective returns the underlying directive.
func (r *DirectiveInstance) GetDirective() directive.Directive {
	return r.dir
}

// Close cancels the directive instance.
func (r *DirectiveInstance) Close() {
	r.relOnce.Do(func() {
		r.refsMtx.Lock()
		r.refs = nil
		r.refsMtx.Unlock()

		r.valsMtx.Lock()
		r.vals = nil
		r.valsMtx.Unlock()

		if r.rel != nil {
			r.rel()
		}
	})
}

// _ is a type assertion
var _ directive.Instance = ((*DirectiveInstance)(nil))

// _ is a type assertion
var _ directive.Reference = ((*directiveInstanceReference)(nil))
