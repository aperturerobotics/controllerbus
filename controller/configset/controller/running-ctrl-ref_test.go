package configset_controller

import (
	"testing"

	"github.com/aperturerobotics/controllerbus/controller/configset"
)

// TestRunningControllerRefStateWithoutController checks that a reference
// resolved before its controller starts reports its initial state to a new
// state callback instead of dereferencing the missing controller.
func TestRunningControllerRefStateWithoutController(t *testing.T) {
	ref := newRunningControllerRef("pending", nil)
	var got configset.State
	ref.AddStateCallback(func(st configset.State) { got = st })
	if got == nil || got.GetId() != "pending" {
		t.Fatalf("initial state = %v", got)
	}
}
