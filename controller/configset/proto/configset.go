package configset_proto

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/configset"
)

// ConfigSetMap implements the controllerbus.ConfigSet as protobuf.
// Cast the proto type to this type.
type ConfigSetMap map[string]*ControllerConfig

// NewConfigSet constructs a configset from a configset.ConfigSet.
func NewConfigSet(c configset.ConfigSet) (*ConfigSet, error) {
	m, err := NewConfigSetMap(c)
	if err != nil {
		return nil, err
	}

	return &ConfigSet{Configurations: m}, nil
}

// NewConfigSetMap packs a configset to a proto object.
func NewConfigSetMap(c configset.ConfigSet) (ConfigSetMap, error) {
	if c == nil {
		return nil, nil
	}
	m := make(ConfigSetMap)
	var err error
	for k, v := range c {
		m[k], err = NewControllerConfig(v)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

// Resolve resolves the configset into a configset.ConfigSet
func (c *ConfigSet) Resolve(ctx context.Context, b bus.Bus) (configset.ConfigSet, error) {
	return ConfigSetMap(c.GetConfigurations()).Resolve(ctx, b)
}

// Resolve resolves the configset into a configset.ConfigSet
func (c ConfigSetMap) Resolve(ctx context.Context, b bus.Bus) (configset.ConfigSet, error) {
	if c == nil {
		return nil, nil
	}

	oc := configset.ConfigSet{}
	var err error
	for k, v := range c {
		oc[k], err = v.Resolve(ctx, b)
		if err != nil {
			return nil, err
		}
	}
	return oc, nil
}
