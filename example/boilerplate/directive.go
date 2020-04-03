package boilerplate

import (
	"github.com/aperturerobotics/controllerbus/directive"
)

// Boilerplate prints a boilerplate with a message, as a example.
type Boilerplate interface {
	// Networked indicates PrintBoilerplate is a networked directive.
	directive.Networked

	// BoilerplateMessage returns the message to print with the boilerplate.
	BoilerplateMessage() string
}

// BoilerplateResult is the result type for Boilerplate.
type BoilerplateResult interface {
	// Value indicates this is a directive value
	directive.Value
	// GetPrintedLen returns the number of characters printed.
	// If zero then the message was not printed.
	GetPrintedLen() uint32
}
