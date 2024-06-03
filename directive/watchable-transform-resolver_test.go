package directive_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestWatchableTransformResolver(t *testing.T) {
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
				return directive.R(directive.NewWatchableTransformResolver(ctr, func(ctx context.Context, val int) (string, bool, error) {
					if val == 0 {
						return "", false, nil
					}
					if val < 0 {
						return "", false, errors.New("negative value")
					}
					return "value is " + string(rune(val)), true, nil
				}), nil)
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
		added   string
		removed string
	}

	cbChan := make(chan cbValues, 1)
	di, dirRef, err := dc.AddDirective(dir, directive.NewCallbackHandler(func(addedValue directive.AttachedValue) {
		cbChan <- cbValues{added: addedValue.GetValue().(string)}
	}, func(removedValue directive.AttachedValue) {
		cbChan <- cbValues{removed: removedValue.GetValue().(string)}
	}, nil))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer dirRef.Release()

	// Set value and expect it to be transformed and added
	ctr.SetValue(42)
	values := <-cbChan
	if values.added != "value is *" {
		t.Errorf("Expected added value to be 'value is *', got '%s'", values.added)
	}

	// Change value and expect old value removed, new value transformed and added
	ctr.SetValue(99)
	values = <-cbChan
	if values.removed != "value is *" {
		t.Errorf("Expected removed value to be 'value is *', got '%s'", values.removed)
	}
	values = <-cbChan
	if values.added != "value is c" {
		t.Errorf("Expected added value to be 'value is c', got '%s'", values.added)
	}

	// Set value to 0 and expect it to be removed
	ctr.SetValue(0)
	values = <-cbChan
	if values.removed != "value is c" {
		t.Errorf("Expected removed value to be 'value is c', got '%s'", values.removed)
	}

	// Set negative value and expect error
	ctr.SetValue(-1)
	errCh := make(chan error, 1)

	// TODO: Fix this sequence of events:
	// - Controller starts
	// - Resolver starts
	// - Resolver adds a value
	// - Max value cap is reached, directive becomes idle
	// - Idle callback is called
	// - Resolver exits with an error
	// - Resolver is already idle so idle callbacks are not called w/ that error.
	<-time.After(time.Millisecond * 100)

	di.AddIdleCallback(func(isIdle bool, errs []error) {
		for _, err := range errs {
			if err == nil {
				continue
			}
			select {
			case errCh <- err:
			default:
			}
		}
	})
	err = <-errCh
	if err == nil || err.Error() != "negative value" {
		t.Errorf("Expected error 'negative value', got '%v'", err)
	}
}
