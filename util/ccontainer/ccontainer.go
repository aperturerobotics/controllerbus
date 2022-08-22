package ccontainer

import (
	"context"
	"sync"
)

// CContainer is a concurrent container.
type CContainer[T comparable] struct {
	mtx  sync.Mutex
	val  T
	wake chan struct{}
}

// NewCContainer builds a CContainer with an initial value.
func NewCContainer[T comparable](val T) *CContainer[T] {
	return &CContainer[T]{val: val, wake: make(chan struct{})}
}

// WaitValueWithValidator waits for any value that matches the validator in the container.
// errCh is an optional channel to read an error from.
func (c *CContainer[T]) WaitValueWithValidator(
	ctx context.Context,
	valid func(v T) (bool, error),
	errCh <-chan error,
) (T, error) {
	var ok bool
	var err error
	var emptyValue T
	for {
		c.mtx.Lock()
		val := c.val
		c.mtx.Unlock()
		if valid != nil {
			ok, err = valid(val)
		} else {
			ok = val != emptyValue
			err = nil
		}
		if err != nil {
			return emptyValue, err
		}
		if ok {
			return val, nil
		}

		select {
		case <-ctx.Done():
			return emptyValue, ctx.Err()
		case err := <-errCh:
			if err != nil {
				return emptyValue, err
			}
		case c.wake <- struct{}{}:
			// woken, value changed
		}
	}
}

// WaitValue waits for any non-nil value in the container.
// errCh is an optional channel to read an error from.
func (c *CContainer[T]) WaitValue(ctx context.Context, errCh <-chan error) (T, error) {
	return c.WaitValueWithValidator(ctx, func(v T) (bool, error) {
		var emptyValue T
		return v != emptyValue, nil
	}, errCh)
}

// WaitValueChange waits for a value that is different than the given.
// errCh is an optional channel to read an error from.
func (c *CContainer[T]) WaitValueChange(ctx context.Context, old T, errCh <-chan error) (T, error) {
	return c.WaitValueWithValidator(ctx, func(v T) (bool, error) {
		return v != old, nil
	}, errCh)
}

// WaitValueEmpty waits for a untyped nil value.
// errCh is an optional channel to read an error from.
func (c *CContainer[T]) WaitValueEmpty(ctx context.Context, errCh <-chan error) error {
	_, err := c.WaitValueWithValidator(ctx, func(v T) (bool, error) {
		var emptyValue T
		return v == emptyValue, nil
	}, errCh)
	return err
}

// SetValue sets the ccontainer value.
//
// Be sure to check for nil when setting if necessary: untyped nil is still
// considered a set value.
func (c *CContainer[T]) SetValue(val T) {
	c.mtx.Lock()
	c.val = val
	c.wakeWaiting()
	c.mtx.Unlock()
}

// wakeWaiting wakes any waiting goroutines
func (c *CContainer[T]) wakeWaiting() {
	for {
		select {
		case <-c.wake:
			// woke one routine
		default:
			// none left
			return
		}
	}
}
