package keyed

import (
	"context"

	"github.com/sirupsen/logrus"
)

// NewLogExitedCallback returns a ExitedCb which logs when a controller exited.
func NewLogExitedCallback[T comparable](le *logrus.Entry) func(key string, routine Routine, data T, err error) {
	return func(key string, routine Routine, data T, err error) {
		if err != nil && err != context.Canceled {
			le.WithError(err).Warnf("controller exited: %s", key)
		} else {
			le.Debugf("controller exited: %s", key)
		}
	}
}
