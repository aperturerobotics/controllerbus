//go:build !tinygo

package controller_exec

import (
	"context"
	"sort"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	configset_json "github.com/aperturerobotics/controllerbus/controller/configset/json"
)

// ExecControllerYAMLFromConfigSet converts a config set to a ExecControllerRequest w/ YAML.
func ExecControllerYAMLFromConfigSet(cs configset.ConfigSet) (*ExecControllerRequest, error) {
	// encode to yaml
	dat, err := configset_json.MarshalYAML(cs)
	if err != nil {
		return nil, err
	}
	return &ExecControllerRequest{
		ConfigSetYaml: string(dat),
	}, nil
}

func resolveYAMLConfigSet(
	ctx context.Context,
	cbus bus.Bus,
	confsYAML []byte,
	confSet configset.ConfigSet,
	overwrite bool,
) error {
	addedConfs, err := configset_json.UnmarshalYAML(ctx, cbus, confsYAML, confSet, overwrite)
	if err != nil {
		return err
	}
	sort.Strings(addedConfs)
	return nil
}
