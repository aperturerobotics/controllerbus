package refcount

import (
	"context"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/util/ccontainer"
)

// TestRefCount tests the RefCount mechanism.
func TestRefCount(t *testing.T) {
	ctx := context.Background()
	target := ccontainer.NewCContainer[*string](nil)
	targetErr := ccontainer.NewCContainer[*error](nil)
	var valCalled, relCalled bool
	rc := NewRefCount(nil, target, targetErr, func(ctx context.Context) (*string, func(), error) {
		val := "hello world"
		valCalled = true
		return &val, func() {
			relCalled = true
		}, nil
	})

	ref := rc.AddRef(nil)
	<-time.After(time.Millisecond * 50)
	if valCalled || relCalled {
		t.Fail()
	}

	rc.SetContext(ctx)
	<-time.After(time.Millisecond * 50)
	if !valCalled || relCalled {
		t.Fail()
	}

	var gotValue *string
	var gotErr error
	gotErr = AccessRefCount(ctx, rc, func(val *string) error {
		gotValue = val
		return nil
	})

	waitVal, err := target.WaitValue(ctx, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	if waitVal != gotValue || gotErr != nil || relCalled {
		t.Fail()
	}
	ref.Release()

	if !relCalled {
		t.Fail()
	}
}
