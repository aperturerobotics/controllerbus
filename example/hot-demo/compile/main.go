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
	codegenFilePath = "codegen-main.go"
	binaryID        = "controllerbus/examples/hot-demo/codegen-demo/1"
	binaryVersion   = "0.0.1"
)

var outPath = "../../../cmd/controllerbus/plugins/"
var packagesList = []string{
	"github.com/aperturerobotics/controllerbus/example/boilerplate/controller",
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
	hotWatcher := hot_compiler.NewWatcher(le, packagesList, false)
	return hotWatcher.WatchCompilePlugin(
		ctx,
		codegenFilePath,
		outPath,
		binaryID,
		binaryVersion,
		nil,
	)

}
