package hot_compile_demo_controller

import (
	"context"
	"os"
	"time"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"

	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"

	// The following import should be correctly transformed.
	_ "github.com/aperturerobotics/controllerbus/plugin"
)

// Version is the version of the controller implementation.
var Version = semver.MustParse("0.0.1")

// ControllerID is the ID of the controller.
const ControllerID = "controllerbus/example/hot-demo/demo-controller"

// Controller implements the demo controller.
type Controller struct {
	// le is the log entry
	le *logrus.Entry
	// bus is the controller bus
	bus bus.Bus
	// exitAfter exits after a period of time
	exitAfter time.Duration
}

// NewController constructs a new entity graph controller.
func NewController(
	le *logrus.Entry,
	bus bus.Bus,
	conf *Config,
) *Controller {
	exitAfter, _ := conf.ParseExitAfterDur()
	return &Controller{
		le:        le,
		bus:       bus,
		exitAfter: exitAfter,
	}
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() *controller.Info {
	return controller.NewInfo(
		ControllerID,
		Version,
		"demo controller",
	)
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any unexpected errors are returned for logging.
// It is safe to add a reference to the directive during this call.
// The context passed is canceled when the directive instance expires.
func (c *Controller) HandleDirective(
	ctx context.Context,
	inst directive.Instance,
) ([]directive.Resolver, error) {
	return nil, nil
}

// Execute executes the controller goroutine.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *Controller) Execute(ctx context.Context) error {
	c.le.Info("hello from hot compile demo controller")

	if c.exitAfter != 0 {
		c.le.Infof("exiting after %s as configured", c.exitAfter.String())
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.exitAfter):
			os.Exit(0)
		}
	}

	return nil
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (c *Controller) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*Controller)(nil))
