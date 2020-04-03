// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/aperturerobotics/controllerbus/example/boilerplate/v1/boilerplate.proto

package boilerplate_v1

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

// Boilerplate implements the boilerplate directive.
type Boilerplate struct {
	// MessageText is the message to print with the boilerplate.
	// This is an example field.
	// The keyword "message" prevents us from using that as the field name.
	MessageText          string   `protobuf:"bytes,1,opt,name=message_text,json=messageText,proto3" json:"message_text,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Boilerplate) Reset()         { *m = Boilerplate{} }
func (m *Boilerplate) String() string { return proto.CompactTextString(m) }
func (*Boilerplate) ProtoMessage()    {}
func (*Boilerplate) Descriptor() ([]byte, []int) {
	return fileDescriptor_bf002109b6e414ab, []int{0}
}

func (m *Boilerplate) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Boilerplate.Unmarshal(m, b)
}
func (m *Boilerplate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Boilerplate.Marshal(b, m, deterministic)
}
func (m *Boilerplate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Boilerplate.Merge(m, src)
}
func (m *Boilerplate) XXX_Size() int {
	return xxx_messageInfo_Boilerplate.Size(m)
}
func (m *Boilerplate) XXX_DiscardUnknown() {
	xxx_messageInfo_Boilerplate.DiscardUnknown(m)
}

var xxx_messageInfo_Boilerplate proto.InternalMessageInfo

func (m *Boilerplate) GetMessageText() string {
	if m != nil {
		return m.MessageText
	}
	return ""
}

// BoilerplateResult implements the boilerplate directive result.
type BoilerplateResult struct {
	// PrintedLen is the final length of the printed message.
	PrintedLen           uint32   `protobuf:"varint,1,opt,name=printed_len,json=printedLen,proto3" json:"printed_len,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BoilerplateResult) Reset()         { *m = BoilerplateResult{} }
func (m *BoilerplateResult) String() string { return proto.CompactTextString(m) }
func (*BoilerplateResult) ProtoMessage()    {}
func (*BoilerplateResult) Descriptor() ([]byte, []int) {
	return fileDescriptor_bf002109b6e414ab, []int{1}
}

func (m *BoilerplateResult) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BoilerplateResult.Unmarshal(m, b)
}
func (m *BoilerplateResult) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BoilerplateResult.Marshal(b, m, deterministic)
}
func (m *BoilerplateResult) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BoilerplateResult.Merge(m, src)
}
func (m *BoilerplateResult) XXX_Size() int {
	return xxx_messageInfo_BoilerplateResult.Size(m)
}
func (m *BoilerplateResult) XXX_DiscardUnknown() {
	xxx_messageInfo_BoilerplateResult.DiscardUnknown(m)
}

var xxx_messageInfo_BoilerplateResult proto.InternalMessageInfo

func (m *BoilerplateResult) GetPrintedLen() uint32 {
	if m != nil {
		return m.PrintedLen
	}
	return 0
}

func init() {
	proto.RegisterType((*Boilerplate)(nil), "boilerplate.v1.Boilerplate")
	proto.RegisterType((*BoilerplateResult)(nil), "boilerplate.v1.BoilerplateResult")
}

func init() {
	proto.RegisterFile("github.com/aperturerobotics/controllerbus/example/boilerplate/v1/boilerplate.proto", fileDescriptor_bf002109b6e414ab)
}

var fileDescriptor_bf002109b6e414ab = []byte{
	// 174 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0xcd, 0xb1, 0x6a, 0xc3, 0x30,
	0x14, 0x85, 0x61, 0xbc, 0x14, 0x2a, 0xb7, 0x85, 0x7a, 0xea, 0xd6, 0xd6, 0x53, 0x27, 0xab, 0xa6,
	0x7d, 0x82, 0xcc, 0x99, 0x44, 0x76, 0x23, 0x39, 0x07, 0x47, 0x20, 0xeb, 0x8a, 0xab, 0x2b, 0xe3,
	0xc7, 0x0f, 0x18, 0x43, 0x3c, 0x9e, 0x0f, 0x7e, 0x8e, 0x32, 0x93, 0x97, 0x5b, 0x71, 0xdd, 0x48,
	0xb3, 0xb6, 0x09, 0x2c, 0x85, 0xc1, 0xe4, 0x48, 0xfc, 0x98, 0xf5, 0x48, 0x51, 0x98, 0x42, 0x00,
	0xbb, 0x92, 0x35, 0x56, 0x3b, 0xa7, 0x00, 0xed, 0xc8, 0x07, 0x70, 0x0a, 0x56, 0xa0, 0x97, 0xfe,
	0x38, 0xbb, 0xc4, 0x24, 0xd4, 0xbc, 0x1d, 0x69, 0xe9, 0xdb, 0x5f, 0x55, 0x9f, 0x1e, 0xd2, 0x7c,
	0xab, 0x97, 0x19, 0x39, 0xdb, 0x09, 0x83, 0x60, 0x95, 0x8f, 0xea, 0xab, 0xfa, 0x79, 0x36, 0xf5,
	0x6e, 0x17, 0xac, 0xd2, 0xfe, 0xab, 0xf7, 0x43, 0x61, 0x90, 0x4b, 0x90, 0xe6, 0x53, 0xd5, 0x89,
	0x7d, 0x14, 0x5c, 0x87, 0x80, 0xb8, 0x65, 0xaf, 0x46, 0xed, 0x74, 0x46, 0x74, 0x4f, 0xdb, 0xfd,
	0xdf, 0x3d, 0x00, 0x00, 0xff, 0xff, 0x23, 0x55, 0x40, 0x59, 0xd4, 0x00, 0x00, 0x00,
}