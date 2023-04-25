package configset_proto

import (
	"context"
	"testing"

	"github.com/aperturerobotics/controllerbus/controller/configset"
	configset_controller "github.com/aperturerobotics/controllerbus/controller/configset/controller"
	"github.com/aperturerobotics/controllerbus/core"
	boilerplate_controller "github.com/aperturerobotics/controllerbus/example/boilerplate/controller"
	"github.com/sirupsen/logrus"
	jsonpb "google.golang.org/protobuf/encoding/protojson"
)

// TestE2E tests configset proto end to end.
func TestE2E(t *testing.T) {
	c := make(configset.ConfigSet)
	c["test"] = configset.NewControllerConfig(1, &configset_controller.Config{})
	m, err := NewConfigSet(c, true)
	if err != nil {
		t.Fatal(err.Error())
	}
	mDat, err := m.MarshalVT()
	if err != nil {
		t.Fatal(err.Error())
	}

	mDec := &ConfigSet{}
	if err := mDec.UnmarshalVT(mDat); err != nil {
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

	cs, err := mDec.Resolve(ctx, b)
	if err != nil {
		t.Fatal(err.Error())
	}
	_ = sr
	_ = cs
}

// TestJSON ensures that we can marshal and unmarshal from JSON.
// Tests when we have JSON in the Config field.
func TestJSON(t *testing.T) {
	mEnc := &ConfigSet{
		Configurations: map[string]*ControllerConfig{
			"test-config": {
				Rev:    1,
				Id:     boilerplate_controller.ConfigID,
				Config: []byte(`{"exampleField":"hello world"}`),
			},
		},
	}
	t.Logf("%v", mEnc)
	dat, err := jsonpb.Marshal(mEnc)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("%v", string(dat))

	mDec := &ConfigSet{}
	if err := jsonpb.Unmarshal(dat, mDec); err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("%v", mDec)

	if !mEnc.EqualVT(mDec) {
		t.Fatal("values not identical after decoding")
	}
}
