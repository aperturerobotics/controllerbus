// Code generated by protoc-gen-go-lite. DO NOT EDIT.
// protoc-gen-go-lite version: v0.5.0
// source: github.com/aperturerobotics/controllerbus/controller/configset/proto/configset.proto

package configset_proto

import (
	base64 "encoding/base64"
	io "io"
	strconv "strconv"
	strings "strings"

	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
	json "github.com/aperturerobotics/protobuf-go-lite/json"
	errors "github.com/pkg/errors"
)

// ConfigSet contains a configuration set.
type ConfigSet struct {
	unknownFields []byte
	// Configs contains the controller configurations.
	Configs map[string]*ControllerConfig `protobuf:"bytes,1,rep,name=configs,proto3" json:"configs,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *ConfigSet) Reset() {
	*x = ConfigSet{}
}

func (*ConfigSet) ProtoMessage() {}

func (x *ConfigSet) GetConfigs() map[string]*ControllerConfig {
	if x != nil {
		return x.Configs
	}
	return nil
}

// ControllerConfig contains a controller configuration.
//
// protobuf-go-lite:disable-json
type ControllerConfig struct {
	unknownFields []byte
	// Id is the config ID.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Rev is the revision number of the configuration.
	Rev uint64 `protobuf:"varint,2,opt,name=rev,proto3" json:"rev,omitempty"`
	// Config is the configuration object.
	// Proto supports: protobuf (binary) and json (starting with {).
	// Json supports: protobuf (base64) and json (inline object).
	Config []byte `protobuf:"bytes,3,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *ControllerConfig) Reset() {
	*x = ControllerConfig{}
}

func (*ControllerConfig) ProtoMessage() {}

func (x *ControllerConfig) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ControllerConfig) GetRev() uint64 {
	if x != nil {
		return x.Rev
	}
	return 0
}

func (x *ControllerConfig) GetConfig() []byte {
	if x != nil {
		return x.Config
	}
	return nil
}

type ConfigSet_ConfigsEntry struct {
	unknownFields []byte
	Key           string            `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value         *ControllerConfig `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *ConfigSet_ConfigsEntry) Reset() {
	*x = ConfigSet_ConfigsEntry{}
}

func (*ConfigSet_ConfigsEntry) ProtoMessage() {}

func (x *ConfigSet_ConfigsEntry) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *ConfigSet_ConfigsEntry) GetValue() *ControllerConfig {
	if x != nil {
		return x.Value
	}
	return nil
}

func (m *ConfigSet) CloneVT() *ConfigSet {
	if m == nil {
		return (*ConfigSet)(nil)
	}
	r := new(ConfigSet)
	if rhs := m.Configs; rhs != nil {
		tmpContainer := make(map[string]*ControllerConfig, len(rhs))
		for k, v := range rhs {
			tmpContainer[k] = v.CloneVT()
		}
		r.Configs = tmpContainer
	}
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *ConfigSet) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (m *ControllerConfig) CloneVT() *ControllerConfig {
	if m == nil {
		return (*ControllerConfig)(nil)
	}
	r := new(ControllerConfig)
	r.Id = m.Id
	r.Rev = m.Rev
	if rhs := m.Config; rhs != nil {
		tmpBytes := make([]byte, len(rhs))
		copy(tmpBytes, rhs)
		r.Config = tmpBytes
	}
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *ControllerConfig) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (this *ConfigSet) EqualVT(that *ConfigSet) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if len(this.Configs) != len(that.Configs) {
		return false
	}
	for i, vx := range this.Configs {
		vy, ok := that.Configs[i]
		if !ok {
			return false
		}
		if p, q := vx, vy; p != q {
			if p == nil {
				p = &ControllerConfig{}
			}
			if q == nil {
				q = &ControllerConfig{}
			}
			if !p.EqualVT(q) {
				return false
			}
		}
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *ConfigSet) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*ConfigSet)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}
func (this *ControllerConfig) EqualVT(that *ControllerConfig) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if this.Id != that.Id {
		return false
	}
	if this.Rev != that.Rev {
		return false
	}
	if string(this.Config) != string(that.Config) {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *ControllerConfig) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*ControllerConfig)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}

