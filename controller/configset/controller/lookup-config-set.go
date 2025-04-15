package configset_controller

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/controller/configset"
	"github.com/aperturerobotics/controllerbus/directive"
)

// lookupConfigSetResolver is an LookupConfigSet resolver.
type lookupConfigSetResolver struct {
	// c is the controller
	c *Controller
	// ctx is the directive context
	ctx context.Context
	// di is the directive instance
	di directive.Instance
	// dir is the directive
	dir configset.LookupConfigSet

	// refsMtx guards refs
	refsMtx sync.Mutex
	// refs are the references
	refs map[configset.Reference]uint32
}

func newLookupConfigSetResolver(
	c *Controller,
	ctx context.Context,
	di directive.Instance,
	dir configset.LookupConfigSet,
) *lookupConfigSetResolver {
	r := &lookupConfigSetResolver{
		c:    c,
		ctx:  ctx,
		di:   di,
		dir:  dir,
		refs: make(map[configset.Reference]uint32),
	}
	go r.waitCleanupRefs(ctx)
	return r
}

// Resolve resolves the values, emitting them to the handler.
// Expect LookupConfigSet to not have a value limit.
func (r *lookupConfigSetResolver) Resolve(
	ctx context.Context,
	handler directive.ResolverHandler,
) error {
	ctrlKeys := r.dir.GetLookupConfigSetControllerKeys()
	if len(ctrlKeys) == 0 {
		return nil
	}

	// Catalog existing references
	existingRefs := make(map[string]directive.Reference)
	r.refsMtx.Lock()
	for ref := range r.refs {
		existingRefs[ref.GetConfigKey()] = ref
	}
	r.refsMtx.Unlock()

	// For each key/value controller config...
	for _, ctrlKey := range ctrlKeys {
		if ctrlKey == "" {
			continue
		}

		if _, ok := existingRefs[ctrlKey]; ok {
			continue
		}

		// Add reference to running instance
		ref, err := r.c.AddConfigSetReference(ctx, ctrlKey)
		if err != nil {
			if err == context.Canceled {
				return err
			}
			r.c.le.WithError(err).Warn("unable to push controller config ref")
			continue
		}

		r.refsMtx.Lock()
		r.refs[ref] = 0
		r.refsMtx.Unlock()
		ref.AddStateCallback(func(st configset.State) {
			r.refsMtx.Lock()
			v, vOk := r.refs[ref]
			if !vOk {
				r.refsMtx.Unlock()
				return
			}
			if v != 0 {
				handler.RemoveValue(v)
			}
			var val configset.LookupConfigSetValue = st //nolint:staticcheck
			id, accepted := handler.AddValue(val)
			if accepted {
				r.refs[ref] = id
			} else {
				r.refs[ref] = 0
			}
			r.refsMtx.Unlock()
		})
	}

	return nil
}

// waitCleanupRefs waits for the context to complete, then release refs
func (r *lookupConfigSetResolver) waitCleanupRefs(ctx context.Context) {
	<-ctx.Done()
	r.refsMtx.Lock()
	for ref := range r.refs {
		ref.Release()
		delete(r.refs, ref)
	}
	r.refsMtx.Unlock()
}

// _ is a type assertion
var _ directive.Resolver = ((*lookupConfigSetResolver)(nil))
