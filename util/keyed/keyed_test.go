package keyed

import (
	"context"
	"strconv"
	"testing"
)

// TestKeyed tests the keyed goroutine manager.
func TestKeyed(t *testing.T) {
	ctx := context.Background()
	vals := make(chan string, 10)
	k := NewKeyed(ctx, func(key string) Routine {
		return func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return context.Canceled
			case vals <- key:
				return nil
			}
		}
	})

	nsend := 100
	keys := make([]string, nsend)
	for i := 0; i < nsend; i++ {
		key := "routine-" + strconv.Itoa(i)
		keys[i] = key
	}
	k.SyncKeys(keys, false)

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
