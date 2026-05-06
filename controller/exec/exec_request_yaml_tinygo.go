//go:build tinygo

package controller_exec

import (
	"context"
	"errors"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/configset"
)

// ExecControllerYAMLFromConfigSet converts a config set to a ExecControllerRequest w/ YAML.
func ExecControllerYAMLFromConfigSet(cs configset.ConfigSet) (*ExecControllerRequest, error) {
	return nil, errors.New("yaml controller config sets are unsupported in tinygo")
}

func resolveYAMLConfigSet(
	ctx context.Context,
	cbus bus.Bus,
	confsYAML []byte,
	confSet configset.ConfigSet,
	overwrite bool,
) error {
	return errors.New("yaml controller config sets are unsupported in tinygo")
}
