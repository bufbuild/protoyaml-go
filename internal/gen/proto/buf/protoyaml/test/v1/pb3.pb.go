// Copyright 2023-2024 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: buf/protoyaml/test/v1/pb3.proto

package testv1

import (
	proto3 "buf.build/go/protoyaml/internal/gen/proto/bufext/cel/expr/conformance/proto3"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Proto3Test struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Values        []*proto3.TestAllTypes `protobuf:"bytes,1,rep,name=values,proto3" json:"values,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Proto3Test) Reset() {
	*x = Proto3Test{}
	mi := &file_buf_protoyaml_test_v1_pb3_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Proto3Test) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Proto3Test) ProtoMessage() {}

func (x *Proto3Test) ProtoReflect() protoreflect.Message {
	mi := &file_buf_protoyaml_test_v1_pb3_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Proto3Test.ProtoReflect.Descriptor instead.
func (*Proto3Test) Descriptor() ([]byte, []int) {
	return file_buf_protoyaml_test_v1_pb3_proto_rawDescGZIP(), []int{0}
}

func (x *Proto3Test) GetValues() []*proto3.TestAllTypes {
	if x != nil {
		return x.Values
	}
	return nil
}

var File_buf_protoyaml_test_v1_pb3_proto protoreflect.FileDescriptor

const file_buf_protoyaml_test_v1_pb3_proto_rawDesc = "" +
	"\n" +
	"\x1fbuf/protoyaml/test/v1/pb3.proto\x12\x15buf.protoyaml.test.v1\x1a7bufext/cel/expr/conformance/proto3/test_all_types.proto\"V\n" +
	"\n" +
	"Proto3Test\x12H\n" +
	"\x06values\x18\x01 \x03(\v20.bufext.cel.expr.conformance.proto3.TestAllTypesR\x06valuesB\xe4\x01\n" +
	"\x19com.buf.protoyaml.test.v1B\bPb3ProtoP\x01ZFbuf.build/go/protoyaml/internal/gen/proto/buf/protoyaml/test/v1;testv1\xa2\x02\x03BPT\xaa\x02\x15Buf.Protoyaml.Test.V1\xca\x02\x15Buf\\Protoyaml\\Test\\V1\xe2\x02!Buf\\Protoyaml\\Test\\V1\\GPBMetadata\xea\x02\x18Buf::Protoyaml::Test::V1b\x06proto3"

var (
	file_buf_protoyaml_test_v1_pb3_proto_rawDescOnce sync.Once
	file_buf_protoyaml_test_v1_pb3_proto_rawDescData []byte
)

func file_buf_protoyaml_test_v1_pb3_proto_rawDescGZIP() []byte {
	file_buf_protoyaml_test_v1_pb3_proto_rawDescOnce.Do(func() {
		file_buf_protoyaml_test_v1_pb3_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_buf_protoyaml_test_v1_pb3_proto_rawDesc), len(file_buf_protoyaml_test_v1_pb3_proto_rawDesc)))
	})
	return file_buf_protoyaml_test_v1_pb3_proto_rawDescData
}

var file_buf_protoyaml_test_v1_pb3_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_buf_protoyaml_test_v1_pb3_proto_goTypes = []any{
	(*Proto3Test)(nil),          // 0: buf.protoyaml.test.v1.Proto3Test
	(*proto3.TestAllTypes)(nil), // 1: bufext.cel.expr.conformance.proto3.TestAllTypes
}
var file_buf_protoyaml_test_v1_pb3_proto_depIdxs = []int32{
	1, // 0: buf.protoyaml.test.v1.Proto3Test.values:type_name -> bufext.cel.expr.conformance.proto3.TestAllTypes
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_buf_protoyaml_test_v1_pb3_proto_init() }
func file_buf_protoyaml_test_v1_pb3_proto_init() {
	if File_buf_protoyaml_test_v1_pb3_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_buf_protoyaml_test_v1_pb3_proto_rawDesc), len(file_buf_protoyaml_test_v1_pb3_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_buf_protoyaml_test_v1_pb3_proto_goTypes,
		DependencyIndexes: file_buf_protoyaml_test_v1_pb3_proto_depIdxs,
		MessageInfos:      file_buf_protoyaml_test_v1_pb3_proto_msgTypes,
	}.Build()
	File_buf_protoyaml_test_v1_pb3_proto = out.File
	file_buf_protoyaml_test_v1_pb3_proto_goTypes = nil
	file_buf_protoyaml_test_v1_pb3_proto_depIdxs = nil
}
