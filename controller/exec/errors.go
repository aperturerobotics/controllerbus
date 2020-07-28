package controller_exec

import "errors"

// ErrAllControllersFailed is returned if all controller configs failed to run.
var ErrAllControllersFailed = errors.New("all controller configurations failed")
