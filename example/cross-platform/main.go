package main

import (
	"context"
	"os"
	"runtime"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/core"
	"github.com/aperturerobotics/controllerbus/example/boilerplate"
	boilerplate_controller "github.com/aperturerobotics/controllerbus/example/boilerplate/controller"
	boilerplate_v1 "github.com/aperturerobotics/controllerbus/example/boilerplate/v1"
	"github.com/sirupsen/logrus"
)

func Run(ctx context.Context, le *logrus.Entry) error {
	b, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		return err
	}
	sr.AddFactory(boilerplate_controller.NewFactory(b))

	execDir := resolver.NewLoadControllerWithConfig(&boilerplate_controller.Config{
		ExampleField: "hello cross-platform",
	})
	_, ctrlRef, err := bus.ExecOneOff(ctx, b, execDir, false, nil)
	if err != nil {
		return err
	}
	defer ctrlRef.Release()

	res, resRef, err := bus.ExecOneOff(ctx, b, &boilerplate_v1.Boilerplate{
		MessageText: "hello from a directive",
	}, false, nil)
	if err != nil {
		return err
	}
	resRef.Release()
	plen := res.GetValue().(boilerplate.BoilerplateResult).GetPrintedLen()
	le.Infof("successfully executed directive, logged %d chars", plen)
	return nil
}

func main() {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)
	if err := Run(ctx, le); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}

	// to prevent an error, prevent program from exiting in js
	if runtime.GOOS == "js" {
		le.Info("program complete")
		<-ctx.Done()
	}
}
