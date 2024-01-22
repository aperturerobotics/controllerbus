package boilerplate_v1

import (
	"errors"
	"time"

	"github.com/aperturerobotics/controllerbus/directive"
	directive_proto "github.com/aperturerobotics/controllerbus/directive/proto"
	"github.com/aperturerobotics/controllerbus/example/boilerplate"
)

// BoilerplateMessage returns the message to print.
func (b *Boilerplate) BoilerplateMessage() string {
	return b.GetMessageText()
}

// GetNetworkedCodec returns the encoder / decoder for this directive.
func (b *Boilerplate) GetNetworkedCodec() directive.NetworkedCodec {
	return directive_proto.GetProtobufCodec()
}

// GetName returns the directive's type name.
// This is not necessarily unique, and is primarily intended for display.
func (b *Boilerplate) GetName() string {
	return "Boilerplate"
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (b *Boilerplate) Validate() error {
	if b.GetMessageText() == "" {
		return errors.New("message text must be specified")
	}
	return nil
}

// GetValueOptions returns options relating to value handling.
func (b *Boilerplate) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{
		// MaxValueCount indicates a maximum number of values to retrieve.
		// The resolvers will be canceled when this many values are gathered.
		// If zero, accepts infinite values.
		MaxValueCount: 1,

		// MaxValueHardCap indicates MaxValueCount is a hard cap. If it is not a
		// hard cap, any values found after resolvers are canceled is accepted. If
		// it is a hard cap, any values found after resolvers are canceled will be
		// rejected.
		MaxValueHardCap: true,

		// UnrefDisposeDur is the duration to wait to dispose a directive after all
		// references have been released.
		UnrefDisposeDur: time.Millisecond * 10,
	}
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (b *Boilerplate) IsEquivalent(other directive.Directive) bool {
	ot, otOk := other.(boilerplate.Boilerplate)
	if !otOk {
		return false
	}

	return ot.BoilerplateMessage() == b.BoilerplateMessage()
}

// GetDebugVals returns the directive arguments as key/value pairs.
// This should be something like param1="test", param2="test".
// This is not necessarily unique, and is primarily intended for display.
func (b *Boilerplate) GetDebugVals() directive.DebugValues {
	return directive.DebugValues{
		"message": []string{b.BoilerplateMessage()},
	}
}

var (
	_ boilerplate.Boilerplate      = ((*Boilerplate)(nil))
	_ directive.Debuggable         = ((*Boilerplate)(nil))
	_ directive.DirectiveWithEquiv = ((*Boilerplate)(nil))
)
