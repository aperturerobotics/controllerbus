package bus_bridge

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"
)

// BusBridgeResolver resolves directives by forwarding to antihero bus.
type BusBridgeResolver struct {
	// target is the target bus
	target bus.Bus
	// dir is the directive
	dir directive.Directive
}

// NewBusBridgeResolver constructs a new BusBridgeResolver.
func NewBusBridgeResolver(target bus.Bus, dir directive.Directive) *BusBridgeResolver {
	return &BusBridgeResolver{
		target: target,
		dir:    dir,
	}
}

// Resolve resolves the values, emitting them to the handler.
func (r *BusBridgeResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	if r.target == nil || r.dir == nil {
		return nil
	}

	subCtx, subCtxCancel := context.WithCancel(ctx)
	defer subCtxCancel()

	// mtx guards below fields
	var mtx sync.Mutex
	vmap := make(map[uint32]uint32)

	_, diRef, err := r.target.AddDirective(
		r.dir,
		bus.NewCallbackHandler(
			func(av directive.AttachedValue) {
				// value added
				id, accepted := handler.AddValue(av.GetValue())
				if accepted {
					parentID := av.GetValueID()
					mtx.Lock()
					vmap[parentID] = id
					mtx.Unlock()
				}
			}, func(av directive.AttachedValue) {
				// value removed
				parentID := av.GetValueID()
				mtx.Lock()
				childID, ok := vmap[parentID]
				if ok {
					delete(vmap, parentID)
				}
				mtx.Unlock()
				if ok {
					handler.RemoveValue(childID)
				}
			},
			subCtxCancel,
		),
	)
	if err != nil {
		return err
	}
	defer diRef.Release()

	// wait for ctx cancel, then release
	select {
	case <-ctx.Done():
	case <-subCtx.Done():
	}
	return nil
}

// _ is a type assertion
var _ directive.Resolver = ((*BusBridgeResolver)(nil))
