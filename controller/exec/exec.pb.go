// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0-devel
// 	protoc        v4.25.1
// source: github.com/aperturerobotics/controllerbus/controller/exec/exec.proto

package controller_exec

import (
	reflect "reflect"
	sync "sync"

	controller "github.com/aperturerobotics/controllerbus/controller"
	proto "github.com/aperturerobotics/controllerbus/controller/configset/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// ControllerStatus holds basic status for a controller.
type ControllerStatus int32

const (
	// ControllerStatus_UNKNOWN is unrecognized.
	ControllerStatus_ControllerStatus_UNKNOWN ControllerStatus = 0
	// ControllerStatus_CONFIGURING indicates the controller is configuring.
	ControllerStatus_ControllerStatus_CONFIGURING ControllerStatus = 1
	// ControllerStatus_RUNNING indicates the controller is running.
	ControllerStatus_ControllerStatus_RUNNING ControllerStatus = 2
	// ControllerStatus_ERROR indicates the controller is terminated with an error.
	ControllerStatus_ControllerStatus_ERROR ControllerStatus = 3
)

// Enum value maps for ControllerStatus.
var (
	ControllerStatus_name = map[int32]string{
		0: "ControllerStatus_UNKNOWN",
		1: "ControllerStatus_CONFIGURING",
		2: "ControllerStatus_RUNNING",
		3: "ControllerStatus_ERROR",
	}
	ControllerStatus_value = map[string]int32{
		"ControllerStatus_UNKNOWN":     0,
		"ControllerStatus_CONFIGURING": 1,
		"ControllerStatus_RUNNING":     2,
		"ControllerStatus_ERROR":       3,
	}
)

func (x ControllerStatus) Enum() *ControllerStatus {
	p := new(ControllerStatus)
	*p = x
	return p
}

func (x ControllerStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ControllerStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_enumTypes[0].Descriptor()
}

func (ControllerStatus) Type() protoreflect.EnumType {
	return &file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_enumTypes[0]
}

func (x ControllerStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ControllerStatus.Descriptor instead.
func (ControllerStatus) EnumDescriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_rawDescGZIP(), []int{0}
}

// ExecControllerRequest is a protobuf request to execute a controller.
type ExecControllerRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ConfigSet is the controller config set to execute.
	ConfigSet *proto.ConfigSet `protobuf:"bytes,1,opt,name=config_set,json=configSet,proto3" json:"config_set,omitempty"`
	// ConfigSetYaml is optionally the YAML form of config_set to parse.
	// Merged with config_set.
	ConfigSetYaml string `protobuf:"bytes,2,opt,name=config_set_yaml,json=configSetYaml,proto3" json:"config_set_yaml,omitempty"`
	// ConfigSetYamlOverwrite sets if the yaml portion overwrites the proto portion.
	ConfigSetYamlOverwrite bool `protobuf:"varint,3,opt,name=config_set_yaml_overwrite,json=configSetYamlOverwrite,proto3" json:"config_set_yaml_overwrite,omitempty"`
}

func (x *ExecControllerRequest) Reset() {
	*x = ExecControllerRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecControllerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecControllerRequest) ProtoMessage() {}

func (x *ExecControllerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecControllerRequest.ProtoReflect.Descriptor instead.
func (*ExecControllerRequest) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_rawDescGZIP(), []int{0}
}

func (x *ExecControllerRequest) GetConfigSet() *proto.ConfigSet {
	if x != nil {
		return x.ConfigSet
	}
	return nil
}

func (x *ExecControllerRequest) GetConfigSetYaml() string {
	if x != nil {
		return x.ConfigSetYaml
	}
	return ""
}

func (x *ExecControllerRequest) GetConfigSetYamlOverwrite() bool {
	if x != nil {
		return x.ConfigSetYamlOverwrite
	}
	return false
}

// ExecControllerResponse is a protobuf response stream.
type ExecControllerResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Id is the configset identifier for this status report.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Status is the controller execution status.
	Status ControllerStatus `protobuf:"varint,2,opt,name=status,proto3,enum=controller.exec.ControllerStatus" json:"status,omitempty"`
	// ControllerInfo may contain the running controller info.
	ControllerInfo *controller.Info `protobuf:"bytes,3,opt,name=controller_info,json=controllerInfo,proto3" json:"controller_info,omitempty"`
	// ErrorInfo may contain the error information.
	ErrorInfo string `protobuf:"bytes,4,opt,name=error_info,json=errorInfo,proto3" json:"error_info,omitempty"`
}

func (x *ExecControllerResponse) Reset() {
	*x = ExecControllerResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecControllerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecControllerResponse) ProtoMessage() {}

