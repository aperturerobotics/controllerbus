package configset_proto

import (
	"testing"

	"github.com/ghodss/yaml"
)

var mockControllerConfig = `
id: test/controller
revision: 4
config:
  test: true
`

// TestParseControllerConfig tests parsing a controller config yaml.
func TestParseControllerConfig(t *testing.T) {
	conf := &ControllerConfig{}
	if err := yaml.Unmarshal([]byte(mockControllerConfig), conf); err != nil {
		t.Fatal(err.Error())
	}
}
