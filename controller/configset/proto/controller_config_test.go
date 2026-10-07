package configset_proto

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aperturerobotics/controllerbus/core"
	cbyaml "github.com/aperturerobotics/controllerbus/yaml"
	"github.com/sirupsen/logrus"
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

// TestResolveUnknownConfigID tests that resolving a config id no factory
// registers fails instead of waiting.
func TestResolveUnknownConfigID(t *testing.T) {
	b, _, err := core.NewCoreBus(t.Context(), logrus.NewEntry(logrus.New()))
	if err != nil {
		t.Fatal(err)
	}
	conf := &ControllerConfig{Id: "test/unregistered"}
	if _, err := conf.Resolve(t.Context(), b); !errors.Is(err, ErrUnknownConfigID) {
		t.Fatalf("Resolve = %v, want ErrUnknownConfigID", err)
	}
}
