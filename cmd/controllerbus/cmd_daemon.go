package main

import (
	"context"
	"os"

	"github.com/aperturerobotics/cli"
	bus_api "github.com/aperturerobotics/controllerbus/bus/api"
	api_controller "github.com/aperturerobotics/controllerbus/bus/api/controller"
	cbcli "github.com/aperturerobotics/controllerbus/cli"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	configset_controller "github.com/aperturerobotics/controllerbus/controller/configset/controller"
	configset_json "github.com/aperturerobotics/controllerbus/controller/configset/json"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/core"
	boilerplate_controller "github.com/aperturerobotics/controllerbus/example/boilerplate/controller"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var daemonFlags cbcli.DaemonArgs

var pluginDir string

func init() {
	var dflags []cli.Flag
	dflags = append(dflags, (&daemonFlags).BuildFlags()...)
	dflags = append(dflags, &cli.StringFlag{
		Name:        "hot-load-dir",
		Usage:       "path to dir to hot-load shared-object plugins",
		Value:       pluginDir,
		Destination: &pluginDir,
	})
	commands = append(
		commands,
		&cli.Command{
			Name:   "daemon",
			Usage:  "run a controller-bus daemon, loading plugins",
			Action: runDaemon,
			Flags:  dflags,
		},
	)
}

// runDaemon runs the daemon.
func runDaemon(c *cli.Context) error {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	// TODO: add hot loading controller factories here.
	b, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		return err
	}
	sr.AddFactory(api_controller.NewFactory(b))
	sr.AddFactory(boilerplate_controller.NewFactory(b))

	// Construct hot loader
	if pluginDir != "" {
		hlRef, err := addHotLoader(b, sr)
		if err != nil {
			return err
		}
		if hlRef != nil {
			defer hlRef.Release()
		}
	}

	// ConfigSet controller
	_, csRef, err := b.AddDirective(
		resolver.NewLoadControllerWithConfig(&configset_controller.Config{}),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "construct configset controller")
	}
	defer csRef.Release()

	// Construct config set.
	confSet := configset.ConfigSet{}

	// Load config file
	configLe := le.WithField("config", daemonFlags.ConfigPath)
	if confPath := daemonFlags.ConfigPath; confPath != "" {
		confDat, err := os.ReadFile(confPath)
		if err != nil {
			if os.IsNotExist(err) {
				if daemonFlags.WriteConfig {
					configLe.Info("cannot find config but write-config is set, continuing")
				} else {
					return errors.Wrapf(
						err,
						"cannot find config at %s",
						daemonFlags.ConfigPath,
					)
				}
			} else {
				return errors.Wrap(err, "load config")
			}
		}

		_, err = configset_json.UnmarshalYAML(ctx, b, confDat, confSet, true)
		if err != nil {
			return errors.Wrap(err, "unmarshal config yaml")
		}
	}

	if daemonFlags.ConfigPath != "" && daemonFlags.WriteConfig {
		confDat, err := configset_json.MarshalYAML(confSet)
		if err != nil {
			return errors.Wrap(err, "marshal config")
		}
		err = os.WriteFile(daemonFlags.ConfigPath, confDat, 0o644)
		if err != nil {
			return errors.Wrap(err, "write config file")
		}
	}

	_, configSetRef, err := b.AddDirective(
		configset.NewApplyConfigSet(confSet),
		nil,
	)
	if err != nil {
		return err
	}
	defer configSetRef.Release()

	// Daemon API
	if daemonFlags.APIListen != "" {
		_, _, apiRef, err := loader.WaitExecControllerRunning(
			ctx,
			b,
			resolver.NewLoadControllerWithConfig(&api_controller.Config{
				ListenAddr: daemonFlags.APIListen,
				BusApiConfig: &bus_api.Config{
					EnableExecController: true,
				},
			}),
			nil,
		)
		if err != nil {
			return errors.Wrap(err, "listen on api")
		}
		defer apiRef.Release()
	}

	/* TODO profiling plugin
	if daemonFlags.ProfListen != "" {
		runtime.SetBlockProfileRate(1)
		runtime.SetMutexProfileFraction(1)
		go func() {
			le.Debugf("profiling listener running: %s", daemonFlags.ProfListen)
			err := http.ListenAndServe(daemonFlags.ProfListen, nil)
			le.WithError(err).Warn("profiling listener exited")
		}()
	}
	_ = d
	*/

	// prevent deadlock here.
	<-ctx.Done()
	return nil
}
