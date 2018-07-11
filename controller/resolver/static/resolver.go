package static

import (
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
)

// Resolver r
type Resolver struct {
}

// GetFactoryMatchingConfig returns the factory that matches the config.
// If no factory is found, return nil.
// If an unexpected error occurs, return it.
func (r *Resolver) GetFactoryMatchingConfig(c config.Config) (controller.Factory, error) {
}
