package inmem

import "github.com/aperturerobotics/controllerbus/controller"

// attachedCtrl contains an attached controller
type attachedCtrl struct {
	// ctrl is the controller
	ctrl controller.Controller
	// rel releases the controller
	rel func()
}
