// Code generated by protoc-gen-go-lite. DO NOT EDIT.
// protoc-gen-go-lite version: v0.5.0
// source: github.com/aperturerobotics/controllerbus/config/config.proto

package config

import (
	io "io"
	strings "strings"

	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
	json "github.com/aperturerobotics/protobuf-go-lite/json"
	errors "github.com/pkg/errors"
)

// Placeholder is a config used for type assertions.
type Placeholder struct {
	unknownFields []byte
}

func (x *Placeholder) Reset() {
	*x = Placeholder{}
}

func (*Placeholder) ProtoMessage() {}

func (m *Placeholder) CloneVT() *Placeholder {
	if m == nil {
		return (*Placeholder)(nil)
	}
	r := new(Placeholder)
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *Placeholder) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (this *Placeholder) EqualVT(that *Placeholder) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *Placeholder) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*Placeholder)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}

// MarshalProtoJSON marshals the Placeholder message to JSON.
func (x *Placeholder) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	s.WriteObjectEnd()
}

// MarshalJSON marshals the Placeholder to JSON.
func (x *Placeholder) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the Placeholder message from JSON.
func (x *Placeholder) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
}

// UnmarshalJSON unmarshals the Placeholder from JSON.
func (x *Placeholder) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

func (m *Placeholder) MarshalVT() (dAtA []byte, err error) {
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

func (m *Placeholder) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *Placeholder) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
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
	return len(dAtA) - i, nil
}

func (m *Placeholder) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += len(m.unknownFields)
	return n
}

func (x *Placeholder) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("Placeholder { ")
	sb.WriteString("}")
	return sb.String()
}
func (x *Placeholder) String() string {
	return x.MarshalProtoText()
}
func (m *Placeholder) UnmarshalVT(dAtA []byte) error {
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
			return errors.Errorf("proto: Placeholder: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return errors.Errorf("proto: Placeholder: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
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
