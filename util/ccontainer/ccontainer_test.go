package ccontainer

import (
	"context"
	"testing"
	"time"
)

// TestCContainer tests the concurrent container
func TestCContainer(t *testing.T) {
	ctx := context.Background()
	c := NewCContainer[*int](nil)

	errCh := make(chan error, 1)
	_ = c.WaitValueEmpty(ctx, errCh) // should be instant

	var val = 5
	go c.SetValue(&val)
	gv, err := c.WaitValue(ctx, errCh)
	if err != nil {
		t.Fatal(err.Error())
	}
	if gv == nil || *gv != 5 {
		t.Fail()
	}

	dl, dlCancel := context.WithDeadline(ctx, time.Now().Add(time.Millisecond*1))
	defer dlCancel()
	err = c.WaitValueEmpty(dl, errCh)
	if err != context.DeadlineExceeded {
		t.Fail()
	}

	c.SetValue(nil)
	_ = c.WaitValueEmpty(ctx, errCh) // should be instant
}
