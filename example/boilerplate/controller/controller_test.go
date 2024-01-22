package boilerplate_controller

import (
	"context"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/core"
	"github.com/aperturerobotics/controllerbus/example/boilerplate"
	boilerplate_v1 "github.com/aperturerobotics/controllerbus/example/boilerplate/v1"
	"github.com/sirupsen/logrus"
)

// TestBoilerplateController tests the boilerplate ontroller.
func TestBoilerplateController(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	b, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		t.Fatal(err.Error())
	}
	sr.AddFactory(NewFactory(b))

	execDir := resolver.NewLoadControllerWithConfig(&Config{
		ExampleField: "testing",
	})
	_, _, ctrlRef, err := bus.ExecOneOff(ctx, b, execDir, nil, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer ctrlRef.Release()

	res, _, resRef, err := bus.ExecOneOff(ctx, b, &boilerplate_v1.Boilerplate{
		MessageText: "hello world",
	}, nil, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	resRef.Release()
	plen := res.GetValue().(boilerplate.BoilerplateResult).GetPrintedLen()
	if plen != 55 {
		t.Fatalf("expected length 55 got %d", plen)
	}
	t.Log("successfully executed directive")

	// Give a moment to allow the removed callbacks to fire
	// This happens after UnrefDisposeDur
	<-time.After(time.Millisecond * 100)
}
