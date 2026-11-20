// Package controller declares the Controller, a long-lived process attached
// to a bus, and the Factory that constructs one from its configuration.
//
// A Controller is a directive Handler with a lifetime. The bus calls Execute
// once per attachment and offers the controller every directive while it is
// attached. An instance that fails is closed and never run again, so a
// controller that must recover is run through the loader, whose ExecController
// directive builds a fresh instance from its Factory after backoff and counts
// only the failures since the controller last ran stably.
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
	// Execute runs the controller's background work until ctx is canceled.
	// Returning nil ends that work; the controller stays attached and keeps
	// handling directives until it is removed.
	// Returning an error or panicking detaches this instance and closes it.
	// The bus never restarts it; the loader's ExecController directive
	// constructs a new instance after backoff.
	Execute(ctx context.Context) error
	// Close releases any resources used by the controller.
	// Error indicates any issue encountered releasing.
	Close() error
}
