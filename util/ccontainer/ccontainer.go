package ccontainer

import (
	"context"
	"sync"
)

// CContainer is a concurrent container.
type CContainer struct {
	mtx  sync.Mutex
	val  interface{}
	wake chan struct{}
}

// NewCContainer builds a CContainer with an initial value.
func NewCContainer(val interface{}) *CContainer {
	return &CContainer{val: val, wake: make(chan struct{})}
}

// WaitValueWithValidator waits for any value that matches the validator in the container.
// errCh is an optional channel to read an error from.
func (c *CContainer) WaitValueWithValidator(
	ctx context.Context,
	valid func(v interface{}) (bool, error),
	errCh <-chan error,
) (interface{}, error) {
	var ok bool
	var err error
	for {
		c.mtx.Lock()
		val := c.val
		c.mtx.Unlock()
		if valid != nil {
			ok, err = valid(val)
		} else {
			ok = val != nil
			err = nil
		}
		if err != nil {
			return nil, err
		}
		if ok {
			return val, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-errCh:
			if err != nil {
				return nil, err
			}
		case c.wake <- struct{}{}:
			// woken, value changed
		}
	}
}

// WaitValue waits for any non-nil value in the container.
// errCh is an optional channel to read an error from.
func (c *CContainer) WaitValue(ctx context.Context, errCh <-chan error) (interface{}, error) {
	return c.WaitValueWithValidator(ctx, func(v interface{}) (bool, error) {
		// untyped nil == no value or type
		// no checking for typed nil (typed nil != nil)
		return v != nil, nil
	}, errCh)
}

// WaitValueChange waits for a value that is different than the given.
// errCh is an optional channel to read an error from.
func (c *CContainer) WaitValueChange(ctx context.Context, old interface{}, errCh <-chan error) (interface{}, error) {
	return c.WaitValueWithValidator(ctx, func(v interface{}) (bool, error) {
		return v != old, nil
	}, errCh)
}

// WaitValueEmpty waits for a untyped nil value.
// errCh is an optional channel to read an error from.
func (c *CContainer) WaitValueEmpty(ctx context.Context, errCh <-chan error) error {
	_, err := c.WaitValueWithValidator(ctx, func(v interface{}) (bool, error) {
		return v == nil, nil
	}, errCh)
	return err
}

// SetValue sets the ccontainer value.
//
// Be sure to check for nil when setting if necessary: untyped nil is still
// considered a set value.
func (c *CContainer) SetValue(val interface{}) {
	c.mtx.Lock()
	c.val = val
	c.wakeWaiting()
	c.mtx.Unlock()
}

// wakeWaiting wakes any waiting goroutines
func (c *CContainer) wakeWaiting() {
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
