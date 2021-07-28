//+build !linux js

package core

import (
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
)

// addNativeFactories adds factories specific to this platform.
func addNativeFactories(b bus.Bus, sr *static.Resolver) {
	// noop
}
