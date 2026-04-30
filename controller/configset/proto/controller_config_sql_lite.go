//go:build sql_lite

package configset_proto

import (
	"encoding/base64"

	jsoniter "github.com/aperturerobotics/json-iterator-lite"
	"github.com/aperturerobotics/protobuf-go-lite/json"
	"github.com/pkg/errors"
)

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
		if c.Config[0] == '{' && c.Config[len(c.Config)-1] == '}' {
			iter := jsoniter.ParseBytes(c.Config)
			if iter.Error != nil {
				s.SetError(errors.Wrap(iter.Error, "unable to parse config json"))
				return
			}
			if _, err := s.Write(c.Config); err != nil {
				s.SetError(err)
				return
			}
		} else {
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

// UnmarshalProtoJSON unmarshals the ControllerConfig from JSON state.
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
			next := s.WhatIsNext()
			switch next {
			case jsoniter.StringValue:
				var err error
				c.Config, err = base64.RawStdEncoding.DecodeString(s.ReadString())
				if err != nil {
					s.SetError(errors.Wrap(err, "unmarshal config value as base58 string"))
					return
				}
			case jsoniter.ObjectValue:
				c.Config = s.SkipAndReturnBytes()
			default:
				s.SetError(errors.Errorf("invalid json value for config: type %v", next))
				return
			}
		default:
			s.Skip()
		}
	}
}
