package controller

import (
	"context"

	"github.com/aperturerobotics/controllerbus/directive"
)

// Controller tracks a particular process.
type Controller interface {
	// Handler handles directives.
	directive.Handler

	// GetControllerInfo returns information about the controller.
	GetControllerInfo() *Info
	// Execute executes the controller goroutine.
	// Returning nil ends execution.
	// Returning an error triggers a retry with backoff.
	// Retry will NOT re-construct the controller, just re-start Execute.
	Execute(ctx context.Context) error
	// Close releases any resources used by the controller.
	// Error indicates any issue encountered releasing.
	Close() error
}
