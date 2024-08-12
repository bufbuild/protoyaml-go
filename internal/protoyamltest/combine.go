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

package protoyamltest

import (
	"math"

	"github.com/bufbuild/protoyaml-go/internal/gen/proto/bufext/cel/expr/conformance/proto3"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// InterestingTestValues returns a list of interesting values for testing.
//
// For example extrema, zero values, and other values that exercise edge cases.
func InterestingTestValues() []*proto3.TestAllTypes {
	var interestingValues []*proto3.TestAllTypes
	for _, value := range []bool{true, false} {
		wrapped := &wrapperspb.BoolValue{Value: value}
		anyBool, err := anypb.New(wrapped)
		if err != nil {
			panic(err)
		}
		interestingValues = append(interestingValues,
			&proto3.TestAllTypes{
				SingleBool: value,
			},
			&proto3.TestAllTypes{
				RepeatedBool: []bool{value},
			},
			&proto3.TestAllTypes{
				SingleBoolWrapper: wrapped,
			},
			&proto3.TestAllTypes{
				RepeatedBool: []bool{value},
			},
			&proto3.TestAllTypes{
				RepeatedBoolWrapper: []*wrapperspb.BoolValue{wrapped},
			},
			&proto3.TestAllTypes{
				MapBoolBool: map[bool]bool{value: value},
			},
			&proto3.TestAllTypes{
				SingleValue: &structpb.Value{Kind: &structpb.Value_BoolValue{BoolValue: value}},
			},
			&proto3.TestAllTypes{
				SingleAny: anyBool,
			},
			&proto3.TestAllTypes{
				RepeatedAny: []*anypb.Any{anyBool},
			},
		)
	}
	fields := (&proto3.TestAllTypes{}).ProtoReflect().Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		interestingValues = append(interestingValues,
			interestingFieldValues(fields.Get(i))...)
	}
	return interestingValues
}

func interestingMessageFieldValues(field protoreflect.FieldDescriptor) []*proto3.TestAllTypes {
	var interestingValues []*proto3.TestAllTypes
	switch {
	case field.IsList():
		listVal := &proto3.TestAllTypes{}
		for j := 0; j < 3; j++ {
			newVal := listVal.ProtoReflect().Get(field).List().NewElement()
			PopulateMessage(newVal.Message().Interface(), int64(j))
			listVal.ProtoReflect().Mutable(field).List().Append(newVal)
		}
		interestingValues = append(interestingValues, listVal)
	case field.IsMap():
		// TODO: populate map
	default:
		newVal := &proto3.TestAllTypes{}
		PopulateMessage(newVal.ProtoReflect().Mutable(field).Message().Interface(), 0)
		interestingValues = append(interestingValues, newVal)
	}
	return interestingValues
}

func interestingEnumFieldValues(field protoreflect.FieldDescriptor) []*proto3.TestAllTypes {
	var interestingValues []*proto3.TestAllTypes
	values := interestingEnumValues(field.Enum())
	if field.IsList() {
		listVal := &proto3.TestAllTypes{}
		for _, value := range values {
			listVal.ProtoReflect().Mutable(field).List().Append(protoreflect.ValueOfEnum(value))
		}
		interestingValues = append(interestingValues, listVal)
	} else {
		for _, value := range values {
			newVal := &proto3.TestAllTypes{}
			newVal.ProtoReflect().Set(field, protoreflect.ValueOfEnum(value))
			interestingValues = append(interestingValues, newVal)
		}
	}
	return interestingValues
}

func interestingI32FieldValues(field protoreflect.FieldDescriptor) []*proto3.TestAllTypes {
	var interestingValues []*proto3.TestAllTypes
	values := interestingIntegers(32)
	if field.IsList() {
		listVal := &proto3.TestAllTypes{}
		for _, value := range values {
			listVal.ProtoReflect().Mutable(field).List().Append(protoreflect.ValueOfInt32(int32(value)))
		}
		interestingValues = append(interestingValues, listVal)
	} else {
		for _, value := range values {
			newVal := &proto3.TestAllTypes{}
			newVal.ProtoReflect().Set(field, protoreflect.ValueOfInt32(int32(value)))
			interestingValues = append(interestingValues, newVal)
		}
	}
	return interestingValues
}

