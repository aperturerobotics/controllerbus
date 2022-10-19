package keyed

import (
	"time"

	"github.com/sirupsen/logrus"
)

// Option is an option for a Keyed instance.
type Option[T comparable] interface {
	// ApplyToKeyed applies the option to the Keyed.
	ApplyToKeyed(k *Keyed[T])
}

type option[T comparable] struct {
	cb func(k *Keyed[T])
}

// newOption constructs a new option.
func newOption[T comparable](cb func(k *Keyed[T])) *option[T] {
	return &option[T]{
		cb: cb,
	}
}

// ApplyToKeyed applies the option to the Keyed instance.
func (o *option[T]) ApplyToKeyed(k *Keyed[T]) {
	if o.cb != nil {
		o.cb(k)
	}
}

// WithReleaseDelay adds a delay after removing a key before canceling the routine.
func WithReleaseDelay[T comparable](delay time.Duration) Option[T] {
	if delay < 0 {
		delay *= -1
	}
	return newOption(func(k *Keyed[T]) {
		k.releaseDelay = delay
	})
}

// WithExitCb adds a callback after a routine exits.
func WithExitCb[T comparable](cb func(key string, routine Routine, data T, err error)) Option[T] {
	return newOption(func(k *Keyed[T]) {
		k.exitedCbs = append(k.exitedCbs, cb)
	})
}

// WithExitLogger adds a exited callback which logs information about the exit.
func WithExitLogger[T comparable](le *logrus.Entry) Option[T] {
	return WithExitCb[T](NewLogExitedCallback[T](le))
}
