// Copyright 2023 Buf Technologies, Inc.
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
	"fmt"
	"math/rand"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const maxDepth = 6

func PopulateMessage(msg proto.Message, seed int64) {
	populateMessage(rand.New(rand.NewSource(seed)), msg, 0)
}

func populateMessage(rnd *rand.Rand, msg proto.Message, depth int) {
	// Special cases for known types
	switch msg := msg.(type) {
	case *anypb.Any:
		// TODO: Populate with a known type
		return
	case *durationpb.Duration:
		// Valid values are between -315,576,000,000 and +315,576,000,000 inclusive.
		if rnd.Intn(2) == 0 {
			msg.Seconds = rnd.Int63n(631152000000) - 315576000000
		} else {
			msg.Seconds = 0
		}
		if rnd.Intn(2) == 0 {
			// Valid values are between 0 and +999,999,999 inclusive.
			msg.Nanos = rnd.Int31n(1000000000)
		} else {
			msg.Nanos = 0
		}
		switch {
		case msg.GetSeconds() < 0:
			msg.Nanos = -msg.GetNanos()
		case msg.GetSeconds() == 0:
			if rnd.Intn(2) == 0 {
				msg.Nanos = -msg.GetNanos()
			}
		}
		return
	case *timestamppb.Timestamp:
		// seconds of UTC time since Unix epoch 1970-01-01T00:00:00Z. Must be from
		// 0001-01-01T00:00:00Z to 9999-12-31T23:59:59Z inclusive.
		msg.Seconds = rnd.Int63n(253402300800) - 62135596800

		// Valid values are between 0 and +999,999,999 inclusive.
		msg.Nanos = rnd.Int31n(1000000000)
		return
	case *structpb.Value:
		// Exactly one field must be non-empty.
		idx := rnd.Intn(msg.ProtoReflect().Descriptor().Fields().Len())
		field := msg.ProtoReflect().Descriptor().Fields().Get(idx)
		if field.Name() == "number_value" {
			// Must be finite.
			msg.Kind = &structpb.Value_NumberValue{NumberValue: rnd.Float64()}
		} else {
			populateField(rnd, field, msg, depth)
		}
		return
	}
	if depth > maxDepth {
		return
	}

	// For each field, decide whether to set it
	fields := msg.ProtoReflect().Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if field.ContainingOneof() == nil {
			populateField(rnd, field, msg, depth+1)
		}
	}

	// For each oneof, decide which field to set.
	oneofs := msg.ProtoReflect().Descriptor().Oneofs()
	for i := 0; i < oneofs.Len(); i++ {
		oneof := oneofs.Get(i)
		oneofFields := oneof.Fields()
		idx := rnd.Intn(oneofFields.Len())
		field := oneofFields.Get(idx)
		populateField(rnd, field, msg, depth)
	}
}

func populateField(rnd *rand.Rand, field protoreflect.FieldDescriptor, msg proto.Message, depth int) {
	switch {
	case field.IsList():
		populateList(rnd, field, msg.ProtoReflect().Mutable(field).List(), depth)
	case field.IsMap():
		populateMap(rnd, field, msg.ProtoReflect().Mutable(field).Map(), depth)
	case field.Kind() == protoreflect.MessageKind:
		populateMessage(rnd, msg.ProtoReflect().Mutable(field).Message().Interface(), depth)
	default:
		msg.ProtoReflect().Set(field, populateScalar(rnd, field))
	}
}

func populateList(rnd *rand.Rand, field protoreflect.FieldDescriptor, list protoreflect.List, depth int) {
	if depth > maxDepth {
		return
	}
	length := rnd.Intn(10)
	for i := 0; i < length; i++ {
		switch field.Kind() {
		case protoreflect.MessageKind:
			msg := list.NewElement()
			populateMessage(rnd, msg.Message().Interface(), depth+1)
			list.Append(msg)
		default:
			list.Append(populateScalar(rnd, field))
		}
	}
}

