package config

// Placeholder is used for type assertions.
// var _ controller.Controller = (*GenericController[config.Placeholder])(nil)

// Validate validates the configuration.
// This is a cursory validation to see if the values "look correct."
func (c *Placeholder) Validate() error {
	return nil
}

// GetConfigID returns the unique string for this configuration type.
func (c *Placeholder) GetConfigID() string {
	return "controllerbus/placeholder"
}

// EqualsConfig checks if the config is equal to another.
func (c *Placeholder) EqualsConfig(other Config) bool {
	_, ok := other.(*Placeholder)
	return ok
}

// _ is a type assertion
var _ Config = ((*Placeholder)(nil))
