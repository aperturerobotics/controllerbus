package promise

import (
	"context"
	"testing"
	"time"
)

// doSomeWork returns a promise.
func doSomeWork() *Promise[int] {
	result := NewPromise[int]()
	go func() {
		<-time.After(time.Millisecond * 50)
		result.SetResult(10, nil)
	}()
	return result
}

// TestPromise tests the Promise mechanics.
func TestPromise(t *testing.T) {
	result := doSomeWork()
	val, err := result.Await(context.Background())
	if err != nil {
		t.Fatal(err.Error())
	}
	if val != 10 {
		t.Fail()
	}
}
