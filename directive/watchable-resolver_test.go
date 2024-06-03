package directive_test

import (
	"context"
	"testing"

	"github.com/aperturerobotics/controllerbus/bus/inmem"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/callback"
	"github.com/aperturerobotics/controllerbus/directive"
	cdc "github.com/aperturerobotics/controllerbus/directive/controller"
	boilerplate_v1 "github.com/aperturerobotics/controllerbus/example/boilerplate/v1"
	"github.com/aperturerobotics/util/ccontainer"
	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
)

func TestWatchableResolver(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	dc := cdc.NewController(ctx, le)
	b := inmem.NewBus(dc)

	ctr := ccontainer.NewCContainer(0)
	testCtrl := callback.NewCallbackController(
		controller.NewInfo("test", semver.MustParse("0.0.0"), "test controller"),
		nil,
		func(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
			switch di.GetDirective().(type) {
			case *boilerplate_v1.Boilerplate:
				return directive.R(directive.NewWatchableResolver(ctr), nil)
			}
			return nil, nil
		},
		nil,
	)

	relCtrl, err := b.AddController(ctx, testCtrl, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer relCtrl()

	dir := &boilerplate_v1.Boilerplate{MessageText: "test"}

	type cbValues struct {
		added   int
		removed int
	}

	cbChan := make(chan cbValues, 1)
	_, dirRef, err := dc.AddDirective(dir, directive.NewCallbackHandler(func(addedValue directive.AttachedValue) {
		cbChan <- cbValues{added: addedValue.GetValue().(int)}
	}, func(removedValue directive.AttachedValue) {
		cbChan <- cbValues{removed: removedValue.GetValue().(int)}
	}, nil))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer dirRef.Release()

	// Set value and expect it to be added
	ctr.SetValue(42)
	values := <-cbChan
	if values.added != 42 {
		t.Errorf("Expected added value to be 42, got %d", values.added)
	}
	if values.removed != 0 {
		t.Errorf("Expected removed value to be 0, got %d", values.removed)
	}

	// Change value and expect old value removed, new value added
	ctr.SetValue(99)
	values = <-cbChan
	if values.removed != 42 {
		t.Errorf("Expected removed value to be 42, got %d", values.removed)
	}
	values = <-cbChan
	if values.added != 99 {
		t.Errorf("Expected added value to be 99, got %d", values.added)
	}

	// Clear value and expect it to be removed
	var empty int
	ctr.SetValue(empty)
	values = <-cbChan
	if values.removed != 99 {
		t.Errorf("Expected removed value to be 99, got %d", values.removed)
	}
}
