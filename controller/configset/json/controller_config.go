package configset_json

import (
	"context"
	"errors"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/configset"
)

// ControllerConfig implements the JSON unmarshaling logic for a configset
// ControllerConfig.
type ControllerConfig struct {
	// Revision is the revision number.
	Revision uint64 `json:"revision,omitempty"`
	// Id is the configuration ID.
	Id string `json:"id"`
	// Config is the configuration object.
	Config *Config `json:"config"`
}

// NewControllerConfig builds a new controller config.
func NewControllerConfig(c configset.ControllerConfig) *ControllerConfig {
	return &ControllerConfig{
		Revision: c.GetRevision(),
		Id:       c.GetConfig().GetConfigID(),
		Config: &Config{
			underlying: c.GetConfig(),
		},
	}
}

// GetRevision returns the revision.
func (c *ControllerConfig) GetRevision() uint64 {
	return c.Revision
}

// Resolve resolves the config into a configset.ControllerConfig
func (c *ControllerConfig) Resolve(ctx context.Context, b bus.Bus) (configset.ControllerConfig, error) {
	if c == nil {
		return nil, nil
	}
	if c.Config == nil {
		return nil, errors.New("config was not specified")
	}
	if c.Id == "" {
		return nil, errors.New("config id was not specified")
	}

	if err := c.Config.Resolve(ctx, c.Id, b); err != nil {
		return nil, err
	}

	if c.Config.GetConfig() == nil {
		return nil, errors.New("config parsed to null")
	}

	return configset.NewControllerConfig(c.Revision, c.Config.GetConfig()), nil
}
