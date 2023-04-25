package configset_json

import (
	"context"
	"errors"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/configset"
)

// ConfigSet implements the JSON unmarshaling logic for a configset.
type ConfigSet map[string]*ControllerConfig

// NewConfigSet constructs a new ConfigSet.
func NewConfigSet(c configset.ConfigSet) ConfigSet {
	if c == nil {
		return nil
	}
	m := make(ConfigSet)
	for k, v := range c {
		m[k] = NewControllerConfig(v)
	}
	return m
}

// Resolve just resolves a configset to a configset.ConfigSet.
func (c ConfigSet) Resolve(ctx context.Context, b bus.Bus) (configset.ConfigSet, error) {
	if c == nil {
		return nil, errors.New("configset was nil")
	}

	m := make(configset.ConfigSet)
	for k, v := range c {
		cc, err := v.Resolve(ctx, b)
		if err != nil {
			return nil, err
		}

		m[k] = configset.NewControllerConfig(cc.GetRev(), cc.GetConfig())
	}

	return m, nil
}

// MarshalJSON marshals a map.
// func (c ConfigSet) MarshalJSON() ([]byte, error) {
// }

// UnmarshalJSON unmarshals a map.
// func (c ConfigSet) UnmarshalJSON(data []byte) error {
// }

// _ is a type assertion
// var _ json.Marshaler = ((ConfigSet)(nil))

// _ is a type assertion
// var _ json.Unmarshaler = ((ConfigSet)(nil))
