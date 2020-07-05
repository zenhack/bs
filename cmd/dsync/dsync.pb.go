// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.24.0
// 	protoc        v3.6.1
// source: dsync.proto

package main

import (
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type Bytes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	B []byte `protobuf:"bytes,1,opt,name=b,proto3" json:"b,omitempty"`
}

func (x *Bytes) Reset() {
	*x = Bytes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsync_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Bytes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Bytes) ProtoMessage() {}

func (x *Bytes) ProtoReflect() protoreflect.Message {
	mi := &file_dsync_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Bytes.ProtoReflect.Descriptor instead.
func (*Bytes) Descriptor() ([]byte, []int) {
	return file_dsync_proto_rawDescGZIP(), []int{0}
}

func (x *Bytes) GetB() []byte {
	if x != nil {
		return x.B
	}
	return nil
}

type AnchorTuple struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	A   string               `protobuf:"bytes,1,opt,name=a,proto3" json:"a,omitempty"`
	Ref []byte               `protobuf:"bytes,2,opt,name=ref,proto3" json:"ref,omitempty"`
	T   *timestamp.Timestamp `protobuf:"bytes,3,opt,name=t,proto3" json:"t,omitempty"`
}

func (x *AnchorTuple) Reset() {
	*x = AnchorTuple{}
	if protoimpl.UnsafeEnabled {
		mi := &file_dsync_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AnchorTuple) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AnchorTuple) ProtoMessage() {}

func (x *AnchorTuple) ProtoReflect() protoreflect.Message {
	mi := &file_dsync_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AnchorTuple.ProtoReflect.Descriptor instead.
func (*AnchorTuple) Descriptor() ([]byte, []int) {
	return file_dsync_proto_rawDescGZIP(), []int{1}
}

func (x *AnchorTuple) GetA() string {
	if x != nil {
		return x.A
	}
	return ""
}

func (x *AnchorTuple) GetRef() []byte {
	if x != nil {
		return x.Ref
	}
	return nil
}

func (x *AnchorTuple) GetT() *timestamp.Timestamp {
	if x != nil {
		return x.T
	}
	return nil
}

var File_dsync_proto protoreflect.FileDescriptor

var file_dsync_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x64, 0x73, 0x79, 0x6e, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x6d,
	0x61, 0x69, 0x6e, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x15, 0x0a, 0x05, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x0c, 0x0a, 0x01, 0x62, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x01, 0x62, 0x22, 0x57, 0x0a, 0x0b, 0x41, 0x6e, 0x63, 0x68,
	0x6f, 0x72, 0x54, 0x75, 0x70, 0x6c, 0x65, 0x12, 0x0c, 0x0a, 0x01, 0x61, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x01, 0x61, 0x12, 0x10, 0x0a, 0x03, 0x72, 0x65, 0x66, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x03, 0x72, 0x65, 0x66, 0x12, 0x28, 0x0a, 0x01, 0x74, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x01,
	0x74, 0x32, 0x9e, 0x01, 0x0a, 0x07, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x12, 0x27, 0x0a,
	0x07, 0x44, 0x6f, 0x4f, 0x66, 0x66, 0x65, 0x72, 0x12, 0x0b, 0x2e, 0x6d, 0x61, 0x69, 0x6e, 0x2e,
	0x42, 0x79, 0x74, 0x65, 0x73, 0x1a, 0x0b, 0x2e, 0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x42, 0x79, 0x74,
	0x65, 0x73, 0x28, 0x01, 0x30, 0x01, 0x12, 0x30, 0x0a, 0x07, 0x44, 0x6f, 0x42, 0x6c, 0x6f, 0x62,
	0x73, 0x12, 0x0b, 0x2e, 0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x42, 0x79, 0x74, 0x65, 0x73, 0x1a, 0x16,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x28, 0x01, 0x12, 0x38, 0x0a, 0x09, 0x44, 0x6f, 0x41, 0x6e,
	0x63, 0x68, 0x6f, 0x72, 0x73, 0x12, 0x11, 0x2e, 0x6d, 0x61, 0x69, 0x6e, 0x2e, 0x41, 0x6e, 0x63,
	0x68, 0x6f, 0x72, 0x54, 0x75, 0x70, 0x6c, 0x65, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x28, 0x01, 0x42, 0x08, 0x5a, 0x06, 0x2e, 0x3b, 0x6d, 0x61, 0x69, 0x6e, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_dsync_proto_rawDescOnce sync.Once
	file_dsync_proto_rawDescData = file_dsync_proto_rawDesc
)

func file_dsync_proto_rawDescGZIP() []byte {
	file_dsync_proto_rawDescOnce.Do(func() {
		file_dsync_proto_rawDescData = protoimpl.X.CompressGZIP(file_dsync_proto_rawDescData)
	})
	return file_dsync_proto_rawDescData
}

var file_dsync_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_dsync_proto_goTypes = []interface{}{
	(*Bytes)(nil),               // 0: main.Bytes
	(*AnchorTuple)(nil),         // 1: main.AnchorTuple
	(*timestamp.Timestamp)(nil), // 2: google.protobuf.Timestamp
	(*empty.Empty)(nil),         // 3: google.protobuf.Empty
}
var file_dsync_proto_depIdxs = []int32{
	2, // 0: main.AnchorTuple.t:type_name -> google.protobuf.Timestamp
	0, // 1: main.Replica.DoOffer:input_type -> main.Bytes
	0, // 2: main.Replica.DoBlobs:input_type -> main.Bytes
	1, // 3: main.Replica.DoAnchors:input_type -> main.AnchorTuple
	0, // 4: main.Replica.DoOffer:output_type -> main.Bytes
	3, // 5: main.Replica.DoBlobs:output_type -> google.protobuf.Empty
	3, // 6: main.Replica.DoAnchors:output_type -> google.protobuf.Empty
	4, // [4:7] is the sub-list for method output_type
	1, // [1:4] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_dsync_proto_init() }
func file_dsync_proto_init() {
	if File_dsync_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_dsync_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Bytes); i {
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
		file_dsync_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AnchorTuple); i {
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
			RawDescriptor: file_dsync_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_dsync_proto_goTypes,
		DependencyIndexes: file_dsync_proto_depIdxs,
		MessageInfos:      file_dsync_proto_msgTypes,
	}.Build()
	File_dsync_proto = out.File
	file_dsync_proto_rawDesc = nil
	file_dsync_proto_goTypes = nil
	file_dsync_proto_depIdxs = nil
}
