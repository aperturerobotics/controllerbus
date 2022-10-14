package plugin_loader_filesystem

import (
	"context"
	"os"
	"path"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Version is the version of the controller implementation.
var Version = semver.MustParse("0.0.1")

// ControllerID is the ID of the controller.
const ControllerID = "controllerbus/plugin/loader/shared-library/filesystem/1"

// Controller is the hot plugin filesystem loading controller.
//
// NOTE: Go plugins cannot be unloaded (Go 1.16). This is only a prototype.
type Controller struct {
	// le is the root logger
	le *logrus.Entry
	// bus is the controller bus
	bus bus.Bus
	// dir is the directory to watch
	dir string
	// watch indicates to watch the filesystem
	watch bool
}

// NewController constructs a new controller.
func NewController(le *logrus.Entry, bus bus.Bus, conf *Config) (*Controller, error) {
	dir := path.Clean(conf.GetDir())
	if _, err := os.Stat(dir); err != nil {
		return nil, errors.Wrapf(err, "stat %s", dir)
	}
	return &Controller{
		le:  le,
		bus: bus,
		dir: dir,

		watch: conf.GetWatch(),
	}, nil
}

// Execute executes the given controller.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *Controller) Execute(ctx context.Context) error {
	w := NewWatcher(c.le, c.bus)
	return w.Execute(ctx, c.dir, c.watch)
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *Controller) HandleDirective(
	ctx context.Context,
	di directive.Instance,
) ([]directive.Resolver, error) {
	return nil, nil
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() *controller.Info {
	return controller.NewInfo(
		ControllerID,
		Version,
		"plugin shared filesystem loader: "+c.dir,
	)
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (c *Controller) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*Controller)(nil))
