package directive_proto

import (
	"errors"

	"github.com/aperturerobotics/controllerbus/directive"
	"google.golang.org/protobuf/proto"
)

// ErrNotProtobuf is returned when using a protobuf codec on a non-protobuf message.
var ErrNotProtobuf = errors.New("protobuf codec used on non-protobuf message")

// protoCodec implements NetworkCodec with protobuf.
type protoCodec struct{}

// protoCodecSingleton is a single proto codec instance.
var protoCodecSingleton = protoCodec{}

// GetProtobufCodec returns the global instance of the protobuf codec.
func GetProtobufCodec() directive.NetworkedCodec {
	return &protoCodecSingleton
}

// vtMarshal is the vtProtobuf marshal type.
type vtMarshal interface {
	MarshalVT() ([]byte, error)
}

// Marshal encodes the networked directive.
func (c *protoCodec) Marshal(dir directive.Networked) ([]byte, error) {
	if dir == nil {
		return nil, nil
	}

	vtpb, vtpbOk := dir.(vtMarshal)
	if vtpbOk {
		return vtpb.MarshalVT()
	}
	pb, pbOk := dir.(proto.Message)
	if !pbOk {
		return nil, ErrNotProtobuf
	}
	return proto.Marshal(pb)
}

// vtUnmarshal is the vtProtobuf unmarshal type.
type vtUnmarshal interface {
	UnmarshalVT([]byte) error
}

// Unmarshal decodes the data to the networked directive.
// The type must match the expected type for the codec.
func (c *protoCodec) Unmarshal(data []byte, dir directive.Networked) error {
	if dir == nil {
		return nil
	}

	vtpb, vtpbOk := dir.(vtUnmarshal)
	if vtpbOk {
		return vtpb.UnmarshalVT(data)
	}
	pb, pbOk := dir.(proto.Message)
	if !pbOk {
		return ErrNotProtobuf
	}
	return proto.Unmarshal(data, pb)
}

// _ is a type assertion
var _ directive.NetworkedCodec = ((*protoCodec)(nil))
