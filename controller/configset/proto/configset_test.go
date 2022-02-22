package configset_proto

import (
	"context"
	"testing"

	"github.com/aperturerobotics/controllerbus/controller/configset"
	configset_controller "github.com/aperturerobotics/controllerbus/controller/configset/controller"
	"github.com/aperturerobotics/controllerbus/core"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

// TestE2E tests configset proto end to end.
func TestE2E(t *testing.T) {
	c := make(configset.ConfigSet)
	c["test"] = configset.NewControllerConfig(1, &configset_controller.Config{})
	m, err := NewConfigSet(c)
	if err != nil {
		t.Fatal(err.Error())
	}
	mDat, err := proto.Marshal(m)
	if err != nil {
		t.Fatal(err.Error())
	}

	mDec := &ConfigSet{}
	if err := proto.Unmarshal(mDat, mDec); err != nil {
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
