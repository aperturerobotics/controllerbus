package configset_controller

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/controller/configset"
	"github.com/aperturerobotics/controllerbus/directive"
)

// applyConfigSetResolver is an ApplyConfigSet resolver.
type applyConfigSetResolver struct {
	// c is the controller
	c *Controller
	// ctx is the directive context
	ctx context.Context
	// di is the directive instance
	di directive.Instance
	// dir is the directive
	dir configset.ApplyConfigSet

	// refsMtx guards refs
	refsMtx sync.Mutex
	// refs are the references
	refs map[configset.Reference]uint32
}

func newApplyConfigSetResolver(
	c *Controller,
	ctx context.Context,
	di directive.Instance,
	dir configset.ApplyConfigSet,
) *applyConfigSetResolver {
	r := &applyConfigSetResolver{
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
// Expect ApplyConfigSet to not have a value limit.
func (r *applyConfigSetResolver) Resolve(
	ctx context.Context,
	handler directive.ResolverHandler,
) error {
	confSet := r.dir.GetApplyConfigSet()
	if confSet == nil {
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
	for k, v := range confSet {
		if k == "" {
			continue
		}

		if _, ok := existingRefs[k]; ok {
			continue
		}

		ref, err := r.c.PushControllerConfig(ctx, k, v)
		if err != nil {
			if err == context.Canceled {
				return err
			}
			r.c.le.WithError(err).Warn("unable to push controller config")
			continue
		}

		// Add reference to running instance
		r.refsMtx.Lock()
		r.refs[ref] = 0
		r.refsMtx.Unlock()
		ref.AddStateCb(func(st configset.State) {
			r.refsMtx.Lock()
			v, vOk := r.refs[ref]
			if !vOk {
				r.refsMtx.Unlock()
				return
			}
			if v != 0 {
				handler.RemoveValue(v)
			}
			id, accepted := handler.AddValue(configset.ApplyConfigSetValue(st))
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
func (r *applyConfigSetResolver) waitCleanupRefs(ctx context.Context) {
	<-ctx.Done()
	r.refsMtx.Lock()
	for ref := range r.refs {
		ref.Release()
		delete(r.refs, ref)
	}
	r.refsMtx.Unlock()
}

// _ is a type assertion
var _ directive.Resolver = ((*applyConfigSetResolver)(nil))
