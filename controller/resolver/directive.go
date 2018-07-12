package resolver

import (
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/directive"
)

// LoadControllerWithConfig is a directive indicating a controller should be
// loaded given a configuration.
type LoadControllerWithConfig interface {
	// Directive indicates this is a directive.
	directive.Directive

	// GetDesiredControllerConfig returns the desired controller config.
	GetDesiredControllerConfig() config.Config
}
