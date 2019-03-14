// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/aperturerobotics/controllerbus/controller/exec/exec.proto

package controller_exec

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// ControllerStatus holds basic status for a controller.
type ControllerStatus int32

const (
	// ControllerStatus_UNKNOWN is unrecognized.
	ControllerStatus_ControllerStatus_UNKNOWN ControllerStatus = 0
	// ControllerStatus_CONFIGURING indicates the controller is configuring.
	ControllerStatus_ControllerStatus_CONFIGURING ControllerStatus = 1
	// ControllerStatus_RUNNING indicates the controller is running.
	ControllerStatus_ControllerStatus_RUNNING ControllerStatus = 2
)

var ControllerStatus_name = map[int32]string{
	0: "ControllerStatus_UNKNOWN",
	1: "ControllerStatus_CONFIGURING",
	2: "ControllerStatus_RUNNING",
}

var ControllerStatus_value = map[string]int32{
	"ControllerStatus_UNKNOWN":     0,
	"ControllerStatus_CONFIGURING": 1,
	"ControllerStatus_RUNNING":     2,
}

func (x ControllerStatus) String() string {
	return proto.EnumName(ControllerStatus_name, int32(x))
}

func (ControllerStatus) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_979e75fb0350f746, []int{0}
}

func init() {
	proto.RegisterEnum("controller.exec.ControllerStatus", ControllerStatus_name, ControllerStatus_value)
}

func init() {
	proto.RegisterFile("github.com/aperturerobotics/controllerbus/controller/exec/exec.proto", fileDescriptor_979e75fb0350f746)
}

var fileDescriptor_979e75fb0350f746 = []byte{
	// 150 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x72, 0x49, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0x2c, 0x48, 0x2d, 0x2a, 0x29, 0x2d, 0x4a, 0x2d,
	0xca, 0x4f, 0xca, 0x2f, 0xc9, 0x4c, 0x2e, 0xd6, 0x4f, 0xce, 0xcf, 0x2b, 0x29, 0xca, 0xcf, 0xc9,
	0x49, 0x2d, 0x4a, 0x2a, 0x45, 0xe6, 0xe9, 0xa7, 0x56, 0xa4, 0x26, 0x83, 0x09, 0xbd, 0x82, 0xa2,
	0xfc, 0x92, 0x7c, 0x21, 0x7e, 0x84, 0x9c, 0x1e, 0x48, 0x58, 0xab, 0x80, 0x4b, 0xc0, 0x19, 0x2e,
	0x14, 0x5c, 0x92, 0x58, 0x52, 0x5a, 0x2c, 0x24, 0xc3, 0x25, 0x81, 0x2e, 0x16, 0x1f, 0xea, 0xe7,
	0xed, 0xe7, 0x1f, 0xee, 0x27, 0xc0, 0x20, 0xa4, 0xc0, 0x25, 0x83, 0x21, 0xeb, 0xec, 0xef, 0xe7,
	0xe6, 0xe9, 0x1e, 0x1a, 0xe4, 0xe9, 0xe7, 0x2e, 0xc0, 0x88, 0x55, 0x7f, 0x50, 0xa8, 0x9f, 0x1f,
	0x48, 0x96, 0x29, 0x89, 0x0d, 0xec, 0x12, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0x7e, 0xba,
	0x55, 0x26, 0xd1, 0x00, 0x00, 0x00,
}
