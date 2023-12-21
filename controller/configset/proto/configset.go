package configset_proto

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	"github.com/pkg/errors"
	jsonpb "google.golang.org/protobuf/encoding/protojson"
)

// ConfigSetMap implements the controllerbus.ConfigSet as protobuf.
// Cast the proto type to this type.
type ConfigSetMap map[string]*ControllerConfig

// NewConfigSet constructs a configset from a configset.ConfigSet.
func NewConfigSet(c configset.ConfigSet, useJson bool) (*ConfigSet, error) {
	m, err := NewConfigSetMap(c, useJson)
	if err != nil {
		return nil, err
	}

	return &ConfigSet{Configurations: m}, nil
}

// NewConfigSetMap packs a configset to a proto object.
func NewConfigSetMap(c configset.ConfigSet, useJson bool) (ConfigSetMap, error) {
	if c == nil {
		return nil, nil
	}
	m := make(ConfigSetMap)
	var err error
	for k, v := range c {
		m[k], err = NewControllerConfig(v, useJson)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

// MergeConfigSets merges multiple config sets maps to one ConfigSet.
func MergeConfigSets(sets ...*ConfigSet) *ConfigSet {
	out := &ConfigSet{Configurations: make(ConfigSetMap)}
	maps := make([]ConfigSetMap, 0, len(sets))
	for _, set := range sets {
		if csm := set.GetConfigurations(); len(csm) != 0 {
			maps = append(maps, csm)
		}
	}
	MergeConfigSetMaps(out.Configurations, maps...)
	return out
}

// MergeConfigSetMaps merges multiple config set maps to one ConfigSetMap.
func MergeConfigSetMaps(out ConfigSetMap, sets ...ConfigSetMap) {
	if out == nil {
		return
	}
	for _, set := range sets {
		for k, v := range set {
			if v == nil {
				continue
			}
			vRev := v.GetRev()
			existing, existingOk := out[k]
			if existingOk && existing.GetRev() > vRev {
				continue
			}
			out[k] = v
		}
	}
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

// Validate validates the ConfigSet.
func (c *ConfigSet) Validate() error {
	if err := ConfigSetMap(c.GetConfigurations()).Validate(); err != nil {
		return err
	}
	return nil
}

// Validate validates the ConfigSetMap.
func (c ConfigSetMap) Validate() error {
	for configID, configObj := range c {
		if err := configObj.Validate(); err != nil {
			return errors.Wrapf(err, "configurations[%s]", configID)
		}
	}
	return nil
}

// ApplyConfig applies a config to the ConfigSetMap.
func (c ConfigSetMap) ApplyConfig(id string, conf config.Config, rev uint64, useJson bool) (*ControllerConfig, error) {
	ctrlConf, err := NewControllerConfig(configset.NewControllerConfig(rev, conf), useJson)
	if err != nil {
		return nil, err
	}

	c[id] = ctrlConf
	return ctrlConf, nil
}
