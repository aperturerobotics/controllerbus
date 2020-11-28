package ccontainer

import (
	"context"
	"testing"
	"time"
)

// TestCContainer tests the concurrent container
func TestCContainer(t *testing.T) {
	ctx := context.Background()
	c := NewCContainer(nil, nil)
	_ = c.WaitValueEmpty(ctx) // should be instant
	nvalCh := make(chan interface{}, 1)
	go func() {
		nv, err := c.WaitValue(ctx)
		if err != nil {
			t.Fatal(err.Error())
		}
		nvalCh <- nv
	}()
	c.SetValue(5)
	gv := <-nvalCh
	if gv != 5 {
		t.Fail()
	}
	dl, _ := context.WithDeadline(ctx, time.Now().Add(time.Millisecond*1))
	<-dl.Done()
	err := c.WaitValueEmpty(dl)
	if err != context.DeadlineExceeded {
		t.Fail()
	}
	c.SetValue(nil)
	_ = c.WaitValueEmpty(ctx) // should be instant
}
