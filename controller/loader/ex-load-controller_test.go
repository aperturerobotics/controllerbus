package loader_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/core"
	"github.com/aperturerobotics/controllerbus/directive"
	boilerplate_controller "github.com/aperturerobotics/controllerbus/example/boilerplate/controller"
	"github.com/sirupsen/logrus"
)

// failingController runs until fail is closed, then exits with an error.
type failingController struct {
	fail <-chan struct{}
}

func (c *failingController) GetControllerInfo() *controller.Info {
	return controller.NewInfo("controllerbus/test/failing", controller.MustParseVersion("0.0.1"), "fails on demand")
}

func (c *failingController) HandleDirective(context.Context, directive.Instance) ([]directive.Resolver, error) {
	return nil, nil
}

func (c *failingController) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.fail:
		return errors.New("controller failed")
	}
}

func (c *failingController) Close() error {
	return nil
}

// failingFactory constructs one failingController.
type failingFactory struct {
	ctrl *failingController
}

func (f *failingFactory) GetConfigID() string {
	return "controllerbus/test/failing"
}

func (f *failingFactory) ConstructConfig() config.Config {
	return &boilerplate_controller.Config{}
}

func (f *failingFactory) Construct(context.Context, config.Config, controller.ConstructOpts) (controller.Controller, error) {
	return f.ctrl, nil
}

func (f *failingFactory) GetVersion() controller.Version {
	return controller.MustParseVersion("0.0.1")
}

// TestWaitExecControllerRunningDisposesOnExit checks that a returned controller
// which exits with an error releases its holder.
func TestWaitExecControllerRunningDisposesOnExit(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	b, _, err := core.NewCoreBus(ctx, logrus.NewEntry(logrus.New()))
	if err != nil {
		t.Fatal(err)
	}

	fail := make(chan struct{})
	disposed := make(chan struct{})
	factory := &failingFactory{ctrl: &failingController{fail: fail}}
	_, _, ref, err := loader.WaitExecControllerRunning(ctx, b, loader.NewExecController(factory, &boilerplate_controller.Config{}), func() {
		close(disposed)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer ref.Release()

	close(fail)
	select {
	case <-disposed:
	case <-ctx.Done():
		t.Fatal("dispose was not called after the controller exited with an error")
	}
}

// flakyFactory fails its first construction and then constructs ctrl.
type flakyFactory struct {
	failingFactory
	constructs atomic.Int32
}

func (f *flakyFactory) Construct(ctx context.Context, conf config.Config, opts controller.ConstructOpts) (controller.Controller, error) {
	if f.constructs.Add(1) == 1 {
		return nil, errors.New("construct failed")
	}
	return f.failingFactory.Construct(ctx, conf, opts)
}

// TestWaitExecControllerRunningRetry checks that the retrying wait outlasts a
// failed start and returns the controller once the loader runs it.
func TestWaitExecControllerRunningRetry(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	b, _, err := core.NewCoreBus(ctx, logrus.NewEntry(logrus.New()))
	if err != nil {
		t.Fatal(err)
	}

	factory := &flakyFactory{failingFactory: failingFactory{ctrl: &failingController{}}}
	ctrl, _, ref, err := loader.WaitExecControllerRunningRetryTyped[*failingController](ctx, b, loader.NewExecController(factory, &boilerplate_controller.Config{}), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ref.Release()
	if ctrl != factory.ctrl {
		t.Fatal("unexpected controller")
	}
	if constructs := factory.constructs.Load(); constructs != 2 {
		t.Fatalf("expected a retried construction, got %d", constructs)
	}
}