// MarshalProtoJSON marshals the ConfigSet_ConfigsEntry message to JSON.
func (x *ConfigSet_ConfigsEntry) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.Key != "" || s.HasField("key") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("key")
		s.WriteString(x.Key)
	}
	if x.Value != nil || s.HasField("value") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("value")
		x.Value.MarshalProtoJSON(s.WithField("value"))
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the ConfigSet_ConfigsEntry to JSON.
func (x *ConfigSet_ConfigsEntry) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the ConfigSet_ConfigsEntry message from JSON.
func (x *ConfigSet_ConfigsEntry) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.ReadAny() // ignore unknown field
		case "key":
			s.AddField("key")
			x.Key = s.ReadString()
		case "value":
			if s.ReadNil() {
				x.Value = nil
				return
			}
			x.Value = &ControllerConfig{}
			x.Value.UnmarshalProtoJSON(s.WithField("value", true))
		}
	})
}

// UnmarshalJSON unmarshals the ConfigSet_ConfigsEntry from JSON.
func (x *ConfigSet_ConfigsEntry) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the ConfigSet message to JSON.
func (x *ConfigSet) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.Configs != nil || s.HasField("configs") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("configs")
		s.WriteObjectStart()
		var wroteElement bool
		for k, v := range x.Configs {
			s.WriteMoreIf(&wroteElement)
			s.WriteObjectStringField(k)
			v.MarshalProtoJSON(s.WithField("configs"))
		}
		s.WriteObjectEnd()
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the ConfigSet to JSON.
func (x *ConfigSet) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the ConfigSet message from JSON.
func (x *ConfigSet) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.ReadAny() // ignore unknown field
		case "configs":
			s.AddField("configs")
			if s.ReadNil() {
				x.Configs = nil
				return
			}
			x.Configs = make(map[string]*ControllerConfig)
			s.ReadStringMap(func(key string) {
				var v ControllerConfig
				v.UnmarshalProtoJSON(s)
				x.Configs[key] = &v
			})
		}
	})
}

// UnmarshalJSON unmarshals the ConfigSet from JSON.
func (x *ConfigSet) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

func (m *ConfigSet) MarshalVT() (dAtA []byte, err error) {
	if m == nil {
		return nil, nil
	}
	size := m.SizeVT()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBufferVT(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ConfigSet) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *ConfigSet) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
	if m == nil {
		return 0, nil
	}
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.unknownFields != nil {
		i -= len(m.unknownFields)
		copy(dAtA[i:], m.unknownFields)
	}
	if len(m.Configs) > 0 {
		for k := range m.Configs {
			v := m.Configs[k]
			baseI := i
			size, err := v.MarshalToSizedBufferVT(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(size))
			i--
			dAtA[i] = 0x12
			i -= len(k)
			copy(dAtA[i:], k)
			i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(k)))
			i--
			dAtA[i] = 0xa
			i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *ControllerConfig) MarshalVT() (dAtA []byte, err error) {
	if m == nil {
		return nil, nil
	}
	size := m.SizeVT()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBufferVT(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ControllerConfig) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *ControllerConfig) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
	if m == nil {
		return 0, nil
	}
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.unknownFields != nil {
		i -= len(m.unknownFields)
		copy(dAtA[i:], m.unknownFields)
	}
	if len(m.Config) > 0 {
		i -= len(m.Config)
		copy(dAtA[i:], m.Config)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.Config)))
		i--
		dAtA[i] = 0x1a
	}
	if m.Rev != 0 {
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(m.Rev))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Id) > 0 {
		i -= len(m.Id)
		copy(dAtA[i:], m.Id)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.Id)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *ConfigSet) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Configs) > 0 {
		for k, v := range m.Configs {
			_ = k
			_ = v
			l = 0
			if v != nil {
				l = v.SizeVT()
			}
			l += 1 + protobuf_go_lite.SizeOfVarint(uint64(l))
			mapEntrySize := 1 + len(k) + protobuf_go_lite.SizeOfVarint(uint64(len(k))) + l
			n += mapEntrySize + 1 + protobuf_go_lite.SizeOfVarint(uint64(mapEntrySize))
		}
	}
	n += len(m.unknownFields)
	return n
}

