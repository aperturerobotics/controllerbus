package controller_exec

import "errors"

// GetError returns an error if the response indicated one, or nil for success.
//
// If no error info was provided, assumes ErrAllControllersFailed
func (e *ExecControllerResponse) GetError() error {
	if e.GetStatus() != ControllerStatus_ControllerStatus_ERROR {
		return nil
	}

	errStr := e.GetErrorInfo()
	if len(errStr) != 0 {
		return errors.New(errStr)
	}

	return ErrAllControllersFailed
}
