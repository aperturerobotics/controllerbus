package main

import (
	"context"
	"os"

	"github.com/aperturerobotics/controllerbus/bus/inmem"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	cdc "github.com/aperturerobotics/controllerbus/directive/controller"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := execToy(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func execToy() error {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	dc := cdc.NewController(ctx, le)
	b := inmem.NewBus(dc)

	// Loader controller constructs and executes controllers
	cl, err := loader.NewController(le, b)
	if err != nil {
		return err
	}

	// Execute the loader controller
	releaseCl, err := b.AddController(ctx, cl, nil)
	if err != nil {
		return err
	}
	// releaseCl() would remove the loader controller
	defer releaseCl()

	le.Debug("loader controller attached")

	// Issue directive to run the toy controller.
	loadToy := loader.NewExecController(NewToyFactory(), &ToyControllerConfig{
		Name: "world",
	})
	ctrl, _, valRef, err := loader.WaitExecControllerRunning(ctx, b, loadToy, nil)
	if err != nil {
		panic(err)
	}
	defer valRef.Release()

	tc := ctrl.(*ToyController)
	le.Debug("toy controller resolved")
	tc.SayHello()

	return nil
}
