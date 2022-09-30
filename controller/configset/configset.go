package configset

import (
	"context"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
)

// Controller is a configset controller.
type Controller interface {
	// Controller indicates this is a controllerbus controller.
	controller.Controller

	// PushControllerConfig pushes a controller configuration for a given key, if
	// the version is newer. Returns a reference to the configset, or an error.
	//
	// Ctx is used for the context of the function call only.
	PushControllerConfig(
		ctx context.Context,
		key string,
		conf ControllerConfig,
	) (Reference, error)
}

// ConfigSet is a key/value set of controller configs.
type ConfigSet map[string]ControllerConfig

// Equal checks if the configset is equal to the other.
func (c ConfigSet) Equal(os ConfigSet) bool {
	if os == nil || c == nil {
		return false
	}
	for k, v := range c {
		ov, ok := os[k]
		if !ok {
			return false
		}
		if !ov.GetConfig().EqualsConfig(v.GetConfig()) {
			return false
		}
		if ov.GetRevision() != v.GetRevision() {
			return false
		}
	}

	for k := range os {
		if _, ok := c[k]; !ok {
			return false
		}
	}

	return true
}

// MergeConfigSets merges multiple config sets to one ConfigSet.
func MergeConfigSets(sets ...ConfigSet) ConfigSet {
	out := make(ConfigSet)
	for _, set := range sets {
		for k, v := range set {
			if v == nil {
				continue
			}
			vRev := v.GetRevision()
			existing, existingOk := out[k]
			if existingOk && existing.GetRevision() > vRev {
				continue
			}
			out[k] = v
		}
	}
	return out
}

// ControllerConfig is a wrapped controller configuration.
type ControllerConfig interface {
	// GetRevision returns the revision.
	GetRevision() uint64
	// GetConfig returns the config object.
	GetConfig() config.Config
}

// Reference is a reference to a pushed controller config. The reference is used
// to monitor the state of the controller and release the reference when no
// longer needed, so that the configured controllers can be properly terminated
// when no longer needed.
type Reference interface {
	// GetConfigKey returns the configset key for this controller.
	GetConfigKey() string
	// GetState returns the current state object.
	GetState() State
	// AddStateCb adds a callback that is called when the state changes.
	// Should not block.
	// Will be called with the initial state.
	AddStateCb(func(State))
	// Release releases the reference.
	Release()
}

// State contains controller state.
type State interface {
	// GetId returns the controller id.
	GetId() string
	// GetControllerConfig returns the current controller config in use.
	GetControllerConfig() ControllerConfig
	// GetController returns the controller instance if running.
	// Returns nil otherwise.
	GetController() controller.Controller
	// GetError returns any error processing the controller config.
	GetError() error
}
