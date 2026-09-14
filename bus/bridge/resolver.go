package bus_bridge

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"
)

// BusBridgeResolver resolves directives by forwarding to another bus.
type BusBridgeResolver struct {
	// target is the target bus.
	target bus.Bus
	// dir is the directive forwarded to the target bus.
	dir directive.Directive
}

// NewBusBridgeResolver constructs a new BusBridgeResolver.
func NewBusBridgeResolver(target bus.Bus, dir directive.Directive) *BusBridgeResolver {
	return &BusBridgeResolver{
		target: target,
		dir:    dir,
	}
}

// Resolve forwards values, idle state, and resolver errors until either side closes.
func (r *BusBridgeResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	// An unconfigured bridge has no work to forward.
	if r.target == nil || r.dir == nil {
		return nil
	}

	// Retain the target directive and map its values into the source resolver.
	subCtx, subCtxCancel := context.WithCancel(ctx)
	defer subCtxCancel()

	// mtx guards the mapping from target value IDs to source value IDs.
	var mtx sync.Mutex
	vmap := make(map[uint32]uint32)

	di, diRef, err := r.target.AddDirective(
		r.dir,
		bus.NewCallbackHandler(
			func(av directive.AttachedValue) {
				// Publish the target value and retain its source ID for removal.
				id, accepted := handler.AddValue(av.GetValue())
				if accepted {
					parentID := av.GetValueID()
					mtx.Lock()
					vmap[parentID] = id
					mtx.Unlock()
				}
			}, func(av directive.AttachedValue) {
				// Remove the corresponding source value when the target retracts it.
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

	// Optional lookups can settle only when the target's idle state reaches them.
	errCh := make(chan error, 1)
	defer di.AddIdleCallback(func(idle bool, errs []error) {
		for _, err := range errs {
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
		}
		handler.MarkIdle(idle)
	})()

	// Release the bridge on cancellation, target disposal, or a target failure.
	select {
	case <-ctx.Done():
	case <-subCtx.Done():
	case err := <-errCh:
		return err
	}
	return nil
}

// _ is a type assertion
var _ directive.Resolver = (*BusBridgeResolver)(nil)