func interestingI64FieldValues(field protoreflect.FieldDescriptor) []*proto3.TestAllTypes {
	var interestingValues []*proto3.TestAllTypes
	values := interestingIntegers(64)
	if field.IsList() {
		listVal := &proto3.TestAllTypes{}
		for _, value := range values {
			listVal.ProtoReflect().Mutable(field).List().Append(protoreflect.ValueOfInt64(value))
		}
		interestingValues = append(interestingValues, listVal)
	} else {
		for _, value := range values {
			newVal := &proto3.TestAllTypes{}
			newVal.ProtoReflect().Set(field, protoreflect.ValueOfInt64(value))
			interestingValues = append(interestingValues, newVal)
		}
	}
	return interestingValues
}

func interestingU32FieldValues(field protoreflect.FieldDescriptor) []*proto3.TestAllTypes {
	var interestingValues []*proto3.TestAllTypes
	values := interestingUnsigned(32)
	if field.IsList() {
		listVal := &proto3.TestAllTypes{}
		for _, value := range values {
			listVal.ProtoReflect().Mutable(field).List().Append(protoreflect.ValueOfUint32(uint32(value)))
		}
		interestingValues = append(interestingValues, listVal)
	} else {
		for _, value := range values {
			newVal := &proto3.TestAllTypes{}
			newVal.ProtoReflect().Set(field, protoreflect.ValueOfUint32(uint32(value)))
			interestingValues = append(interestingValues, newVal)
		}
	}
	return interestingValues
}

func interestingU64FieldValues(field protoreflect.FieldDescriptor) []*proto3.TestAllTypes {
	var interestingValues []*proto3.TestAllTypes
	values := interestingUnsigned(64)
	if field.IsList() {
		listVal := &proto3.TestAllTypes{}
		for _, value := range values {
			listVal.ProtoReflect().Mutable(field).List().Append(protoreflect.ValueOfUint64(value))
		}
		interestingValues = append(interestingValues, listVal)
	} else {
		for _, value := range values {
			newVal := &proto3.TestAllTypes{}
			newVal.ProtoReflect().Set(field, protoreflect.ValueOfUint64(value))
			interestingValues = append(interestingValues, newVal)
		}
	}
	return interestingValues
}

func interestingFloatFieldValues(field protoreflect.FieldDescriptor) []*proto3.TestAllTypes {
	var interestingValues []*proto3.TestAllTypes
	values := interestingFloats(32)
	if field.IsList() {
		listVal := &proto3.TestAllTypes{}
		for _, value := range values {
			listVal.ProtoReflect().Mutable(field).List().Append(protoreflect.ValueOfFloat32(float32(value)))
		}
		interestingValues = append(interestingValues, listVal)
	} else {
		for _, value := range values {
			newVal := &proto3.TestAllTypes{}
			newVal.ProtoReflect().Set(field, protoreflect.ValueOfFloat32(float32(value)))
			interestingValues = append(interestingValues, newVal)
		}
	}
	return interestingValues
}

func interestingDoubleFieldValues(field protoreflect.FieldDescriptor) []*proto3.TestAllTypes {
	var interestingValues []*proto3.TestAllTypes
	values := interestingFloats(64)
	if field.IsList() {
		listVal := &proto3.TestAllTypes{}
		for _, value := range values {
			listVal.ProtoReflect().Mutable(field).List().Append(protoreflect.ValueOfFloat64(value))
		}
		interestingValues = append(interestingValues, listVal)
	} else {
		for _, value := range values {
			newVal := &proto3.TestAllTypes{}
			newVal.ProtoReflect().Set(field, protoreflect.ValueOfFloat64(value))
			interestingValues = append(interestingValues, newVal)
		}
	}
	return interestingValues
}

func interestingStringFieldValues(field protoreflect.FieldDescriptor) []*proto3.TestAllTypes {
	var interestingValues []*proto3.TestAllTypes
	values := interestingStrings()
	if field.IsList() {
		listVal := &proto3.TestAllTypes{}
		for _, value := range values {
			listVal.ProtoReflect().Mutable(field).List().Append(protoreflect.ValueOfString(value))
		}
		interestingValues = append(interestingValues, listVal)
	} else {
		for _, value := range values {
			newVal := &proto3.TestAllTypes{}
			newVal.ProtoReflect().Set(field, protoreflect.ValueOfString(value))
			interestingValues = append(interestingValues, newVal)
		}
	}
	return interestingValues
}

func interestingBytesFieldValues(field protoreflect.FieldDescriptor) []*proto3.TestAllTypes {
	var interestingValues []*proto3.TestAllTypes
	values := interestingBytes()
	if field.IsList() {
		listVal := &proto3.TestAllTypes{}
		for _, value := range values {
			listVal.ProtoReflect().Mutable(field).List().Append(protoreflect.ValueOfBytes(value))
		}
		interestingValues = append(interestingValues, listVal)
	} else {
		for _, value := range values {
			newVal := &proto3.TestAllTypes{}
			newVal.ProtoReflect().Set(field, protoreflect.ValueOfBytes(value))
			interestingValues = append(interestingValues, newVal)
		}
	}
	return interestingValues
}

