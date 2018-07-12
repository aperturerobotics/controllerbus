package bus

import (
	"github.com/aperturerobotics/controllerbus/controller"
)

// Bus manages running controllers. It has an attached directive controller,
// which is used to build declarative state requests between controllers.
type Bus interface {
	// ExecuteController adds a controller to the bus and calls Execute().
	// Any fatal error in the controller is returned.
	// The controller will receive directive callbacks.
	ExecuteController(controller.Controller) error
}
