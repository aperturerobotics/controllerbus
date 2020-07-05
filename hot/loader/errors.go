package hot_loader

import "github.com/pkg/errors"

func errInvalidPluginType(sym interface{}) error {
	return errors.Errorf(
		"could not load hot plugin, invalid type: %#v",
		sym,
	)
}
