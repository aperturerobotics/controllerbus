package controller

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

// Controller manages running directives and handlers.
type Controller struct {
	// ctx is the root context for running handlers & directives
	ctx context.Context
	// le is the logger for logging events related to the controller
	le *logrus.Entry

	// mtx guards below fields
	mtx sync.Mutex
	// dirID contains the next directive instance id
	dirID uint32
	// dir contains the list of running directive instances
	// sorted by ID
	dir []*directiveInstance
	// hnd contains the list of attached handlers
	hnd []*handler
}

// NewController builds a new directive controller.
func NewController(ctx context.Context, le *logrus.Entry) *Controller {
	return &Controller{
		ctx: ctx,
		le:  le,
	}
}

// GetDirectives returns a list of all currently executing directives.
func (c *Controller) GetDirectives() []directive.Instance {
	c.mtx.Lock()
	dirs := make([]directive.Instance, len(c.dir))
	for i, di := range c.dir {
		dirs[i] = di
	}
	c.mtx.Unlock()
	return dirs
}

// AddDirective adds a directive to the controller.
// This call de-duplicates equivalent directives.
//
// cb receives values in order as they are emitted.
// cb can be nil.
// cb should not block.
//
// Returns the instance, new reference, and anynke error.
func (c *Controller) AddDirective(
	dir directive.Directive,
	ref directive.ReferenceHandler,
) (directive.Instance, directive.Reference, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Check if any equivalent directives exist, if applicable.
	if eqDir, eqDirOk := dir.(directive.DirectiveWithEquiv); eqDirOk {
		for diIdx, di := range c.dir {
			if !di.released.Load() && eqDir.IsEquivalent(di.dir) {
				// Directive is equivalent to di.
				spDir, spDirOk := eqDir.(directive.DirectiveWithSuperceeds)
				if spDirOk && spDir.Superceeds(di.dir) {
					// Remove the other directive (superceed it)
					di.removeLocked(diIdx)
				} else {
					// Add a reference to the other directive & return.
					return di, di.addReferenceLocked(ref, false), nil
				}
			}
		}
	}

	// Push the new directive to the list.
	di, diRef := newDirectiveInstance(c, c.dirID, dir, ref)
	c.dirID++
	di.logger().Debug("added directive")
	c.dir = append(c.dir, di)

	// call all handlers while mtx is unlocked
	handlers := make([]*handler, len(c.hnd))
	copy(handlers, c.hnd)
	c.mtx.Unlock()
	var resolvers []*resolver
	for _, hnd := range handlers {
		if !hnd.rel.Load() {
			nres, err := di.callHandlerUnlocked(hnd)
			if err != nil {
				di.logger().
					WithError(err).
					Warn("directive handler returned error")
				continue
			}
			resolvers = append(resolvers, nres...)
		}
	}
	// attach returned resolvers while mtx is locked
	c.mtx.Lock()
	for _, res := range resolvers {
		di.attachStartResolverLocked(res)
	}
	return di, diRef, nil
}

// AddHandler adds a directive handler.
// The handler will receive calls for all existing directives (initial set).
// An error is returned only if adding the handler failed.
// Returns a function to remove the handler.
// The release function must be non-nil if err is nil, and nil if err != nil.
func (c *Controller) AddHandler(handler directive.Handler) (func(), error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.addHandlerLocked(handler)
}

// addHandlersLocked adds a set of handlers while c.mtx is locked.
func (c *Controller) addHandlerLocked(handler directive.Handler) (func(), error) {
	hnd := newHandler(handler)
	c.hnd = append(c.hnd, hnd)

	dirs := make([]*directiveInstance, len(c.dir))
	copy(dirs, c.dir)
	// unlock temporarily
	c.mtx.Unlock()

	var resolvers []*resolver
	var dis []*directiveInstance
	for _, di := range dirs {
		if !hnd.rel.Load() {
			nres, err := di.callHandlerUnlocked(hnd)
			if err != nil {
				di.logger().
					WithError(err).
					Warn("directive handler returned error")
				continue
			}
			resolvers = append(resolvers, nres...)
			if len(nres) != 0 {
				oldLen := len(dis)
				dis = slices.Grow(dis, len(nres))[:len(dis)+len(nres)]
				for i := oldLen; i < len(dis); i++ {
					dis[i] = di
				}
			}
		}
	}

	// attach returned resolvers while mtx is locked
	c.mtx.Lock()
	for i, res := range resolvers {
		dis[i].attachStartResolverLocked(res)
	}
	relHandler := func() {
		if !hnd.rel.Swap(true) {
			c.mtx.Lock()
			c.removeHandlerLocked(hnd)
			c.mtx.Unlock()
		}
	}

	return relHandler, nil
}

// removeHandlerLocked removes a handler while c.mtx is locked
//
// returns if the handler was found in the set
func (c *Controller) removeHandlerLocked(hnd *handler) bool {
	// mark released
	hnd.rel.Store(true)
	// remove from list
	var found bool
	for i, ih := range c.hnd {
		if ih == hnd {
			c.hnd[i] = c.hnd[len(c.hnd)-1]
			c.hnd[len(c.hnd)-1] = nil
			c.hnd = c.hnd[:len(c.hnd)-1]
			found = true
			break
		}
	}
	if !found {
		return false
	}
	// cancel all resolvers associated with the handler
	for _, di := range c.dir {
		di.removeHandlerLocked(hnd)
	}
	return true
}

// _ is a type assertion
var _ directive.Controller = ((*Controller)(nil))
