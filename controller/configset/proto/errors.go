package configset_proto

import "errors"

var (
	// ErrControllerConfigIdEmpty is returned if the controller config id was empty.
	ErrControllerConfigIdEmpty = errors.New("controller config id empty")
)
