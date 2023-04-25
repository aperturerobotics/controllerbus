package configset_json

import (
	"context"
	"errors"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	"github.com/ghodss/yaml"
)

// ControllerConfig implements the JSON unmarshaling logic for a configset
// ControllerConfig.
type ControllerConfig struct {
	// Rev is the revision number.
	Rev uint64 `json:"rev,omitempty"`
	// Id is the configuration ID.
	Id string `json:"id"`
	// Config is the configuration object.
	Config *Config `json:"config"`
}

// NewControllerConfig builds a new controller config.
func NewControllerConfig(c configset.ControllerConfig) *ControllerConfig {
	return &ControllerConfig{
		Rev: c.GetRev(),
		Id:  c.GetConfig().GetConfigID(),
		Config: &Config{
			underlying: c.GetConfig(),
		},
	}
}

// UnmarshalControllerConfigYAML unmarshals a yaml to a ControllerConfig.
func UnmarshalControllerConfigYAML(data []byte) (*ControllerConfig, error) {
	conf := &ControllerConfig{}
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}
	return conf, nil
}

// GetRev returns the revision.
func (c *ControllerConfig) GetRev() uint64 {
	return c.Rev
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

	return configset.NewControllerConfig(c.Rev, c.Config.GetConfig()), nil
}
