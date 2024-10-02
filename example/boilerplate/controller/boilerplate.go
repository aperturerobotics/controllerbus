package boilerplate_controller

import (
	"context"
	"fmt"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/controllerbus/example/boilerplate"
	boilerplate_v1 "github.com/aperturerobotics/controllerbus/example/boilerplate/v1"
)

// boilerplateResolver resolves Boilerplate against the controller.
type boilerplateResolver struct {
	c   *Controller
	dir boilerplate.Boilerplate
}

// resolveBoilerplate resolves the boilerplate directive.
func (c *Controller) resolveBoilerplate(
	_ context.Context,
	_ directive.Instance,
	dir boilerplate.Boilerplate,
) ([]directive.Resolver, error) {
	return directive.Resolvers(&boilerplateResolver{c: c, dir: dir}), nil
}

// Resolve resolves the values, emitting them to the handler.
// The resolver may be canceled and restarted multiple times.
// Any fatal error resolving the value is returned.
// The resolver will not be retried after returning an error.
// Values will be maintained from the previous call.
func (b *boilerplateResolver) Resolve(
	ctx context.Context,
	handler directive.ResolverHandler,
) error {
	// Call a callback when the resolver is removed
	handler.AddResolverRemovedCallback(func() {
		b.c.le.Info("boilerplate resolver removed")
	})

	fullMsg := fmt.Sprintf(
		"logging message from boilerplate directive: %s",
		b.dir.BoilerplateMessage(),
	)
	b.c.le.Info(fullMsg)

	valID, _ := handler.AddValue(&boilerplate_v1.BoilerplateResult{
		PrintedLen: uint32(len(fullMsg)), //nolint:gosec
	})

	// call a callback when the value is removed
	handler.AddValueRemovedCallback(valID, func() {
		b.c.le.Info("boilerplate resolver value removed")
	})

	return nil
}

// _ is a type assertion
var _ directive.Resolver = ((*boilerplateResolver)(nil))
