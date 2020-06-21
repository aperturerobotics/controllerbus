package main

import (
	"context"
	"os"

	hot_compiler "github.com/aperturerobotics/controllerbus/hot/compiler"
	"github.com/sirupsen/logrus"
)

var outPath = "./hot-demo.cbus.so"

func main() {
	if err := run(); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	le.Info("compiling example package")
	// TODO GeneratePluginWrapperGoMod
	packagePaths := []string{
		"github.com/aperturerobotics/controllerbus/example/boilerplate/controller",
	}
	an, err := hot_compiler.AnalyzePackages(le, packagePaths)
	if err != nil {
		return err
	}
	wr, err := hot_compiler.GeneratePluginWrapper(
		ctx,
		le,
		an,
		"example-binary",
		"0.0.0",
	)
	if err != nil {
		return err
	}
	err = hot_compiler.CompilePluginFromFile(le, wr, "codegen-main.go", outPath)
	if err != nil {
		return err
	}

	le.Info("wrapper generated")
	_ = wr
	// TODO
	return nil
}
