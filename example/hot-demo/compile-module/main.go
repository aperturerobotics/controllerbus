package main

import (
	"context"
	"go/build"
	"go/parser"
	"os"
	"path"

	hot_compiler "github.com/aperturerobotics/controllerbus/hot/compiler"
	"github.com/sirupsen/logrus"
	"golang.org/x/tools/go/loader"
)

const (
	codegenDirPath = "./codegen-module/"
	binaryID       = "controllerbus/examples/hot-demo/codegen-demo/1"
	binaryVersion  = "0.0.1"
)

var outPath = "../../../cmd/controllerbus/plugins/"
var packagesList = []string{
	// usual demo boilerplate controller
	"github.com/aperturerobotics/controllerbus/example/boilerplate/controller",
	// example of a basic demo controller package with a non-trivial relative module reference.
	"compile-module/demo-controller",
	// example of a cross-module reference to a package that does not contain a
	// factory, but rather is a indirect dependency. "lifting" packages like
	// this is necessary for hot-loading packages which reference newer versions
	// of utility packages. this adds thee same plugin id prefix to the package
	// and will load a copy of the package unique to the plugin.
	"github.com/aperturerobotics/controllerbus/hot/plugin",
}

func main() {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	if err := run(ctx, le); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
}

func run(ctx context.Context, le *logrus.Entry) error {
	// clear codegen path
	if err := os.RemoveAll(codegenDirPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	le.Info("detecting the path to cmd/controllerbus...")
	// find the command line package path
	var conf loader.Config
	conf.Build = &build.Default
	conf.AllowErrors = true
	conf.ParserMode |= parser.PackageClauseOnly
	conf.Import("github.com/aperturerobotics/controllerbus/cmd/controllerbus")
	conf.Cwd, _ = os.Getwd()
	prog, err := conf.Load()
	if err != nil {
		return err
	}

	{
		pakInfo := prog.InitialPackages()[0]
		outPath = path.Dir(prog.Fset.File(pakInfo.Files[0].Pos()).Name())
		outPath = path.Join(outPath, "plugins/boilerplate-example.{buildHash}.cbus.so")
		le.Infof("detected controllerbus cli plugins path: %s", outPath)
	}

	le.Info("compiling example package")
	/*
		hotWatcher := hot_compiler.NewWatcher(le, packagesList, false)
		return hotWatcher.WatchCompilePlugin(
			ctx,
			codegenDirPath,
			outPath,
			binaryID,
			binaryVersion,
			nil,
		)
	*/
	if err := os.MkdirAll(codegenDirPath, 0755); err != nil {
		return err
	}

	packagesLookupPath, _ := os.Getwd()
	moduleCompiler, err := hot_compiler.NewModuleCompiler(
		ctx,
		le,
		packagesList,
		packagesLookupPath,
		codegenDirPath,
		outPath,
		"hot-demo-module",
		nil, // preWriteOutFileHook func(nextOutFilePath, nextOutFileContentsPath string) error)
	)
	if err != nil {
		return err
	}
	analysis, err := moduleCompiler.BuildAnalysis()
	if err != nil {
		return err
	}
	cleanupFiles := false
	if err := moduleCompiler.GenerateModules(analysis, "hot-module-abcdef", cleanupFiles); err != nil {
		return err
	}

	return moduleCompiler.CompilePlugin()
}
