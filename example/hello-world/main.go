package main

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus/inmem"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	cdc "github.com/aperturerobotics/controllerbus/directive/controller"
	"github.com/sirupsen/logrus"
)

func execToy() {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	dc := cdc.NewDirectiveController(ctx, le)
	b := inmem.NewBus(dc)

	// Loader controller constructs and executes controllers
	cl, err := loader.NewController(le, b, nil)
	if err != nil {
		panic(err)
	}

	// Execute the loader controller, it never exits.
	go func() {
		_ = b.ExecuteController(ctx, cl)
	}()
	le.Debug("loader controller attached")

	// Issue load directive for toy controller.
	loadToy := loader.NewExecControllerSingleton(NewToyFactory(), &ToyControllerConfig{
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
}

func main() {
	execToy()
}
