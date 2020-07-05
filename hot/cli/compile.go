package cli

import (
	"context"
	"crypto/sha256"
	"errors"
	"os"
	"path"
	"path/filepath"

	hot_compiler "github.com/aperturerobotics/controllerbus/hot/compiler"
	b58 "github.com/mr-tron/base58/base58"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// setupCompiler setups and creates the compiler.
func (c *CompilerArgs) setupCompiler(ctx context.Context, le *logrus.Entry, paks []string) (*hot_compiler.Analysis, *hot_compiler.ModuleCompiler, error) {
	args := paks
	if err := c.Validate(); err != nil {
		return nil, nil, err
	}
	codegenDirPath, err := filepath.Abs(c.CodegenDir)
	if err != nil {
		return nil, nil, err
	}

	if err := os.MkdirAll(codegenDirPath, 0755); err != nil {
		return nil, nil, err
	}

	packageSearchPath, err := os.Getwd()
	if err == nil {
		packageSearchPath, err = filepath.Abs(packageSearchPath)
	}
	if err != nil {
		return nil, nil, err
	}

	// deterministic prefix gen
	var buildUid string
	{
		hs := sha256.New()
		_, _ = hs.Write([]byte(packageSearchPath))
		for _, p := range args {
			_, _ = hs.Write([]byte(p))
		}
		buildUid = b58.Encode(hs.Sum(nil))
	}

	buildPrefix := c.BuildPrefix
	if buildPrefix == "" {
		buildPrefix = "cbus-hot-" + (buildUid[:8])
	}

	pluginBinaryID := c.PluginBinaryID
	if pluginBinaryID == "" {
		pluginBinaryID = "cbus-hot-" + buildUid[:8]
	}

	le.
		WithField("plugin-binary-id", pluginBinaryID).
		WithField("build-prefix", buildPrefix).
		Infof("creating compiler for plugin with packages: %v", args)
	hc, err := hot_compiler.NewModuleCompiler(ctx, le, c.BuildPrefix, codegenDirPath, c.PluginBinaryID)
	if err != nil {
		return nil, nil, err
	}

	le.Infof("analyzing %d packages for plugin", len(args))
	an, err := hot_compiler.AnalyzePackages(ctx, le, packageSearchPath, args)
	if err != nil {
		return nil, nil, err
	}

	return an, hc, nil
}

func (c *CompilerArgs) runCompileOnce(cctx *cli.Context) error {
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
	defer modCompiler.Cleanup()

	pluginBinaryVersion := c.PluginBinaryVersion
	if pluginBinaryVersion == "" {
		pluginBinaryVersion = "cbus-hot-{buildHash}"
	}

	if err := modCompiler.GenerateModules(an, pluginBinaryVersion); err != nil {
		return err
	}

	outputPath, err := filepath.Abs(c.OutputPath)
	if err == nil {
		err = os.MkdirAll(path.Dir(outputPath), 0755)
	}
	if err != nil {
		return err
	}
	return modCompiler.CompilePlugin(outputPath)
}
