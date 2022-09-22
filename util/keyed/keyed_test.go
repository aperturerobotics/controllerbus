package keyed

import (
	"context"
	"strconv"
	"testing"
	"time"
)

// testData contains some test metadata.
type testData struct{}

// TestKeyed tests the keyed goroutine manager.
func TestKeyed(t *testing.T) {
	ctx := context.Background()
	vals := make(chan string, 10)
	k := NewKeyed(func(key string) (Routine, *testData) {
		return func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return context.Canceled
			case vals <- key:
				return nil
			}
		}, &testData{}
	})

	nsend := 100
	keys := make([]string, nsend)
	for i := 0; i < nsend; i++ {
		key := "routine-" + strconv.Itoa(i)
		keys[i] = key
	}
	k.SyncKeys(keys, false)

	// expect nothing to have been pushed to vals yet
	<-time.After(time.Millisecond * 10)
	select {
	case val := <-vals:
		t.Fatalf("unexpected value before set context: %s", val)
	default:
	}

	// start execution
	k.SetContext(ctx, false)

	seen := make(map[string]struct{})
	for {
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err().Error())
		case val := <-vals:
			if _, ok := seen[val]; ok {
				t.Fatalf("duplicate value: %s", val)
			}
			seen[val] = struct{}{}
			if len(seen) == nsend {
				// success
				return
			}
		}
	}
}
