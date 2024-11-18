// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        (unknown)
// source: example/v1/example.proto

package examplev1

import (
	_ "git.maurice.fr/thomas/protoc-gen-go-tmprl/gen/temporal/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type HelloRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *HelloRequest) Reset() {
	*x = HelloRequest{}
	mi := &file_example_v1_example_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HelloRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HelloRequest) ProtoMessage() {}

func (x *HelloRequest) ProtoReflect() protoreflect.Message {
	mi := &file_example_v1_example_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HelloRequest.ProtoReflect.Descriptor instead.
func (*HelloRequest) Descriptor() ([]byte, []int) {
	return file_example_v1_example_proto_rawDescGZIP(), []int{0}
}

func (x *HelloRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type HelloResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data string `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *HelloResponse) Reset() {
	*x = HelloResponse{}
	mi := &file_example_v1_example_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HelloResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HelloResponse) ProtoMessage() {}

func (x *HelloResponse) ProtoReflect() protoreflect.Message {
	mi := &file_example_v1_example_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HelloResponse.ProtoReflect.Descriptor instead.
func (*HelloResponse) Descriptor() ([]byte, []int) {
	return file_example_v1_example_proto_rawDescGZIP(), []int{1}
}

func (x *HelloResponse) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

type MultipleHelloRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Names []string `protobuf:"bytes,1,rep,name=names,proto3" json:"names,omitempty"`
}

func (x *MultipleHelloRequest) Reset() {
	*x = MultipleHelloRequest{}
	mi := &file_example_v1_example_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MultipleHelloRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MultipleHelloRequest) ProtoMessage() {}

func (x *MultipleHelloRequest) ProtoReflect() protoreflect.Message {
	mi := &file_example_v1_example_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MultipleHelloRequest.ProtoReflect.Descriptor instead.
func (*MultipleHelloRequest) Descriptor() ([]byte, []int) {
	return file_example_v1_example_proto_rawDescGZIP(), []int{2}
}

func (x *MultipleHelloRequest) GetNames() []string {
	if x != nil {
		return x.Names
	}
	return nil
}

type MultipleHelloResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []string `protobuf:"bytes,1,rep,name=data,proto3" json:"data,omitempty"`
}

func (x *MultipleHelloResponse) Reset() {
	*x = MultipleHelloResponse{}
	mi := &file_example_v1_example_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MultipleHelloResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MultipleHelloResponse) ProtoMessage() {}

func (x *MultipleHelloResponse) ProtoReflect() protoreflect.Message {
	mi := &file_example_v1_example_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MultipleHelloResponse.ProtoReflect.Descriptor instead.
func (*MultipleHelloResponse) Descriptor() ([]byte, []int) {
	return file_example_v1_example_proto_rawDescGZIP(), []int{3}
}

func (x *MultipleHelloResponse) GetData() []string {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_example_v1_example_proto protoreflect.FileDescriptor

var file_example_v1_example_proto_rawDesc = []byte{
	0x0a, 0x18, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x65, 0x78, 0x61,
	0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x65, 0x78, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x1a, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c,
	0x2f, 0x76, 0x31, 0x2f, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x22, 0x0a, 0x0c, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x22, 0x23, 0x0a, 0x0d, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x2c, 0x0a, 0x14, 0x4d, 0x75, 0x6c, 0x74,
	0x69, 0x70, 0x6c, 0x65, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x14, 0x0a, 0x05, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x05, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x22, 0x2b, 0x0a, 0x15, 0x4d, 0x75, 0x6c, 0x74, 0x69, 0x70,
	0x6c, 0x65, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x32, 0x98, 0x02, 0x0a, 0x0a, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x57, 0x6f, 0x72,
	0x6c, 0x64, 0x12, 0x55, 0x0a, 0x08, 0x53, 0x61, 0x79, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x12, 0x18,
	0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x48, 0x65, 0x6c, 0x6c,
	0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70,
	0x6c, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x14, 0x82, 0xb5, 0x18, 0x10, 0x0a, 0x0e, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
	0x2e, 0x53, 0x61, 0x79, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x12, 0x3c, 0x0a, 0x04, 0x50, 0x69, 0x6e,
	0x67, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x22, 0x04, 0x82, 0xb5, 0x18, 0x00, 0x12, 0x5d, 0x0a, 0x10, 0x53, 0x61, 0x79, 0x4d, 0x75,
	0x6c, 0x74, 0x69, 0x70, 0x6c, 0x65, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x12, 0x20, 0x2e, 0x65, 0x78,
	0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x75, 0x6c, 0x74, 0x69, 0x70, 0x6c,
	0x65, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x21, 0x2e,
	0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x75, 0x6c, 0x74, 0x69,
	0x70, 0x6c, 0x65, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x04, 0x8a, 0xb5, 0x18, 0x00, 0x1a, 0x16, 0x92, 0xb5, 0x18, 0x12, 0x0a, 0x10, 0x68, 0x65,
	0x6c, 0x6c, 0x6f, 0x2d, 0x74, 0x61, 0x73, 0x6b, 0x2d, 0x71, 0x75, 0x65, 0x75, 0x65, 0x42, 0xab,
	0x01, 0x0a, 0x0e, 0x63, 0x6f, 0x6d, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x76,
	0x31, 0x42, 0x0c, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50,
	0x01, 0x5a, 0x42, 0x67, 0x69, 0x74, 0x2e, 0x6d, 0x61, 0x75, 0x72, 0x69, 0x63, 0x65, 0x2e, 0x66,
	0x72, 0x2f, 0x74, 0x68, 0x6f, 0x6d, 0x61, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d,
	0x67, 0x65, 0x6e, 0x2d, 0x67, 0x6f, 0x2d, 0x74, 0x6d, 0x70, 0x72, 0x6c, 0x2f, 0x67, 0x65, 0x6e,
	0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x3b, 0x65, 0x78, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x45, 0x58, 0x58, 0xaa, 0x02, 0x0a, 0x45, 0x78,
	0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x0a, 0x45, 0x78, 0x61, 0x6d, 0x70,
	0x6c, 0x65, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x16, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x5c,
	0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02,
	0x0b, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_example_v1_example_proto_rawDescOnce sync.Once
	file_example_v1_example_proto_rawDescData = file_example_v1_example_proto_rawDesc
)

func file_example_v1_example_proto_rawDescGZIP() []byte {
	file_example_v1_example_proto_rawDescOnce.Do(func() {
		file_example_v1_example_proto_rawDescData = protoimpl.X.CompressGZIP(file_example_v1_example_proto_rawDescData)
	})
	return file_example_v1_example_proto_rawDescData
}

var file_example_v1_example_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_example_v1_example_proto_goTypes = []any{
	(*HelloRequest)(nil),          // 0: example.v1.HelloRequest
	(*HelloResponse)(nil),         // 1: example.v1.HelloResponse
	(*MultipleHelloRequest)(nil),  // 2: example.v1.MultipleHelloRequest
	(*MultipleHelloResponse)(nil), // 3: example.v1.MultipleHelloResponse
	(*emptypb.Empty)(nil),         // 4: google.protobuf.Empty
}
var file_example_v1_example_proto_depIdxs = []int32{
	0, // 0: example.v1.HelloWorld.SayHello:input_type -> example.v1.HelloRequest
	4, // 1: example.v1.HelloWorld.Ping:input_type -> google.protobuf.Empty
	2, // 2: example.v1.HelloWorld.SayMultipleHello:input_type -> example.v1.MultipleHelloRequest
	1, // 3: example.v1.HelloWorld.SayHello:output_type -> example.v1.HelloResponse
	4, // 4: example.v1.HelloWorld.Ping:output_type -> google.protobuf.Empty
	3, // 5: example.v1.HelloWorld.SayMultipleHello:output_type -> example.v1.MultipleHelloResponse
	3, // [3:6] is the sub-list for method output_type
	0, // [0:3] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_example_v1_example_proto_init() }
func file_example_v1_example_proto_init() {
	if File_example_v1_example_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_example_v1_example_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_example_v1_example_proto_goTypes,
		DependencyIndexes: file_example_v1_example_proto_depIdxs,
		MessageInfos:      file_example_v1_example_proto_msgTypes,
	}.Build()
	File_example_v1_example_proto = out.File
	file_example_v1_example_proto_rawDesc = nil
	file_example_v1_example_proto_goTypes = nil
	file_example_v1_example_proto_depIdxs = nil
}
