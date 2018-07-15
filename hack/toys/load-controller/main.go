package main

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus/inmem"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/directive"
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
	cl, err := loader.NewController(le, b)
	if err != nil {
		panic(err)
	}

	// Execute the loader controller, it never exits.
	go b.ExecuteController(ctx, cl)
	le.Debug("loader controller attached")

	// Issue load directive for toy controller.
	resolved := make(chan struct{})
	loadToy := loader.NewExecControllerSingleton(NewToyFactory(), &ToyControllerConfig{
		Name: "world",
	})
	dirInst, dirRef, err := b.AddDirective(loadToy, func(val directive.Value) {
		tc := val.(*ToyController)
		le.Debug("toy controller resolved")
		tc.SayHello()
		close(resolved)
	})
	if err != nil {
		panic(err)
	}
	defer dirRef.Release()
	_ = dirInst

	// Wait for the toy controller to be resolved
	_, _ = <-resolved
}

func main() {
	execToy()
}
