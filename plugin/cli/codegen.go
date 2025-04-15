package cli

import (
	"context"
	"errors"

	"github.com/aperturerobotics/cli"
	"github.com/sirupsen/logrus"
)

func (a *CompilerArgs) runCodegenOnce(cctx *cli.Context) error {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	args := cctx.Args()
	if args.Len() == 0 {
		return errors.New("specify list of packages as arguments")
	}

	an, modCompiler, cleanup, err := a.setupCompiler(ctx, le, args.Slice())
	if err != nil {
		return err
	}
	if !a.NoCleanup {
		defer cleanup()
		defer modCompiler.Cleanup()
	}

	pluginBinaryVersion := a.PluginBinaryVersion
	if pluginBinaryVersion == "" {
		pluginBinaryVersion = "cbus-plugin-{buildHash}"
	}

	return modCompiler.GenerateModules(an, pluginBinaryVersion)
}
