package core

import (
	"context"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/bus/inmem"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
	cdc "github.com/aperturerobotics/controllerbus/directive/controller"
	"github.com/sirupsen/logrus"
)

// NewCoreBus constructs a standard in-memory bus stack.
func NewCoreBus(
	ctx context.Context,
	le *logrus.Entry,
	builtInFactories ...controller.Factory,
) (bus.Bus, *static.Resolver, error) {
	dc := cdc.NewDirectiveController(ctx, le)
	b := inmem.NewBus(dc)

	// Loader controller constructs and executes controllers
	cl, err := loader.NewController(le, b)
	if err != nil {
		return nil, nil, err
	}

	// Execute the loader controller.
	go b.ExecuteController(ctx, cl)

	// If there are any built in factories append them.
	sr := static.NewResolver(builtInFactories...)
	go b.ExecuteController(
		ctx,
		resolver.NewController(b, sr),
	)

	return b, sr, nil
}
