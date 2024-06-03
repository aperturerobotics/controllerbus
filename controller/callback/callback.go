package callback

import (
	"context"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// ExecuteFunc is the type for the execute function.
type ExecuteFunc = func(ctx context.Context) error

// CloseFunc is the type for the close function.
type CloseFunc = func() error

// CallbackController wraps a ExecuteFunc and/or HandlerFunc into a controller.
// This is a utility to quickly build controllers.
type CallbackController struct {
	info      *controller.Info
	execute   ExecuteFunc
	handleDir directive.HandlerFunc
	closeFn   CloseFunc
}

// NewCallbackController constructs a new CallbackController.
func NewCallbackController(info *controller.Info, execute ExecuteFunc, handleDir directive.HandlerFunc, closeFn CloseFunc) *CallbackController {
	return &CallbackController{
		info:      info,
		execute:   execute,
		handleDir: handleDir,
		closeFn:   closeFn,
	}
}

// GetControllerID returns the controller ID.
func (c *CallbackController) GetControllerID() string {
	return c.info.GetId()
}

// GetControllerInfo returns information about the controller.
func (c *CallbackController) GetControllerInfo() *controller.Info {
	return c.info.Clone()
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any unexpected errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *CallbackController) HandleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	if c.handleDir != nil {
		return c.handleDir(ctx, di)
	}
	return nil, nil
}

// Execute executes the given controller.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *CallbackController) Execute(ctx context.Context) error {
	if c.execute != nil {
		return c.execute(ctx)
	}
	return nil
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (c *CallbackController) Close() error {
	if c.closeFn != nil {
		return c.closeFn()
	}
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*CallbackController)(nil))
