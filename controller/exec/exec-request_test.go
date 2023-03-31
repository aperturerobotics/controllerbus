package controller_exec

import (
	"context"
	"testing"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	configset_controller "github.com/aperturerobotics/controllerbus/controller/configset/controller"
	configset_proto "github.com/aperturerobotics/controllerbus/controller/configset/proto"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/core"
	boilerplate "github.com/aperturerobotics/controllerbus/example/boilerplate/controller"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

// TestExecControllerRequest tests the ExecControllerRequest.
func TestExecControllerRequest(t *testing.T) {
	c := make(configset.ConfigSet)
	c["test-will-succeed"] = configset.NewControllerConfig(1, &boilerplate.Config{
		ExampleField: "hello world",
	})
	c["test-will-fail"] = configset.NewControllerConfig(1, &boilerplate.Config{
		FailWithErr: "testing error",
	})
	m, err := configset_proto.NewConfigSet(c, false)
	if err != nil {
		t.Fatal(err.Error())
	}
	req := &ExecControllerRequest{
		ConfigSet: m,
	}
	t.Log(req.String())

	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)
	ctx := context.Background()
	b, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		t.Fatal(err.Error())
	}
	sr.AddFactory(boilerplate.NewFactory(b))

	_, _, csCtrlRef, err := bus.ExecOneOff(
		ctx,
		b,
		resolver.NewLoadControllerWithConfig(&configset_controller.Config{}),
		false,
		nil,
	)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer csCtrlRef.Release()

	respCh := make(chan *ExecControllerResponse, 5)
	errCh := make(chan error, 5)
	go func() {
		errCh <- req.Execute(ctx, b, true, func(resp *ExecControllerResponse) error {
			respCh <- proto.Clone(resp).(*ExecControllerResponse)
			return nil
		})
	}()

	var seenFail, seenSuccess bool
	for {
		select {
		case resp := <-respCh:
			t.Logf("got response: %s", resp.String())
			if resp.Id == "test-will-succeed" &&
				resp.Status == ControllerStatus_ControllerStatus_RUNNING {
				seenSuccess = true
			}
			if resp.Id == "test-will-fail" &&
				resp.Status == ControllerStatus_ControllerStatus_ERROR {
				seenFail = true
			}
			if seenSuccess && seenFail {
				return
			}
		case err := <-errCh:
			if err == nil {
				return
			}
			t.Fatal(err.Error())
		}
	}
}
