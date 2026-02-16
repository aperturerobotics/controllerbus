package plugin_shared_library

import "github.com/pkg/errors"

func errInvalidPluginType(sym any) error {
	return errors.Errorf(
		"could not load plugin, invalid type: %#v",
		sym,
	)
}
