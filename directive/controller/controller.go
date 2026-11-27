package controller

import (
	"context"
	"slices"
	"sync"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// Controller manages running directives and handlers.
type Controller struct {
	// ctx is the root context for running handlers & directives
	ctx context.Context
	// le is the logger for logging events related to the controller
	le *logrus.Entry
	// mtx guards below fields
	mtx sync.Mutex
	// changed is closed when the directives or their references next change.
	// It is nil until a caller waits, and again once closed.
	changed chan struct{}
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
	if le == nil {
		le = logrus.New().WithField("component", "directive-controller")
	}
	return &Controller{
		ctx: ctx,
		le:  le,
	}
}

// GetDirectives returns a list of all currently executing directives.
func (c *Controller) GetDirectives() []directive.Instance {
	// Snapshot the current directive registrations under the controller lock.
	c.mtx.Lock()
	dirs := make([]directive.Instance, len(c.dir))
	for i, di := range c.dir {
		dirs[i] = di
	}
	c.mtx.Unlock()
	return dirs
}

// GetDirectivesWaitCh returns a channel closed when a directive is next added
// or removed, or a reference to a directive is next added or released.
//
// A watcher takes the channel, reads the directives, and waits on the channel:
// a change during the read closes the channel already taken. The controller's
// own mutex guards the channel, so a watcher holds no lock while it reads.
func (c *Controller) GetDirectivesWaitCh() <-chan struct{} {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.changed == nil {
		c.changed = make(chan struct{})
	}
	return c.changed
}

// broadcastLocked wakes the directives watchers while c.mtx is locked.
func (c *Controller) broadcastLocked() {
	if c.changed != nil {
		close(c.changed)
		c.changed = nil
	}
}

// AddDirective adds a directive to the controller.
// This call de-duplicates equivalent directives.
//
// cb receives values in order as they are emitted.
// cb can be nil.
// cb should not block.
//
// Returns the instance, new reference, and any error.
func (c *Controller) AddDirective(
	ctx context.Context,
	dir directive.Directive,
	ref directive.ReferenceHandler,
) (directive.Instance, directive.Reference, error) {
	// Read the caller origin before changing directive registrations.
	origin := directive.OriginFromContext(ctx)
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
					return di, di.addReferenceLocked(origin, ref, false), nil
				}
			}
		}
	}

	// Push the new directive to the list.
	di, diRef := newDirectiveInstance(c, c.dirID, dir, origin, ref)
	c.dirID++
	di.logger().Debug("added directive")
	c.dir = append(c.dir, di)

	// signal directive list changed
	c.broadcastLocked()

	// Defer calling state changed callback.
	defer di.deferCheckStateChanged()()

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
	c.mtx.Lock() // note: unlocked in Defer above

	// Attach returned resolvers and publish readiness for the live directive.
	if !di.released.Load() {
		for _, res := range resolvers {
			di.attachStartResolverLocked(res)
		}

		// mark instance as ready (can be marked idle)
		di.ready = true

		// Check if idle state changed after setting ready=true
		// This ensures idle callbacks are called if idle became true when ready became true
		di.handleIdleStateLocked()
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
	// Publish the handler before calling it for existing directives.
	hnd := newHandler(handler)
	c.hnd = append(c.hnd, hnd)

	// Snapshot the directives that the new handler must receive.
	dirs := make([]*directiveInstance, len(c.dir))
	copy(dirs, c.dir)

	// Call handlers without holding the controller mutex.
	c.mtx.Unlock()

	// Collect each returned resolver with the directive it belongs to.
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
			if len(nres) != 0 {
				resolvers = append(resolvers, nres...)

				// We have to use this loop instead of append because the type changes.
				oldLen := len(dis)
				dis = slices.Grow(dis, len(nres))[:oldLen+len(nres)]
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
		// Serialize removal with resolver attachment and startup.
		c.mtx.Lock()
		defer c.mtx.Unlock()
		if !hnd.rel.Load() {
			c.removeHandlerLocked(hnd)
		}
	}

	return relHandler, nil
}

// removeHandlerLocked removes a handler while c.mtx is locked
//
// returns if the handler was found in the set
func (c *Controller) removeHandlerLocked(hnd *handler) bool {
	// Take the registration out of the handler set.
	idx := slices.Index(c.hnd, hnd)
	if idx < 0 {
		return false
	}
	last := len(c.hnd) - 1
	c.hnd[idx], c.hnd[last] = c.hnd[last], nil
	c.hnd = c.hnd[:last]

	// Retire it, so no pending call attaches resolvers, and cancel its resolvers.
	hnd.retire(c.dir)
	return true
}

// _ is a type assertion
var _ directive.Controller = (*Controller)(nil)
