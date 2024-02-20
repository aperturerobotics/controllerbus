//go:build !js && !wasm
// +build !js,!wasm

package main

import (
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
	"github.com/aperturerobotics/controllerbus/directive"
	hot_loader_filesystem "github.com/aperturerobotics/controllerbus/plugin/loader/shared-library/filesystem"
	"github.com/pkg/errors"
)

// addHotLoader adds the hot loader to the bus.
// no-op on js
func addHotLoader(b bus.Bus, sr *static.Resolver) (directive.Reference, error) {
	sr.AddFactory(hot_loader_filesystem.NewFactory(b))

	_, hlRef, err := b.AddDirective(
		resolver.NewLoadControllerWithConfig(&hot_loader_filesystem.Config{
			Dir:   pluginDir,
			Watch: true,
		}),
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "construct plugin loading controller")
	}
	return hlRef, nil
}
