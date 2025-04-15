package cli

import (
	"context"
	"crypto/sha256"
	"errors"
	"os"
	"path"
	"path/filepath"

	"github.com/aperturerobotics/cli"
	plugin_compiler "github.com/aperturerobotics/controllerbus/plugin/compiler"
	b58 "github.com/mr-tron/base58/base58"
	"github.com/sirupsen/logrus"
)

// setupCompiler setups and creates the compiler.
func (a *CompilerArgs) setupCompiler(
	ctx context.Context,
	le *logrus.Entry,
	paks []string,
) (*plugin_compiler.Analysis, *plugin_compiler.ModuleCompiler, func(), error) {
	rel := func() {}
	args := paks
	err := a.Validate()
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

	if a.CodegenDir == "" {
		// cannot use /tmp for this, need ~/.cache dir
		userCacheDir, err := os.UserCacheDir()
		if err != nil {
			return nil, nil, rel, err
		}
		a.CodegenDir = filepath.Join(userCacheDir, "cbus-codegen-"+buildUid)
		le.Debugf("created tmpdir for code-gen process: %s", a.CodegenDir)
		f := rel
		rel = func() {
			defer os.RemoveAll(a.CodegenDir)
			f()
		}
	}

	codegenDirPath, err := filepath.Abs(a.CodegenDir)
	if err != nil {
		return nil, nil, rel, err
	}

	if err := os.MkdirAll(codegenDirPath, 0o755); err != nil {
		return nil, nil, rel, err
	}

	buildPrefix := a.BuildPrefix
	if buildPrefix == "" {
		buildPrefix = "cbus-plugin-" + (buildUid[:8])
	}

	pluginBinaryID := a.PluginBinaryID
	if pluginBinaryID == "" {
		pluginBinaryID = "cbus-plugin-" + buildUid[:8]
	}

	le.
		WithField("plugin-binary-id", pluginBinaryID).
		WithField("build-prefix", buildPrefix).
		Infof("creating compiler for plugin with packages: %v", args)
	hc, err := plugin_compiler.NewModuleCompiler(ctx, le, a.BuildPrefix, codegenDirPath, a.PluginBinaryID)
	if err != nil {
		return nil, nil, rel, err
	}

	le.Infof("analyzing %d packages for plugin", len(args))
	an, err := plugin_compiler.AnalyzePackages(ctx, le, packageSearchPath, args)
	if err != nil {
		return nil, nil, rel, err
	}

	return an, hc, rel, nil
}

func (a *CompilerArgs) runCompileOnce(cctx *cli.Context) error {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	args := cctx.Args().Slice()
	if len(args) == 0 {
		return errors.New("specify list of packages as arguments")
	}

	an, modCompiler, cleanup, err := a.setupCompiler(ctx, le, args)
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

	if err := modCompiler.GenerateModules(an, pluginBinaryVersion); err != nil {
		return err
	}

	outputPath, err := filepath.Abs(a.OutputPath)
	if err == nil {
		err = os.MkdirAll(path.Dir(outputPath), 0o755)
	}
	if err != nil {
		return err
	}
	return modCompiler.CompilePlugin(outputPath)
}
