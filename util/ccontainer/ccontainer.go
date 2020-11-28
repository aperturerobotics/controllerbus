package ccontainer

import (
	"context"
	"sync"
)

// CContainer is a concurrent container.
type CContainer struct {
	ctx  context.Context // may be nil
	mtx  sync.Mutex
	val  interface{}
	wake chan struct{}
}

// NewCContainer builds a CContainer with a context.
// Note: context can be nil
func NewCContainer(ctx context.Context, val interface{}) *CContainer {
	if ctx == nil {
		ctx = context.Background()
	}
	return &CContainer{ctx: ctx, val: val, wake: make(chan struct{})}
}

// WaitValueWithValidator waits for any value that matches the validator in the container.
func (c *CContainer) WaitValueWithValidator(
	ctx context.Context,
	valid func(v interface{}) (bool, error),
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
		case <-c.ctx.Done():
			return nil, context.Canceled
		case c.wake <- struct{}{}:
			// woken, value changed
		}
	}
}

// WaitValue waits for any non-nil value in the container.
func (c *CContainer) WaitValue(ctx context.Context) (interface{}, error) {
	return c.WaitValueWithValidator(ctx, func(v interface{}) (bool, error) {
		// untyped nil == no value or type
		// no checking for typed nil (typed nil != nil)
		return v != nil, nil
	})
}

// WaitValueChange waits for a value that is different than the given.
func (c *CContainer) WaitValueChange(ctx context.Context, old interface{}) (interface{}, error) {
	return c.WaitValueWithValidator(ctx, func(v interface{}) (bool, error) {
		return v != old, nil
	})
}

// WaitValueEmpty waits for a untyped nil value.
func (c *CContainer) WaitValueEmpty(ctx context.Context) error {
	_, err := c.WaitValueWithValidator(ctx, func(v interface{}) (bool, error) {
		return v == nil, nil
	})
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
