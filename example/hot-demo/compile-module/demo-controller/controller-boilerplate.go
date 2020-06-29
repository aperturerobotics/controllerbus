package hot_compile_demo_controller

import (
	"context"
	"fmt"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/controllerbus/example/boilerplate"
	"github.com/aperturerobotics/controllerbus/example/boilerplate/v1"
)

// boilerplateResolver resolves Boilerplate against the controller.
type boilerplateResolver struct {
	c   *Controller
	dir boilerplate.Boilerplate
}

// resolveBoilerplate resolves the boilerplate directive.
func (c *Controller) resolveBoilerplate(
	ctx context.Context,
	inst directive.Instance,
	dir boilerplate.Boilerplate,
) (directive.Resolver, error) {
	return &boilerplateResolver{c: c, dir: dir}, nil
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
	fullMsg := fmt.Sprintf(
		"logging message from boilerplate directive: %s",
		b.dir.BoilerplateMessage(),
	)
	b.c.le.Info(fullMsg)
	handler.AddValue(&boilerplate_v1.BoilerplateResult{
		PrintedLen: uint32(len(fullMsg)),
	})
	return nil
}

// _ is a type assertion
var _ directive.Resolver = ((*boilerplateResolver)(nil))
