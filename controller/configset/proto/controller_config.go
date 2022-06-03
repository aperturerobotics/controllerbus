package configset_proto

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// NewControllerConfig constructs a new controller config.
func NewControllerConfig(c configset.ControllerConfig) (*ControllerConfig, error) {
	conf := c.GetConfig()
	cID := conf.GetConfigID()
	confData, err := proto.Marshal(conf)
	if err != nil {
		return nil, err
	}
	return &ControllerConfig{
		Id:       cID,
		Config:   confData,
		Revision: c.GetRevision(),
	}, nil
}

// Validate performs cursory validation of the controller config.
func (c *ControllerConfig) Validate() error {
	if len(c.GetId()) == 0 {
		return ErrControllerConfigIdEmpty
	}
	return nil
}

// Resolve resolves the config into a configset.ControllerConfig
func (c *ControllerConfig) Resolve(ctx context.Context, b bus.Bus) (configset.ControllerConfig, error) {
	if len(c.GetId()) == 0 {
		return nil, ErrControllerConfigIdEmpty
	}

	configCtorDir := resolver.NewLoadConfigConstructorByID(c.GetId())
	configCtorVal, configCtorRef, err := bus.ExecOneOff(ctx, b, configCtorDir, false, nil)
	if err != nil {
		if err == context.Canceled {
			return nil, err
		}
		return nil, errors.WithMessage(err, "resolve config object")
	}
	defer configCtorRef.Release()

	ctor, ctorOk := configCtorVal.GetValue().(config.Constructor)
	if !ctorOk {
		return nil, errors.New("load config constructor directive returned invalid object")
	}
	cf := ctor.ConstructConfig()
	if err := proto.Unmarshal(c.GetConfig(), cf); err != nil {
		return nil, err
	}
	return configset.NewControllerConfig(c.GetRevision(), cf), nil
}
