package main

import (
	"context"
	"io/ioutil"
	"os"

	cbcli "github.com/aperturerobotics/controllerbus/cli"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	configset_controller "github.com/aperturerobotics/controllerbus/controller/configset/controller"
	configset_json "github.com/aperturerobotics/controllerbus/controller/configset/json"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/core"
	hot_loader_filesystem "github.com/aperturerobotics/controllerbus/hot/loader/filesystem"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var daemonFlags cbcli.DaemonArgs

var pluginDir string

func init() {
	var dflags []cli.Flag
	dflags = append(dflags, daemonFlags.BuildFlags()...)
	dflags = append(dflags, &cli.StringFlag{
		Name:        "hot-load-dir",
		Usage:       "directory to hot-load plugins from, default `DIR`",
		Value:       "./plugins",
		Destination: &pluginDir,
	})
	commands = append(
		commands,
		cli.Command{
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
	sr.AddFactory(hot_loader_filesystem.NewFactory(b))

	// Construct hot loader
	_, hlRef, err := b.AddDirective(
		resolver.NewLoadControllerWithConfig(&hot_loader_filesystem.Config{
			Dir:   pluginDir,
			Watch: true,
		}),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "construct hot loading controller")
	}
	defer hlRef.Release()

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
		confDat, err := ioutil.ReadFile(confPath)
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
		err = ioutil.WriteFile(daemonFlags.ConfigPath, confDat, 0644)
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
