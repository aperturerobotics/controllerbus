//+build linux

package core

import (
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
	"github.com/aperturerobotics/controllerbus/hot/loader/filesystem"
)

// addNativeFactories adds factories specific to this platform.
func addNativeFactories(b bus.Bus, sr *static.Resolver) {
	sr.AddFactory(hot_loader_filesystem.NewFactory(b))
}