func (x *ExecControllerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecControllerResponse.ProtoReflect.Descriptor instead.
func (*ExecControllerResponse) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_rawDescGZIP(), []int{1}
}

func (x *ExecControllerResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ExecControllerResponse) GetStatus() ControllerStatus {
	if x != nil {
		return x.Status
	}
	return ControllerStatus_ControllerStatus_UNKNOWN
}

func (x *ExecControllerResponse) GetControllerInfo() *controller.Info {
	if x != nil {
		return x.ControllerInfo
	}
	return nil
}

func (x *ExecControllerResponse) GetErrorInfo() string {
	if x != nil {
		return x.ErrorInfo
	}
	return ""
}

var File_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_rawDesc = []byte{
	0x0a, 0x44, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x63, 0x6f,
	0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x62, 0x75, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74,
	0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x65, 0x78, 0x65, 0x63, 0x2f, 0x65, 0x78, 0x65, 0x63,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c,
	0x65, 0x72, 0x2e, 0x65, 0x78, 0x65, 0x63, 0x1a, 0x45, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65, 0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f,
	0x74, 0x69, 0x63, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x62,
	0x75, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x63, 0x6f,
	0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x54,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65, 0x72, 0x74,
	0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74,
	0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x62, 0x75, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f,
	0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x65, 0x74, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x65, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb5, 0x01, 0x0a, 0x15, 0x45, 0x78, 0x65, 0x63, 0x43, 0x6f, 0x6e,
	0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x39,
	0x0a, 0x0a, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x5f, 0x73, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x65, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x53, 0x65, 0x74, 0x52, 0x09,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x53, 0x65, 0x74, 0x12, 0x26, 0x0a, 0x0f, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x5f, 0x73, 0x65, 0x74, 0x5f, 0x79, 0x61, 0x6d, 0x6c, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0d, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x53, 0x65, 0x74, 0x59, 0x61, 0x6d,
	0x6c, 0x12, 0x39, 0x0a, 0x19, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x5f, 0x73, 0x65, 0x74, 0x5f,
	0x79, 0x61, 0x6d, 0x6c, 0x5f, 0x6f, 0x76, 0x65, 0x72, 0x77, 0x72, 0x69, 0x74, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x16, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x53, 0x65, 0x74, 0x59,
	0x61, 0x6d, 0x6c, 0x4f, 0x76, 0x65, 0x72, 0x77, 0x72, 0x69, 0x74, 0x65, 0x22, 0xbd, 0x01, 0x0a,
	0x16, 0x45, 0x78, 0x65, 0x63, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x39, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x21, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f,
	0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x65, 0x78, 0x65, 0x63, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f,
	0x6c, 0x6c, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x39, 0x0a, 0x0f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72,
	0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x63, 0x6f,
	0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0e, 0x63,
	0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1d, 0x0a,
	0x0a, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x2a, 0x8c, 0x01, 0x0a,
	0x10, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x12, 0x1c, 0x0a, 0x18, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12,
	0x20, 0x0a, 0x1c, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x5f, 0x43, 0x4f, 0x4e, 0x46, 0x49, 0x47, 0x55, 0x52, 0x49, 0x4e, 0x47, 0x10,
	0x01, 0x12, 0x1c, 0x0a, 0x18, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x5f, 0x52, 0x55, 0x4e, 0x4e, 0x49, 0x4e, 0x47, 0x10, 0x02, 0x12,
	0x1a, 0x0a, 0x16, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10, 0x03, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_rawDescData = file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_rawDesc
)

func file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_rawDescData
}

var file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_goTypes = []interface{}{
	(ControllerStatus)(0),          // 0: controller.exec.ControllerStatus
	(*ExecControllerRequest)(nil),  // 1: controller.exec.ExecControllerRequest
	(*ExecControllerResponse)(nil), // 2: controller.exec.ExecControllerResponse
	(*proto.ConfigSet)(nil),        // 3: configset.proto.ConfigSet
	(*controller.Info)(nil),        // 4: controller.Info
}
var file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_depIdxs = []int32{
	3, // 0: controller.exec.ExecControllerRequest.config_set:type_name -> configset.proto.ConfigSet
	0, // 1: controller.exec.ExecControllerResponse.status:type_name -> controller.exec.ControllerStatus
	4, // 2: controller.exec.ExecControllerResponse.controller_info:type_name -> controller.Info
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_init() }
func file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_init() {
	if File_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecControllerRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecControllerResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_depIdxs,
		EnumInfos:         file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_enumTypes,
		MessageInfos:      file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto = out.File
	file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_rawDesc = nil
	file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_goTypes = nil
	file_github_com_aperturerobotics_controllerbus_controller_exec_exec_proto_depIdxs = nil
}
