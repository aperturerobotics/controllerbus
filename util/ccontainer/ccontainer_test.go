package ccontainer

import (
	"context"
	"testing"
	"time"
)

// TestCContainer tests the concurrent container
func TestCContainer(t *testing.T) {
	ctx := context.Background()
	c := NewCContainer(ctx, nil)
	_ = c.WaitValueEmpty(ctx) // should be instant
	nvalCh := make(chan interface{}, 1)
	errCh := make(chan error, 1)
	go func() {
		nv, err := c.WaitValue(ctx)
		if err != nil {
			errCh <- err
			return
		}
		nvalCh <- nv
	}()
	c.SetValue(5)
	gv := <-nvalCh
	if gv != 5 {
		t.Fail()
	}
	dl, dlCancel := context.WithDeadline(ctx, time.Now().Add(time.Millisecond*1))
	defer dlCancel()
	select {
	case err := <-errCh:
		t.Fatal(err.Error())
	case <-dl.Done():
	}
	err := c.WaitValueEmpty(dl)
	if err != context.DeadlineExceeded {
		t.Fail()
	}
	c.SetValue(nil)
	_ = c.WaitValueEmpty(ctx) // should be instant
}
