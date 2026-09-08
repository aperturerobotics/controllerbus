package inmem

import "testing"

// TestDetachControllerIsIdempotent checks public removal and one release notification.
func TestDetachControllerIsIdempotent(t *testing.T) {
	releaseCalls := 0
	target := &attachedCtrl{
		rel: func() { releaseCalls++ },
	}
	retained := &attachedCtrl{}
	b := NewBus(nil)
	b.controllers = []*attachedCtrl{target, retained}

	locked := b.bcast.Lock()
	firstBroadcast := locked.WaitCh()
	locked.Unlock()

	b.detachController(target)

	if got := releaseCalls; got != 1 {
		t.Fatalf("release calls after first detach = %d, want 1", got)
	}
	if got := len(b.GetControllers()); got != 1 {
		t.Fatalf("controller count after first detach = %d, want 1", got)
	}
	if !target.detached || retained.detached {
		t.Fatal("retained controller changed after first detach")
	}

	broadcastCalls := 0
	select {
	case <-firstBroadcast:
		broadcastCalls++
	default:
		t.Fatal("first detach did not broadcast")
	}

	locked = b.bcast.Lock()
	secondBroadcast := locked.WaitCh()
	locked.Unlock()

	b.detachController(target)

	if got := releaseCalls; got != 1 {
		t.Fatalf("release calls after second detach = %d, want 1", got)
	}
	if got := len(b.GetControllers()); got != 1 {
		t.Fatalf("controller count after second detach = %d, want 1", got)
	}
	if !target.detached || retained.detached {
		t.Fatal("retained controller changed after second detach")
	}
	select {
	case <-secondBroadcast:
		broadcastCalls++
	default:
	}
	if broadcastCalls != 1 {
		t.Fatalf("broadcast calls after two detaches = %d, want 1", broadcastCalls)
	}
}
