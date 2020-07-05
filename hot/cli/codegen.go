package cli

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func (c *CompilerArgs) runCodegenOnce(cctx *cli.Context) error {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	args := cctx.Args()
	if len(args) == 0 {
		return errors.New("specify list of packages as arguments")
	}

	an, modCompiler, err := c.setupCompiler(ctx, le, args)
	if err != nil {
		return err
	}

	pluginBinaryVersion := c.PluginBinaryVersion
	if pluginBinaryVersion == "" {
		pluginBinaryVersion = "cbus-hot-{buildHash}"
	}

	return modCompiler.GenerateModules(an, pluginBinaryVersion)
}
