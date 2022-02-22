package configset_json

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aperturerobotics/controllerbus/controller/configset"
	configset_controller "github.com/aperturerobotics/controllerbus/controller/configset/controller"
	"github.com/aperturerobotics/controllerbus/core"
	"github.com/sirupsen/logrus"
	// "github.com/aperturerobotics/controllerbus/bus/inmem"
	// directive_controller "github.com/aperturerobotics/controllerbus/directive/controller"
	// "github.com/aperturerobotics/controllerbus/controller/resolver"
	// "github.com/aperturerobotics/controllerbus/controller/resolver/static"
)

var basicOutput = `{"test":{"revision":1,"id":"controllerbus/configset/1","config":{}}}`

// TestMarshalConfigSet tests marshaling a config set to json.
func TestMarshalConfigSet(t *testing.T) {
	c := make(configset.ConfigSet)
	c["test"] = configset.NewControllerConfig(1, &configset_controller.Config{})
	m := NewConfigSet(c)
	dat, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err.Error())
	}
	v := string(dat)
	if v != basicOutput {
		t.Fatalf("unexpected output %s", v)
	}
}

// TestUnmarshalConfigSet tests unmarshaling a config set to json.
func TestUnmarshalConfigSet(t *testing.T) {
	c := new(ConfigSet)
	if err := json.Unmarshal([]byte(basicOutput), c); err != nil {
		t.Fatal(err.Error())
	}

	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)
	ctx := context.Background()
	b, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		t.Fatal(err.Error())
	}

	cs, err := c.Resolve(ctx, b)
	if err != nil {
		t.Fatal(err.Error())
	}
	_ = cs
	_ = sr
}
