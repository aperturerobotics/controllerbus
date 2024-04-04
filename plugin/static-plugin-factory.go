package plugin

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
)

// StaticPluginFactory wraps a factory with a pre-close hook.
type StaticPluginFactory struct {
	controller.Factory
	ctx       context.Context
	ctxCancel context.CancelFunc
	binaryID  string
	bus       bus.Bus
	mtx       sync.Mutex
	preUnload []func()
}

// NewStaticPluginFactory constructs a new static plugin factory.
func NewStaticPluginFactory(
	ctx context.Context,
	ft controller.Factory,
	binaryID string,
	bus bus.Bus,
) *StaticPluginFactory {
	nctx, nctxCancel := context.WithCancel(ctx)
	return &StaticPluginFactory{
		Factory:   ft,
		binaryID:  binaryID,
		bus:       bus,
		ctx:       nctx,
		ctxCancel: nctxCancel,
	}
}

// GetContext returns a context that is canceled if the factory is unloaded.
func (s *StaticPluginFactory) GetFactoryContext() context.Context {
	return s.ctx
}

// Construct constructs the associated controller given configuration.
func (s *StaticPluginFactory) Construct(
	ctx context.Context,
	conf config.Config,
	opts controller.ConstructOpts,
) (controller.Controller, error) {
	subController, err := s.Factory.Construct(ctx, conf, opts)
	if err != nil {
		return subController, err
	}

	// Augment logging fields.
	opts.Logger = opts.Logger.WithField("plugin-binary-id", s.binaryID)

	// Ensure that the controller is removed just before we unload.
	s.mtx.Lock()
	s.preUnload = append(s.preUnload, func() {
		s.bus.RemoveController(subController)
	})
	s.mtx.Unlock()
	return subController, err
}

// Close closes the static plugin factory.
func (s *StaticPluginFactory) Close() {
	s.ctxCancel()
	s.mtx.Lock()
	for _, f := range s.preUnload {
		f()
	}
	s.preUnload = nil
	s.mtx.Unlock()
}

// _ is a type assertion
var _ controller.FactoryWithContext = ((*StaticPluginFactory)(nil))
