package configset

import (
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/golang/protobuf/proto"
)

// NewControllerConfig constructs a controller config object.
func NewControllerConfig(rev uint64, conf config.Config) (*ControllerConfig, error) {
	dat, err := proto.Marshal(conf)
	if err != nil {
		return nil, err
	}

	return &ControllerConfig{
		ConfigId: conf.GetConfigID(),
		Revision: rev,
		Data:     dat,
	}, nil
}
