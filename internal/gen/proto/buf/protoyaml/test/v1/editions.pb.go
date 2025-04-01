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
// source: buf/protoyaml/test/v1/editions.proto

package testv1

import (
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

type OpenEnum int32

const (
	OpenEnum_OPEN_ENUM_UNSPECIFIED OpenEnum = 0
)

// Enum value maps for OpenEnum.
var (
	OpenEnum_name = map[int32]string{
		0: "OPEN_ENUM_UNSPECIFIED",
	}
	OpenEnum_value = map[string]int32{
		"OPEN_ENUM_UNSPECIFIED": 0,
	}
)

func (x OpenEnum) Enum() *OpenEnum {
	p := new(OpenEnum)
	*p = x
	return p
}

func (x OpenEnum) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (OpenEnum) Descriptor() protoreflect.EnumDescriptor {
	return file_buf_protoyaml_test_v1_editions_proto_enumTypes[0].Descriptor()
}

func (OpenEnum) Type() protoreflect.EnumType {
	return &file_buf_protoyaml_test_v1_editions_proto_enumTypes[0]
}

func (x OpenEnum) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use OpenEnum.Descriptor instead.
func (OpenEnum) EnumDescriptor() ([]byte, []int) {
	return file_buf_protoyaml_test_v1_editions_proto_rawDescGZIP(), []int{0}
}

type ClosedEnum int32

const (
	ClosedEnum_CLOSED_ENUM_UNSPECIFIED ClosedEnum = 0
)

// Enum value maps for ClosedEnum.
var (
	ClosedEnum_name = map[int32]string{
		0: "CLOSED_ENUM_UNSPECIFIED",
	}
	ClosedEnum_value = map[string]int32{
		"CLOSED_ENUM_UNSPECIFIED": 0,
	}
)

func (x ClosedEnum) Enum() *ClosedEnum {
	p := new(ClosedEnum)
	*p = x
	return p
}

func (x ClosedEnum) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ClosedEnum) Descriptor() protoreflect.EnumDescriptor {
	return file_buf_protoyaml_test_v1_editions_proto_enumTypes[1].Descriptor()
}

func (ClosedEnum) Type() protoreflect.EnumType {
	return &file_buf_protoyaml_test_v1_editions_proto_enumTypes[1]
}

func (x ClosedEnum) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ClosedEnum.Descriptor instead.
func (ClosedEnum) EnumDescriptor() ([]byte, []int) {
	return file_buf_protoyaml_test_v1_editions_proto_rawDescGZIP(), []int{1}
}

type EditionsTest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          *string                `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Nested        *EditionsTest_Nested   `protobuf:"group,2,opt,name=Nested,json=nested" json:"nested,omitempty"`
	Enum          OpenEnum               `protobuf:"varint,3,opt,name=enum,enum=buf.protoyaml.test.v1.OpenEnum" json:"enum,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *EditionsTest) Reset() {
	*x = EditionsTest{}
	mi := &file_buf_protoyaml_test_v1_editions_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EditionsTest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EditionsTest) ProtoMessage() {}

