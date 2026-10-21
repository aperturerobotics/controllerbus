//go:build js || wasm || wasip1

package main

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
	"github.com/aperturerobotics/controllerbus/directive"
)

// addHotLoader adds the hot loader to the bus.
// no-op on js
func addHotLoader(_ context.Context, b bus.Bus, sr *static.Resolver) (directive.Reference, error) {
	return nil, nil
}