func populateMap(rnd *rand.Rand, field protoreflect.FieldDescriptor, mapVal protoreflect.Map, depth int) {
	if depth > maxDepth {
		return
	}
	keyField := field.MapKey()
	valueField := field.MapValue()
	switch keyField.Kind() {
	case protoreflect.BoolKind:
		switch rand.Intn(3) {
		case 0:
			populateMapValue(rnd, valueField, protoreflect.ValueOfBool(false).MapKey(), mapVal, depth+1)
		case 1:
			populateMapValue(rnd, valueField, protoreflect.ValueOfBool(true).MapKey(), mapVal, depth+1)
		case 2:
			populateMapValue(rnd, valueField, protoreflect.ValueOfBool(true).MapKey(), mapVal, depth+1)
			populateMapValue(rnd, valueField, protoreflect.ValueOfBool(false).MapKey(), mapVal, depth+1)
		}
	default:
		length := rnd.Intn(3)
		for i := 0; i < length; i++ {
			key := populateScalar(rnd, keyField)
			populateMapValue(rnd, valueField, key.MapKey(), mapVal, depth+1)
		}
	}
}

func populateMapValue(rnd *rand.Rand, field protoreflect.FieldDescriptor, mapKey protoreflect.MapKey,
	mapVal protoreflect.Map, depth int) {
	switch field.Kind() {
	case protoreflect.MessageKind:
		populateMessage(rnd, mapVal.Mutable(mapKey).Message().Interface(), depth)
	default:
		mapVal.Set(mapKey, populateScalar(rnd, field))
	}
}

func populateI32(rnd *rand.Rand) protoreflect.Value {
	if rnd.Intn(2) == 0 {
		return protoreflect.ValueOfInt32(int32(rnd.Int()))
	}
	vals := interestingIntegers(32)
	return protoreflect.ValueOfInt32(int32(vals[rnd.Intn(len(vals))]))
}

func populateI64(rnd *rand.Rand) protoreflect.Value {
	if rnd.Intn(2) == 0 {
		return protoreflect.ValueOfInt64(int64(rnd.Int()))
	}
	vals := interestingIntegers(64)
	return protoreflect.ValueOfInt64(vals[rnd.Intn(len(vals))])
}

func populateU32(rnd *rand.Rand) protoreflect.Value {
	if rnd.Intn(2) == 0 {
		return protoreflect.ValueOfUint32(rnd.Uint32())
	}
	vals := interestingUnsigned(32)
	return protoreflect.ValueOfUint32(uint32(vals[rnd.Intn(len(vals))]))
}

func populateU64(rnd *rand.Rand) protoreflect.Value {
	if rnd.Intn(2) == 0 {
		return protoreflect.ValueOfUint64(rnd.Uint64())
	}
	vals := interestingUnsigned(64)
	return protoreflect.ValueOfUint64(vals[rnd.Intn(len(vals))])
}

func populateScalar(rnd *rand.Rand, field protoreflect.FieldDescriptor) protoreflect.Value {
	switch field.Kind() {
	case protoreflect.BoolKind:
		return protoreflect.ValueOfBool(rnd.Intn(2) == 0)
	case protoreflect.EnumKind:
		return populateEnum(rnd, field)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return populateI32(rnd)
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return populateI64(rnd)
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return populateU32(rnd)
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return populateU64(rnd)
	case protoreflect.FloatKind:
		if rnd.Intn(2) == 0 {
			return protoreflect.ValueOfFloat32(rnd.Float32())
		}
		vals := interestingFloats(32)
		return protoreflect.ValueOfFloat32(float32(vals[rnd.Intn(len(vals))]))
	case protoreflect.DoubleKind:
		vals := interestingFloats(64)
		return protoreflect.ValueOfFloat64(vals[rnd.Intn(len(vals))])
	case protoreflect.StringKind:
		vals := interestingStrings()
		return protoreflect.ValueOfString(vals[rnd.Intn(len(vals))])
	case protoreflect.BytesKind:
		vals := interestingBytes()
		return protoreflect.ValueOfBytes(vals[rnd.Intn(len(vals))])
	default:
		panic(fmt.Sprintf("unknown scalar kind: %v", field.Kind()))
	}
}

func populateEnum(rnd *rand.Rand, field protoreflect.FieldDescriptor) protoreflect.Value {
	if field.Enum().FullName() == "google.protobuf.NullValue" {
		return protoreflect.ValueOfEnum(0)
	}
	values := field.Enum().Values()
	switch rand.Intn(3) {
	case 0:
		// Zero
		return protoreflect.ValueOfEnum(0)
	case 1:
		// Known value
		return protoreflect.ValueOfEnum(values.Get(rnd.Intn(values.Len())).Number())
	case 2:
		// Random value
		return protoreflect.ValueOfEnum(protoreflect.EnumNumber(rnd.Int()))
	default:
		panic("unreachable")
	}
}
