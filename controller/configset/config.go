package configset

import (
	"github.com/aperturerobotics/controllerbus/config"
)

// controllerConfig backs ControllerConfig in memory.
type controllerConfig struct {
	rev  uint64
	conf config.Config
}

// NewControllerConfig constructs a controller config object.
func NewControllerConfig(rev uint64, conf config.Config) ControllerConfig {
	return &controllerConfig{
		rev:  rev,
		conf: conf,
	}
}

// GetRev returns the revision.
func (c *controllerConfig) GetRev() uint64 {
	return c.rev
}

// GetConfig returns the config object.
func (c *controllerConfig) GetConfig() config.Config {
	return c.conf
}

// _ is a type assertion
var _ ControllerConfig = ((*controllerConfig)(nil))
