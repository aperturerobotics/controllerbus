package controller_exec

import (
	"errors"
	"strings"
)

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

// FormatLogString formats a log string with info from the response.
func (e *ExecControllerResponse) FormatLogString() string {
	var pts []string
	if id := e.GetId(); len(id) != 0 {
		pts = append(pts, id)
	}
	if status := e.GetStatus(); status != 0 {
		pts = append(pts, status.String())
	}
	if err := e.GetError(); err != nil {
		pts = append(pts, err.Error())
	}
	return strings.Join(pts, ": ")
}
