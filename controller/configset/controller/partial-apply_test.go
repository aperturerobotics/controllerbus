package configset_controller_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	configset_controller "github.com/aperturerobotics/controllerbus/controller/configset/controller"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/core"
	"github.com/aperturerobotics/controllerbus/directive"
	boilerplate "github.com/aperturerobotics/controllerbus/example/boilerplate/controller"
	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"
)

// partialApplyController is a minimal controller used by partialApplyFactory.
// It blocks in Execute until the context is canceled.
type partialApplyController struct {
	le *logrus.Entry
}

// GetControllerInfo returns information about the controller.
func (c *partialApplyController) GetControllerInfo() *controller.Info {
	return controller.NewInfo(
		(&config.Placeholder{}).GetConfigID(),
		semver.MustParse("0.0.1"),
		"partial-apply test controller",
	)
}

// HandleDirective never handles any directives.
func (c *partialApplyController) HandleDirective(
	ctx context.Context,
	di directive.Instance,
) ([]directive.Resolver, error) {
	return nil, nil
}

// Execute blocks until ctx is canceled.
func (c *partialApplyController) Execute(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Close releases resources.
func (c *partialApplyController) Close() error { return nil }

// partialApplyFactory constructs partialApplyController instances using the
// config.Placeholder config type.
type partialApplyFactory struct{}

// GetConfigID returns the config id this factory handles.
func (f *partialApplyFactory) GetConfigID() string {
	return (&config.Placeholder{}).GetConfigID()
}

// ConstructConfig constructs the empty config object.
func (f *partialApplyFactory) ConstructConfig() config.Config { return &config.Placeholder{} }

// GetVersion returns the factory version.
func (f *partialApplyFactory) GetVersion() semver.Version { return semver.MustParse("0.0.1") }

// Construct returns a controller for the given config.
func (f *partialApplyFactory) Construct(
	ctx context.Context,
	cc config.Config,
	opts controller.ConstructOpts,
) (controller.Controller, error) {
	return &partialApplyController{le: opts.GetLogger()}, nil
}

// _ is a type assertion
var _ controller.Factory = ((*partialApplyFactory)(nil))

// _ is a type assertion
var _ controller.Controller = ((*partialApplyController)(nil))

// TestPartialConfigSetApply exercises the partial-apply hypothesis: applying
// a ConfigSet with one missing factory should NOT error, the present factory
// should start, and registering the missing factory later should start the
// second controller.
func TestPartialConfigSetApply(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Register only factory A initially; factory B is intentionally missing.
	sr.AddFactory(boilerplate.NewFactory(b))

	// Boot the configset controller.
	_, _, csCtrlRef, err := bus.ExecOneOff(
		ctx,
		b,
		resolver.NewLoadControllerWithConfig(&configset_controller.Config{}),
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer csCtrlRef.Release()

	cs := make(configset.ConfigSet)
	cs["ctrl-a"] = configset.NewControllerConfig(1, &boilerplate.Config{
		ExampleField: "controller A is present",
	})
	cs["ctrl-b"] = configset.NewControllerConfig(1, &config.Placeholder{})

	// Track latest state per controller key. Values flow from the
	// ApplyConfigSet directive callback handler.
	var (
		stateMtx sync.Mutex
		latest   = make(map[string]configset.State)
		updates  = make(chan struct{}, 64)
	)
	pushUpdate := func(st configset.State) {
		stateMtx.Lock()
		latest[st.GetId()] = st
		stateMtx.Unlock()
		select {
		case updates <- struct{}{}:
		default:
		}
	}

	// Apply the ConfigSet via the ApplyConfigSet directive.
	dir := configset.NewApplyConfigSet(cs)
	_, applyRef, err := b.AddDirective(
		dir,
		bus.NewCallbackHandler(
			func(av directive.AttachedValue) {
				st, ok := av.GetValue().(configset.ApplyConfigSetValue)
				if !ok {
					return
				}
				pushUpdate(st)
			},
			nil,
			nil,
		),
	)
	if err != nil {
		t.Fatalf("AddDirective: %v", err)
	}
	defer applyRef.Release()

	getState := func(key string) configset.State {
		stateMtx.Lock()
		defer stateMtx.Unlock()
		return latest[key]
	}

	waitFor := func(check func() (bool, string)) (string, bool) {
		for {
			ok, msg := check()
			if ok {
				return msg, true
			}
			select {
			case <-ctx.Done():
				return msg, false
			case <-updates:
			}
		}
	}

	// Wait for controller A to come up running.
	if msg, ok := waitFor(func() (bool, string) {
		st := getState("ctrl-a")
		if st == nil {
			return false, "ctrl-a state not yet observed"
		}
		if err := st.GetError(); err != nil {
			return true, "ctrl-a errored: " + err.Error()
		}
		if st.GetController() == nil {
			return false, "ctrl-a controller not yet running"
		}
		return true, "ctrl-a running"
	}); !ok {
		t.Fatalf("timed out waiting for ctrl-a: %s", msg)
	} else {
		t.Log(msg)
	}

	// Sanity: at this point ctrl-b should NOT have a running controller.
	if st := getState("ctrl-b"); st != nil && st.GetController() != nil {
		t.Fatalf("ctrl-b unexpectedly already running before factory registered")
	}

	// Now register the previously-missing factory and verify ctrl-b comes up.
	sr.AddFactory(&partialApplyFactory{})

	if msg, ok := waitFor(func() (bool, string) {
		st := getState("ctrl-b")
		if st == nil {
			return false, "ctrl-b state not yet observed"
		}
		if err := st.GetError(); err != nil {
			return true, "ctrl-b errored: " + err.Error()
		}
		if st.GetController() == nil {
			return false, "ctrl-b controller not yet running"
		}
		return true, "ctrl-b running"
	}); !ok {
		t.Fatalf("timed out waiting for ctrl-b after AddFactory: %s", msg)
	} else {
		t.Log(msg)
	}
}
