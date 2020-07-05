package main

import (
	"context"
	"os"
	"path"
	"path/filepath"

	hot_compiler "github.com/aperturerobotics/controllerbus/hot/compiler"
	"github.com/sirupsen/logrus"
)

const (
	codegenDirPathRel = "./codegen-module/"
	outputPath        = "./plugins/example.{buildHash}.cbus.so"
	binaryID          = "controllerbus/examples/hot-demo/codegen-demo/1"
	binaryVersion     = "0.0.1"
)

var packagesList = []string{
	// usual demo boilerplate controller
	"github.com/aperturerobotics/controllerbus/example/boilerplate/controller",
	// example of a basic demo controller package with a non-trivial relative module reference.
	"./demo-controller",
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
	codegenDirPath, err := filepath.Abs(codegenDirPathRel)
	if err != nil {
		return err
	}
	outputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return err
	}
	// clear codegen path
	if err := os.RemoveAll(codegenDirPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(codegenDirPath, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(path.Dir(outputPath), 0755); err != nil {
		return err
	}

	packageSearchPath, _ := os.Getwd()
	packageSearchPath = filepath.Join(packageSearchPath, "..")

	le.Info("starting example watcher")
	watcher := hot_compiler.NewWatcher(le, packageSearchPath, packagesList)
	return watcher.WatchCompilePlugin(
		ctx,
		codegenDirPath,
		outputPath,
		"hot-watcher-example",
		"version-{buildHash}",
		nil,
	)
}
