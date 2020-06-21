package main

import (
	"context"
	"os"

	hot_compiler "github.com/aperturerobotics/controllerbus/hot/compiler"
	"github.com/sirupsen/logrus"
)

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
	wr, err := hot_compiler.GeneratePluginWrapper(
		ctx,
		le,
		"example-binary",
		"0.0.0",
		[]string{"github.com/aperturerobotics/controllerbus/example/boilerplate/controller"},
	)
	if err != nil {
		return err
	}

	le.Info("wrapper generated")
	_ = wr
	// TODO
	return nil
}
