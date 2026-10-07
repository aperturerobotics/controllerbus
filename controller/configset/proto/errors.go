package configset_proto

import "errors"

// ErrControllerConfigIdEmpty is returned if the controller config id was empty.
var ErrControllerConfigIdEmpty = errors.New("controller config id empty")

// ErrUnknownConfigID is returned if no factory registers the config id.
var ErrUnknownConfigID = errors.New("no factory registers the controller config id")
