package configset_proto

import (
	"context"
	"encoding/base64"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	jsoniter "github.com/aperturerobotics/json-iterator-lite"
	"github.com/aperturerobotics/protobuf-go-lite/json"
	"github.com/pkg/errors"
)

// NewControllerConfig constructs a new controller config.
func NewControllerConfig(c configset.ControllerConfig, useJson bool) (*ControllerConfig, error) {
	conf := c.GetConfig()
	cID := conf.GetConfigID()

	var confData []byte
	var err error
	if useJson {
		confData, err = conf.MarshalJSON()
	} else {
		confData, err = conf.MarshalVT()
	}
	if err != nil {
		return nil, err
	}

	return &ControllerConfig{
		Id:     cID,
		Config: confData,
		Rev:    c.GetRev(),
	}, nil
}

// Validate performs cursory validation of the controller config.
func (c *ControllerConfig) Validate() error {
	if len(c.GetId()) == 0 {
		return ErrControllerConfigIdEmpty
	}
	if conf := c.GetConfig(); len(conf) != 0 {
		// json if first character is {
		if conf[0] == 123 {
			// validate json
			it := jsoniter.ParseBytes(conf)
			if err := it.Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// Resolve resolves the config into a configset.ControllerConfig
func (c *ControllerConfig) Resolve(ctx context.Context, b bus.Bus) (configset.ControllerConfig, error) {
	if len(c.GetId()) == 0 {
		return nil, ErrControllerConfigIdEmpty
	}

	configCtorDir := resolver.NewLoadConfigConstructorByID(c.GetId())
	configCtorVal, _, configCtorRef, err := bus.ExecOneOff(ctx, b, configCtorDir, nil, nil)
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

	// detect json or protobuf & parse
	if configData := c.GetConfig(); len(configData) != 0 {
		// if configData[0] == '{'
		if configData[0] == 123 {
			err = cf.UnmarshalJSON(configData)
		} else {
			err = cf.UnmarshalVT(configData)
		}
		if err != nil {
			return nil, err
		}
	}

	return configset.NewControllerConfig(c.GetRev(), cf), nil
}

// MarshalProtoJSON marshals the ControllerConfig message to JSON.
func (c *ControllerConfig) MarshalProtoJSON(s *json.MarshalState) {
	if c == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if c.Id != "" || s.HasField("id") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("id")
		s.WriteString(c.Id)
	}
	if c.Rev != 0 || s.HasField("rev") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("rev")
		s.WriteUint64(c.Rev)
	}
	if len(c.Config) > 0 || s.HasField("config") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("config")
		// Detect if config is JSON
		if c.Config[0] == '{' && c.Config[len(c.Config)-1] == '}' {
			// Ensure json is parseable
			it := jsoniter.ParseBytes(c.Config)
			if err := it.Error; err != nil {
				s.SetError(errors.Wrap(err, "unable to parse config json"))
				return
			}
			_, err := s.Write(c.Config)
			if err != nil {
				s.SetError(err)
				return
			}
		} else {
			// Base58 encoded string
			s.WriteString(base64.RawStdEncoding.EncodeToString(c.Config))
		}
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the ControllerConfig to JSON.
func (c *ControllerConfig) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(c)
}

// UnmarshalJSON unmarshals the ControllerConfig from JSON.
func (c *ControllerConfig) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, c)
}

func (c *ControllerConfig) UnmarshalProtoJSON(s *json.UnmarshalState) {
	for key := s.ReadObjectField(); key != ""; key = s.ReadObjectField() {
		switch key {
		case "id":
			c.Id = s.ReadString()
		case "rev", "revision":
			c.Rev = s.ReadUint64()
		case "config":
			if s.ReadNil() {
				break
			}
			nextTok := s.WhatIsNext()
			switch nextTok {
			case jsoniter.StringValue:
				// Expect base58 encoded string
				var err error
				c.Config, err = base64.RawStdEncoding.DecodeString(s.ReadString())
				if err != nil {
					s.SetError(errors.Wrap(err, "unmarshal config value as base58 string"))
					return
				}
			case jsoniter.ObjectValue:
				c.Config = s.SkipAndReturnBytes()
			default:
				s.SetError(errors.Errorf("invalid json value for config: type %v", nextTok))
				return
			}
		default:
			s.Skip()
		}
	}
}

// _ is a type assertion
var (
	_ json.Unmarshaler = ((*ControllerConfig)(nil))
	_ json.Marshaler   = ((*ControllerConfig)(nil))
)
