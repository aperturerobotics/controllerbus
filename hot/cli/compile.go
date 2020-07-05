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
func (c *CompilerArgs) setupCompiler(ctx context.Context, le *logrus.Entry, paks []string) (*hot_compiler.Analysis, *hot_compiler.ModuleCompiler, func(), error) {
	rel := func() {}
	args := paks
	err := c.Validate()
	if err != nil {
		return nil, nil, rel, err
	}

	packageSearchPath, err := os.Getwd()
	if err == nil {
		packageSearchPath, err = filepath.Abs(packageSearchPath)
	}
	if err != nil {
		return nil, nil, rel, err
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

	if c.CodegenDir == "" {
		// cannot use /tmp for this, need ~/.cache dir
		// c.CodegenDir, err = ioutil.TempDir("", "cbus-codegen")
		userCacheDir, err := os.UserCacheDir()
		if err != nil {
			return nil, nil, rel, err
		}
		c.CodegenDir = filepath.Join(userCacheDir, "cbus-codegen-"+buildUid)
		le.Debugf("created tmpdir for code-gen process: %s", c.CodegenDir)
		f := rel
		rel = func() {
			defer os.RemoveAll(c.CodegenDir)
			f()
		}
	}

	codegenDirPath, err := filepath.Abs(c.CodegenDir)
	if err != nil {
		return nil, nil, rel, err
	}

	if err := os.MkdirAll(codegenDirPath, 0755); err != nil {
		return nil, nil, rel, err
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
		return nil, nil, rel, err
	}

	le.Infof("analyzing %d packages for plugin", len(args))
	an, err := hot_compiler.AnalyzePackages(ctx, le, packageSearchPath, args)
	if err != nil {
		return nil, nil, rel, err
	}

	return an, hc, rel, nil
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

	an, modCompiler, cleanup, err := c.setupCompiler(ctx, le, args)
	if err != nil {
		return err
	}
	if !c.NoCleanup {
		defer cleanup()
		defer modCompiler.Cleanup()
	}

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