func interestingFieldValues(field protoreflect.FieldDescriptor) []*proto3.TestAllTypes {
	switch field.Kind() {
	case protoreflect.MessageKind:
		return interestingMessageFieldValues(field)
	case protoreflect.EnumKind:
		return interestingEnumFieldValues(field)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return interestingI32FieldValues(field)
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return interestingI64FieldValues(field)
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return interestingU32FieldValues(field)
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return interestingU64FieldValues(field)
	case protoreflect.FloatKind:
		return interestingFloatFieldValues(field)
	case protoreflect.DoubleKind:
		return interestingDoubleFieldValues(field)
	case protoreflect.StringKind:
		return interestingStringFieldValues(field)
	case protoreflect.BytesKind:
		return interestingBytesFieldValues(field)
	}
	return nil
}

func interestingEnumValues(enum protoreflect.EnumDescriptor) []protoreflect.EnumNumber {
	values := enum.Values()
	result := []protoreflect.EnumNumber{}
	for i := 0; i < values.Len(); i++ {
		result = append(result, values.Get(i).Number())
	}
	if enum.FullName() != "google.protobuf.NullValue" {
		result = append(result, 0, -1, math.MaxInt32, math.MinInt32)
	}
	return result
}

func interestingIntegers(bits int) []int64 {
	maxVal := int64(1<<(bits-1) - 1)
	minVal := int64(-1 << (bits - 1))
	result := []int64{
		0,
		1,
		-1,
		maxVal,
		minVal,
	}
	if bits > 53 {
		result = append(result, 1<<53-1, 1<<53, 1<<53+1)
	}
	return result
}

func interestingUnsigned(bits int) []uint64 {
	result := []uint64{
		0,
		1,
		1<<bits - 1,
	}
	if bits > 53 {
		result = append(result, 1<<53-1, 1<<53, 1<<53+1)
	}
	return result
}

func interestingFloats(bits uint8) []float64 {
	result := []float64{}

	// Zeros
	result = append(result, 1/math.Inf(1), -1/math.Inf(1))
	// NaN
	result = append(result, math.NaN())

	// Ones
	result = append(result, 1.0, -1.0)
	// Fractions
	result = append(result, 0.5, -0.5, 0.25, -0.25, 0.125, -0.125)
	// Infinities
	result = append(result, math.Inf(1), math.Inf(-1))
	switch bits {
	case 32:
		// Smallest positive subnormal
		result = append(result, float64(math.Float32frombits(0x00000001)))
		// Largest subnormal
		result = append(result, float64(math.Float32frombits(0x007fffff)))
		// Smallest positive normal
		result = append(result, float64(math.Float32frombits(0x00800000)))
		// Largest normal
		result = append(result, float64(math.Float32frombits(0x7f7fffff)))
	case 64:
		// Smallest positive subnormal
		result = append(result, math.Float64frombits(0x0000000000000001))
		// Largest subnormal
		result = append(result, math.Float64frombits(0x000fffffffffffff))
		// Smallest positive normal
		result = append(result, math.Float64frombits(0x0010000000000000))
		// Largest normal
		result = append(result, math.Float64frombits(0x7fefffffffffffff))
		// Max safe integer
		result = append(result, math.Float64frombits(0x1fffffffffffff))
	default:
		panic("unknown float size")
	}
	return result
}

func interestingStrings() []string {
	return []string{
		// Empty
		"",
		// Whitespace
		" ",
		// "\n", TODO: Uncomment once https://github.com/go-yaml/yaml/issues/1004 is fixed
		"\t",
		"\r",
		// Nonprintable
		"\x00",
		"\x01",
		"\a",
		"\b",
		"\f",
		// Escaped
		"\\",
		"\\\\",
		"\\\"",
		"\\'",
		"\\a",
		"\\b",
		"\\f",
		// Ascii
		"hello",
		// Unicode
		"你好",
		"こんにちは",
		"☺☹",
		"😀😁",
	}
}

func interestingBytes() [][]byte {
	// All the interesting strings
	result := [][]byte{}
	for _, s := range interestingStrings() {
		result = append(result, []byte(s))
	}
	// Invalid UTF-8
	result = append(result, []byte{0xff, 0xfe, 0xfd})
	return result
}
