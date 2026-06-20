package configset_proto

import (
	"encoding/json"
	"testing"

	cbyaml "github.com/aperturerobotics/controllerbus/yaml"
)

var mockControllerConfig = `
id: test/controller
rev: 4
config:
  test: true
`

// TestParseControllerConfig tests parsing a controller config yaml.
func TestParseControllerConfig(t *testing.T) {
	conf := &ControllerConfig{}
	jdat, err := cbyaml.YAMLToJSON([]byte(mockControllerConfig))
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := json.Unmarshal(jdat, conf); err != nil {
		t.Fatal(err.Error())
	}
}
