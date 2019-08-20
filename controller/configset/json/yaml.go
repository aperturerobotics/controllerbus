package configset_json

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

// MarshalYAML marshals a config set to yaml.
func MarshalYAML(cs configset.ConfigSet) ([]byte, error) {
	ycs := NewConfigSet(cs)
	return yaml.Marshal(ycs)
}

// UnmarshalYAML unmarshals a yaml to a config set, optionally overwriting existing.
func UnmarshalYAML(
	ctx context.Context,
	b bus.Bus,
	data []byte,
	ocs configset.ConfigSet,
	overwriteExisting bool,
) error {
	cs := make(ConfigSet)
	if err := yaml.Unmarshal(data, &cs); err != nil {
		return err
	}

	for id, cconf := range cs {
		_, exists := ocs[id]
		if exists && !overwriteExisting {
			continue
		}
		cfg, err := cconf.Resolve(ctx, b)
		if err != nil {
			return errors.Wrapf(err, "unmarshal config %q", id)
		}
		ocs[id] = cfg
	}

	return nil
}