func (x *EditionsTest) ProtoReflect() protoreflect.Message {
	mi := &file_buf_protoyaml_test_v1_editions_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EditionsTest.ProtoReflect.Descriptor instead.
func (*EditionsTest) Descriptor() ([]byte, []int) {
	return file_buf_protoyaml_test_v1_editions_proto_rawDescGZIP(), []int{0}
}

func (x *EditionsTest) GetName() string {
	if x != nil && x.Name != nil {
		return *x.Name
	}
	return ""
}

func (x *EditionsTest) GetNested() *EditionsTest_Nested {
	if x != nil {
		return x.Nested
	}
	return nil
}

func (x *EditionsTest) GetEnum() OpenEnum {
	if x != nil {
		return x.Enum
	}
	return OpenEnum_OPEN_ENUM_UNSPECIFIED
}

type EditionsTest_Nested struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Ids           []int64                `protobuf:"varint,1,rep,name=ids" json:"ids,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *EditionsTest_Nested) Reset() {
	*x = EditionsTest_Nested{}
	mi := &file_buf_protoyaml_test_v1_editions_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EditionsTest_Nested) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EditionsTest_Nested) ProtoMessage() {}

func (x *EditionsTest_Nested) ProtoReflect() protoreflect.Message {
	mi := &file_buf_protoyaml_test_v1_editions_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EditionsTest_Nested.ProtoReflect.Descriptor instead.
func (*EditionsTest_Nested) Descriptor() ([]byte, []int) {
	return file_buf_protoyaml_test_v1_editions_proto_rawDescGZIP(), []int{0, 0}
}

func (x *EditionsTest_Nested) GetIds() []int64 {
	if x != nil {
		return x.Ids
	}
	return nil
}

var File_buf_protoyaml_test_v1_editions_proto protoreflect.FileDescriptor

const file_buf_protoyaml_test_v1_editions_proto_rawDesc = "" +
	"\n" +
	"$buf/protoyaml/test/v1/editions.proto\x12\x15buf.protoyaml.test.v1\"\xd3\x01\n" +
	"\fEditionsTest\x12\x19\n" +
	"\x04name\x18\x01 \x01(\tB\x05\xaa\x01\x02\b\x03R\x04name\x12I\n" +
	"\x06nested\x18\x02 \x01(\v2*.buf.protoyaml.test.v1.EditionsTest.NestedB\x05\xaa\x01\x02(\x02R\x06nested\x12:\n" +
	"\x04enum\x18\x03 \x01(\x0e2\x1f.buf.protoyaml.test.v1.OpenEnumB\x05\xaa\x01\x02\b\x02R\x04enum\x1a!\n" +
	"\x06Nested\x12\x17\n" +
	"\x03ids\x18\x01 \x03(\x03B\x05\xaa\x01\x02\x18\x02R\x03ids*%\n" +
	"\bOpenEnum\x12\x19\n" +
	"\x15OPEN_ENUM_UNSPECIFIED\x10\x00*/\n" +
	"\n" +
	"ClosedEnum\x12\x1b\n" +
	"\x17CLOSED_ENUM_UNSPECIFIED\x10\x00\x1a\x04:\x02\x10\x02B\xe9\x01\n" +
	"\x19com.buf.protoyaml.test.v1B\rEditionsProtoP\x01ZFbuf.build/go/protoyaml/internal/gen/proto/buf/protoyaml/test/v1;testv1\xa2\x02\x03BPT\xaa\x02\x15Buf.Protoyaml.Test.V1\xca\x02\x15Buf\\Protoyaml\\Test\\V1\xe2\x02!Buf\\Protoyaml\\Test\\V1\\GPBMetadata\xea\x02\x18Buf::Protoyaml::Test::V1b\beditionsp\xe8\a"

var (
	file_buf_protoyaml_test_v1_editions_proto_rawDescOnce sync.Once
	file_buf_protoyaml_test_v1_editions_proto_rawDescData []byte
)

func file_buf_protoyaml_test_v1_editions_proto_rawDescGZIP() []byte {
	file_buf_protoyaml_test_v1_editions_proto_rawDescOnce.Do(func() {
		file_buf_protoyaml_test_v1_editions_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_buf_protoyaml_test_v1_editions_proto_rawDesc), len(file_buf_protoyaml_test_v1_editions_proto_rawDesc)))
	})
	return file_buf_protoyaml_test_v1_editions_proto_rawDescData
}

var file_buf_protoyaml_test_v1_editions_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_buf_protoyaml_test_v1_editions_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_buf_protoyaml_test_v1_editions_proto_goTypes = []any{
	(OpenEnum)(0),               // 0: buf.protoyaml.test.v1.OpenEnum
	(ClosedEnum)(0),             // 1: buf.protoyaml.test.v1.ClosedEnum
	(*EditionsTest)(nil),        // 2: buf.protoyaml.test.v1.EditionsTest
	(*EditionsTest_Nested)(nil), // 3: buf.protoyaml.test.v1.EditionsTest.Nested
}
var file_buf_protoyaml_test_v1_editions_proto_depIdxs = []int32{
	3, // 0: buf.protoyaml.test.v1.EditionsTest.nested:type_name -> buf.protoyaml.test.v1.EditionsTest.Nested
	0, // 1: buf.protoyaml.test.v1.EditionsTest.enum:type_name -> buf.protoyaml.test.v1.OpenEnum
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_buf_protoyaml_test_v1_editions_proto_init() }
func file_buf_protoyaml_test_v1_editions_proto_init() {
	if File_buf_protoyaml_test_v1_editions_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_buf_protoyaml_test_v1_editions_proto_rawDesc), len(file_buf_protoyaml_test_v1_editions_proto_rawDesc)),
			NumEnums:      2,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_buf_protoyaml_test_v1_editions_proto_goTypes,
		DependencyIndexes: file_buf_protoyaml_test_v1_editions_proto_depIdxs,
		EnumInfos:         file_buf_protoyaml_test_v1_editions_proto_enumTypes,
		MessageInfos:      file_buf_protoyaml_test_v1_editions_proto_msgTypes,
	}.Build()
	File_buf_protoyaml_test_v1_editions_proto = out.File
	file_buf_protoyaml_test_v1_editions_proto_goTypes = nil
	file_buf_protoyaml_test_v1_editions_proto_depIdxs = nil
}
