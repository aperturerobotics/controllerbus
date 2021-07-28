package main

import (
	"context"
	"os"
	"plugin"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	"github.com/aperturerobotics/controllerbus/controller/configset/json"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/core"
	"github.com/aperturerobotics/controllerbus/directive"
	cbus_plugin "github.com/aperturerobotics/controllerbus/plugin"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := run(); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
}

const configSetYaml = `
boilerplate-demo-0:
  config:
    exampleField: testing
  id: controllerbus/example/boilerplate/1
  revision: 1
`

// run runs the program.
func run() error {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	// Create bus
	cb, _, err := core.NewCoreBus(ctx, le)
	if err != nil {
		return err
	}

	// Create the request for the controller we want /before/ loading.
	// This is just a demo of how you can do this out of order.
	// Note the config type has to be resolved in the background as well.
	resolvedCh := make(chan directive.Reference, 1)
	errCh := make(chan error, 1)
	go func() {
		// looks up the controller by ID
		// dynamically loaded from the plugin
		confSetSrc := configset.ConfigSet{}
		_, err := configset_json.UnmarshalYAML(
			ctx,
			cb,
			[]byte(configSetYaml),
			confSetSrc,
			true,
		)
		if err != nil {
			errCh <- err
		}
		_, avRef, err := bus.ExecOneOff(
			ctx,
			cb,
			resolver.NewLoadControllerWithConfig(confSetSrc["boilerplate-demo-0"].GetConfig()),
			nil,
		)
		if err != nil {
			errCh <- err
		} else {
			resolvedCh <- avRef
		}
	}()

	// Create plugin host
	// TODO

	// Load the plugin.
	pg, err := plugin.Open("./plugin/plugin-bin.so")
	if err != nil {
		return err
	}
	sym, err := pg.Lookup("ControllerBusHotPlugin")
	if err != nil {
		return err
	}
	hotPluginPtr := sym.(*cbus_plugin.Plugin)
	hotPlugin := *hotPluginPtr
	defer hotPlugin.PrePluginUnload()

	// TODO move this to common implementation
	hotResolver, err := hotPlugin.NewPluginResolver(ctx, cb)
	if err != nil {
		return err
	}
	defer hotResolver.PrePluginUnload()

	// start resolver controller
	go cb.ExecuteController(
		ctx,
		resolver.NewController(le, cb, hotResolver),
	)

	// wait for result
	select {
	case err := <-errCh:
		return err
	case res := <-resolvedCh:
		le.Info("resolved controller successfully")
		res.Release()
		return nil
	}
}
