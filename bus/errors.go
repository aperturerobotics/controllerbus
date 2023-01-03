package bus

import "errors"

var (
	// ErrDirectiveDisposed is returned when the directive was unexpectedly disposed.
	ErrDirectiveDisposed = errors.New("directive disposed unexpectedly")
)
