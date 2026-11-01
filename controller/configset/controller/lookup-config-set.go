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
	// Require controller keys before retaining their references.
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

		// Refuse a late reference after the directive's cleanup detached its map.
		r.refsMtx.Lock()
		if err := r.ctx.Err(); err != nil {
			r.refsMtx.Unlock()
			ref.Release()
			return err
		}
		r.refs[ref] = 0
		r.refsMtx.Unlock()

		// Publish each reference state while it remains in the active map.
		ref.AddStateCallback(func(st configset.State) {
			// Ignore a retired reference while its old callback drains.
			r.refsMtx.Lock()
			v, vOk := r.refs[ref]
			if !vOk {
				r.refsMtx.Unlock()
				return
			}

			// Replace the prior value with the current controller state.
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

// waitCleanupRefs detaches references before releasing them outside callbacks' lock order.
func (r *lookupConfigSetResolver) waitCleanupRefs(ctx context.Context) {
	// Wait until the owning directive ends.
	<-ctx.Done()

	// Detach callbacks' map before waiting on a reference's own mutex.
	r.refsMtx.Lock()
	refs := r.refs
	r.refs = make(map[configset.Reference]uint32)
	r.refsMtx.Unlock()

	// Release without refsMtx: state callbacks take the reference mutex first.
	for ref := range refs {
		ref.Release()
	}
}

// _ is a type assertion
var _ directive.Resolver = (*lookupConfigSetResolver)(nil)
