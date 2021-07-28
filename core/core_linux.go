// +build linux
// +build !js

package core

import (
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
	// hot_loader_filesystem  "github.com/aperturerobotics/controllerbus/plugin/loader/shared-library/filesystem"
)

// addNativeFactories adds factories specific to this platform.
func addNativeFactories(b bus.Bus, sr *static.Resolver) {
	// sr.AddFactory(hot_loader_filesystem.NewFactory(b))
}
