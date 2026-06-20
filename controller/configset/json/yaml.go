package configset_json

import (
	"context"
	"encoding/json"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	cbyaml "github.com/aperturerobotics/controllerbus/yaml"
	"github.com/pkg/errors"
)

// MarshalYAML marshals a config set to yaml.
func MarshalYAML(cs configset.ConfigSet) ([]byte, error) {
	ycs := NewConfigSet(cs)
	jdat, err := json.Marshal(ycs)
	if err != nil {
		return nil, err
	}
	return cbyaml.JSONToYAML(jdat)
}

// UnmarshalYAML unmarshals a yaml to a config set, optionally overwriting existing.
//
// Returns all added configs.
func UnmarshalYAML(
	ctx context.Context,
	b bus.Bus,
	data []byte,
	ocs configset.ConfigSet,
	overwriteExisting bool,
) ([]string, error) {
	cs := make(ConfigSet)
	jdat, err := cbyaml.YAMLToJSON(data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jdat, &cs); err != nil {
		return nil, err
	}

	var added []string
	for id, cconf := range cs {
		_, exists := ocs[id]
		if exists && !overwriteExisting {
			continue
		}
		cfg, err := cconf.Resolve(ctx, b)
		if err != nil {
			return added, errors.Wrapf(err, "unmarshal config %q", id)
		}
		ocs[id] = cfg
		added = append(added, id)
	}

	return added, nil
}
