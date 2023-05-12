package configset_proto

import (
	"context"
	"encoding/base64"
	"encoding/json"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/valyala/fastjson"
	jsonpb "google.golang.org/protobuf/encoding/protojson"
)

// NewControllerConfig constructs a new controller config.
func NewControllerConfig(c configset.ControllerConfig, useJson bool) (*ControllerConfig, error) {
	conf := c.GetConfig()
	cID := conf.GetConfigID()

	var confData []byte
	var err error
	if useJson {
		m := &jsonpb.MarshalOptions{}
		confData, err = m.Marshal(conf)
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
			var p fastjson.Parser
			_, err := p.ParseBytes(conf)
			if err != nil {
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
			err = jsonpb.Unmarshal(configData, cf)
		} else {
			err = cf.UnmarshalVT(configData)
		}
		if err != nil {
			return nil, err
		}
	}

	return configset.NewControllerConfig(c.GetRev(), cf), nil
}

// UnmarshalJSON unmarshals json to the controller config.
// For the config field: supports JSON, YAML, or a string containing either.
func (c *ControllerConfig) UnmarshalJSON(data []byte) error {
	jdata, err := yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}
	var p fastjson.Parser
	v, err := p.ParseBytes(jdata)
	if err != nil {
		return err
	}
	if v.Exists("id") {
		c.Id = string(v.GetStringBytes("id"))
	}
	// backwards compatible
	if v.Exists("revision") {
		c.Rev = v.GetUint64("revision")
	}
	if v.Exists("rev") {
		c.Rev = v.GetUint64("rev")
	}
	if v.Exists("config") {
		var configVal *fastjson.Value
		configStr := v.GetStringBytes("config")
		if len(configStr) != 0 {
			// parse json and/or yaml
			configJson, err := yaml.YAMLToJSON(configStr)
			if err != nil {
				return err
			}
			var cj fastjson.Parser
			configVal, err = cj.ParseBytes(configJson)
			if err != nil {
				return err
			}
		} else {
			// expect a object value
			configVal = v.Get("config")
			if t := configVal.Type(); t != fastjson.TypeObject {
				return errors.Errorf("config: expected json object but got %s", t.String())
			}
		}
		// re-marshal to json
		c.Config = configVal.MarshalTo(nil)
	}
	return nil
}

// MarshalJSON marshals json from the controller config.
// For the config field: supports JSON, YAML, or a string containing either.
func (c *ControllerConfig) MarshalJSON() ([]byte, error) {
	outCtr := gabs.New()

	// marshal the regular fields
	if rev := c.GetRev(); rev != 0 {
		_, err := outCtr.Set(rev, "rev")
		if err != nil {
			return nil, err
		}
	}
	if configID := c.GetId(); configID != "" {
		_, err := outCtr.Set(configID, "id")
		if err != nil {
			return nil, err
		}
	}

	if confFieldData := c.GetConfig(); len(confFieldData) != 0 {
		// detect if the config field is json, if so, set it as inline json.
		if confFieldData[0] == '{' && confFieldData[len(confFieldData)-1] == '}' {
			confJSON, err := gabs.ParseJSON(confFieldData)
			if err != nil {
				return nil, err
			}
			_, err = outCtr.Set(confJSON, "config")
			if err != nil {
				return nil, err
			}
		} else {
			// otherwise encode it as base64 (this is what jsonpb does)
			_, err := outCtr.Set(base64.StdEncoding.EncodeToString(confFieldData), "config")
			if err != nil {
				return nil, err
			}
		}
	}

	// finalize the json
	return outCtr.EncodeJSON(), nil
}

// _ is a type assertion
var (
	_ json.Unmarshaler = ((*ControllerConfig)(nil))
	_ json.Marshaler   = ((*ControllerConfig)(nil))
)
