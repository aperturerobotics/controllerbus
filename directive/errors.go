package directive

import "errors"

// ErrDirectiveDisposed is returned when the directive was unexpectedly disposed.
var ErrDirectiveDisposed = errors.New("directive disposed unexpectedly")
