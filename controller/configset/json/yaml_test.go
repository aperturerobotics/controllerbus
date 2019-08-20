package configset_json

import (
	"context"
	"testing"

	"github.com/aperturerobotics/controllerbus/controller/configset"
	"github.com/aperturerobotics/controllerbus/core"
	"github.com/aperturerobotics/controllerbus/example/boilerplate"
	"github.com/sirupsen/logrus"
	// "github.com/aperturerobotics/controllerbus/bus/inmem"
	// directive_controller "github.com/aperturerobotics/controllerbus/directive/controller"
	// "github.com/aperturerobotics/controllerbus/controller/resolver"
	// "github.com/aperturerobotics/controllerbus/controller/resolver/static"
)

var basicYAMLOutput = `test:
  config:
    example_field: test 123
  id: controllerbus/example/boilerplate/1
  revision: 1
`

// TestMarshalConfigSetYAML tests marshaling a config set to yaml.
func TestMarshalConfigSetYAML(t *testing.T) {
	c := make(configset.ConfigSet)
	c["test"] = configset.NewControllerConfig(1, &boilerplate.Config{ExampleField: "test 123"})
	dat, err := MarshalYAML(c)
	if err != nil {
		t.Fatal(err.Error())
	}
	v := string(dat)
	if v != basicYAMLOutput {
		t.Fatalf("unexpected output %s", v)
	}
}

// TestUnmarshalConfigSetYAML tests unmarshaling a config set from yaml.
func TestUnmarshalConfigSetYAML(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)
	ctx := context.Background()
	b, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		t.Fatal(err.Error())
	}
	sr.AddFactory(boilerplate.NewFactory(b))

	ocs := make(configset.ConfigSet)
	if err := UnmarshalYAML(ctx, b, []byte(basicYAMLOutput), ocs, true); err != nil {
		t.Fatal(err.Error())
	}
}