func (m *ControllerConfig) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	if m.Rev != 0 {
		n += 1 + protobuf_go_lite.SizeOfVarint(uint64(m.Rev))
	}
	l = len(m.Config)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	n += len(m.unknownFields)
	return n
}

func (x *ConfigSet_ConfigsEntry) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("ConfigsEntry { ")
	if x.Key != "" {
		sb.WriteString(" key: ")
		sb.WriteString(strconv.Quote(x.Key))
	}
	if x.Value != nil {
		sb.WriteString(" value: ")
		sb.WriteString(x.Value.MarshalProtoText())
	}
	sb.WriteString("}")
	return sb.String()
}
func (x *ConfigSet_ConfigsEntry) String() string {
	return x.MarshalProtoText()
}
func (x *ConfigSet) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("ConfigSet { ")
	if len(x.Configs) > 0 {
		sb.WriteString(" configs: {")
		for k, v := range x.Configs {
			sb.WriteString(" ")
			sb.WriteString(strconv.Quote(k))
			sb.WriteString(": ")
			sb.WriteString(v.MarshalProtoText())
		}
		sb.WriteString(" }")
	}
	sb.WriteString("}")
	return sb.String()
}
func (x *ConfigSet) String() string {
	return x.MarshalProtoText()
}
func (x *ControllerConfig) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("ControllerConfig { ")
	if x.Id != "" {
		sb.WriteString(" id: ")
		sb.WriteString(strconv.Quote(x.Id))
	}
	if x.Rev != 0 {
		sb.WriteString(" rev: ")
		sb.WriteString(strconv.FormatUint(uint64(x.Rev), 10))
	}
	if len(x.Config) > 0 {
		sb.WriteString(" config: ")
		sb.WriteString("\"")
		sb.WriteString(base64.StdEncoding.EncodeToString(x.Config))
		sb.WriteString("\"")
	}
	sb.WriteString("}")
	return sb.String()
}
func (x *ControllerConfig) String() string {
	return x.MarshalProtoText()
}
func (m *ConfigSet) UnmarshalVT(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return protobuf_go_lite.ErrIntOverflow
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return errors.Errorf("proto: ConfigSet: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return errors.Errorf("proto: ConfigSet: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return errors.Errorf("proto: wrong wireType = %d for field Configs", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Configs == nil {
				m.Configs = make(map[string]*ControllerConfig)
			}
			var mapkey string
			var mapvalue *ControllerConfig
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protobuf_go_lite.ErrIntOverflow
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					wire |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				fieldNum := int32(wire >> 3)
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return protobuf_go_lite.ErrIntOverflow
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return protobuf_go_lite.ErrInvalidLength
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey < 0 {
						return protobuf_go_lite.ErrInvalidLength
					}
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var mapmsglen int
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return protobuf_go_lite.ErrIntOverflow
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapmsglen |= int(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					if mapmsglen < 0 {
						return protobuf_go_lite.ErrInvalidLength
					}
					postmsgIndex := iNdEx + mapmsglen
					if postmsgIndex < 0 {
						return protobuf_go_lite.ErrInvalidLength
					}
					if postmsgIndex > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = &ControllerConfig{}
					if err := mapvalue.UnmarshalVT(dAtA[iNdEx:postmsgIndex]); err != nil {
						return err
					}
					iNdEx = postmsgIndex
				} else {
					iNdEx = entryPreIndex
					skippy, err := protobuf_go_lite.Skip(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if (skippy < 0) || (iNdEx+skippy) < 0 {
						return protobuf_go_lite.ErrInvalidLength
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.Configs[mapkey] = mapvalue
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := protobuf_go_lite.Skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.unknownFields = append(m.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ControllerConfig) UnmarshalVT(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return protobuf_go_lite.ErrIntOverflow
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return errors.Errorf("proto: ControllerConfig: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return errors.Errorf("proto: ControllerConfig: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return errors.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return errors.Errorf("proto: wrong wireType = %d for field Rev", wireType)
			}
			m.Rev = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Rev |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return errors.Errorf("proto: wrong wireType = %d for field Config", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Config = append(m.Config[:0], dAtA[iNdEx:postIndex]...)
			if m.Config == nil {
				m.Config = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := protobuf_go_lite.Skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.unknownFields = append(m.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
