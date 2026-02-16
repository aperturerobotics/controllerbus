package configset_json

import (
	"context"
	"encoding/json"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/pkg/errors"
	// "github.com/aperturerobotics/controllerbus/controller/configset"
)

// Config implements JSON unmarshaling and marshaling logic.
type Config struct {
	pendingParseData string
	underlying       config.Config
}

// NewConfig constructs a new Config for JSON marshaling.
func NewConfig(c config.Config) *Config {
	return &Config{underlying: c}
}

// NewConfigWithJSON returns a Config object for JSON parsing.
func NewConfigWithJSON(data string) *Config {
	return &Config{pendingParseData: data}
}

// Resolve constructs the underlying config from the pending parse data.
func (c *Config) Resolve(ctx context.Context, configID string, b bus.Bus) error {
	if c == nil {
		return errors.New("cannot resolve nil config")
	}

	configCtorDir := resolver.NewLoadConfigConstructorByID(configID)
	configCtorVal, _, configCtorRef, err := bus.ExecOneOff(ctx, b, configCtorDir, nil, nil)
	if err != nil {
		return errors.WithMessage(err, "resolve config object")
	}
	defer configCtorRef.Release()

	ctor, ctorOk := configCtorVal.GetValue().(config.Constructor)
	if !ctorOk {
		return errors.New("load config constructor directive returned invalid object")
	}
	c.underlying = ctor.ConstructConfig()
	if c.underlying == nil {
		return errors.New("load config constructor returned nil object")
	}
	if err := c.underlying.UnmarshalJSON([]byte(c.pendingParseData)); err != nil {
		return err
	}
	c.pendingParseData = ""
	return nil
}

// UnmarshalJSON unmarshals a controller config JSON blob pushing the data into
// the pending parse buffer.
func (c *Config) UnmarshalJSON(data []byte) error {
	// assert that the object is a map
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	c.pendingParseData = string(data)
	return nil
}

// MarshalJSON marshals a controller config JSON blob.
func (c *Config) MarshalJSON() ([]byte, error) {
	if c.underlying == nil {
		return []byte("null"), nil
	}
	return c.underlying.MarshalJSON()
}

// GetConfig returns the underlying config after Resolve.
func (c *Config) GetConfig() config.Config {
	return c.underlying
}

// _ is a type assertion
var (
	_ json.Unmarshaler = ((*Config)(nil))
	_ json.Marshaler   = ((*Config)(nil))
)
