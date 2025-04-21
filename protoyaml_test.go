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

package protoyaml

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"buf.build/go/protoyaml/internal/gen/proto/bufext/cel/expr/conformance/proto3"
	"buf.build/go/protoyaml/internal/protoyamltest"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

func TestParseFieldPath(t *testing.T) {
	t.Parallel()
	for i, testCase := range []struct {
		Path   string
		Expect []string
	}{
		{Path: "", Expect: nil},
		{Path: "foo", Expect: []string{"foo"}},
		{Path: "foo.bar", Expect: []string{"foo", "bar"}},
		{Path: "foo[0]", Expect: []string{"foo", "0"}},
		{Path: "foo[0].bar", Expect: []string{"foo", "0", "bar"}},
		{Path: "foo[0][1]", Expect: []string{"foo", "0", "1"}},
		{Path: "foo.0.1.bar", Expect: []string{"foo", "0", "1", "bar"}},
		{Path: "foo.bar[0]", Expect: []string{"foo", "bar", "0"}},
		{Path: "foo.bar[0].baz", Expect: []string{"foo", "bar", "0", "baz"}},
		{Path: `foo["bar"].baz`, Expect: []string{"foo", "bar", "baz"}},
		{Path: `foo["bar"].baz[0]`, Expect: []string{"foo", "bar", "baz", "0"}},
		{Path: `foo["b\"ar"].baz[0]`, Expect: []string{"foo", "b\"ar", "baz", "0"}},
		{Path: `foo["b.ar"].baz`, Expect: []string{"foo", "b.ar", "baz"}},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			result, err := parseFieldPath(testCase.Path)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result, testCase.Expect) {
				t.Fatalf("Expected %v, got %v", testCase.Expect, result)
			}
		})
	}
}

// Ensure the value can round trip as a dynamic message.
func testDynamicRoundTrip(t *testing.T, desc protoreflect.MessageDescriptor, data []byte) {
	t.Helper()
	dynval := dynamicpb.NewMessage(desc)
	if err := Unmarshal(data, dynval); err != nil {
		t.Fatal(err)
	}
	dyndata, err := Marshal(dynval)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(dyndata) {
		t.Fatalf("Expected %s, got %s", string(data), string(dyndata))
	}
}

func TestDuration_DynamicWithNanos(t *testing.T) {
	t.Parallel()
	for _, testCase := range []*durationpb.Duration{
		{Seconds: 3600, Nanos: 1001},
		{Seconds: -3600, Nanos: -1001},
		{Seconds: 3600},
		{Seconds: -3600},
		{Nanos: 1001},
		{Nanos: -1001},
	} {
		data, err := Marshal(testCase)
		if err != nil {
			t.Fatal(err)
		}
		actual := &durationpb.Duration{}
		if err := Unmarshal(data, actual); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(testCase, actual) {
			t.Fatalf("Expected %v, got %v", testCase, actual)
		}
		testDynamicRoundTrip(t, testCase.ProtoReflect().Descriptor(), data)
	}
}

func TestTimestamp_DynamicWithNanos(t *testing.T) {
	t.Parallel()
	for _, testCase := range []*timestamppb.Timestamp{
		{Seconds: 3600, Nanos: 1001},
		{Seconds: -3600, Nanos: 1001},
	} {
		data, err := Marshal(testCase)
		if err != nil {
			t.Fatal(err)
		}
		actual := &timestamppb.Timestamp{}
		if err := Unmarshal(data, actual); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(testCase, actual) {
			t.Fatalf("Expected %v, got %v", testCase, actual)
		}
		testDynamicRoundTrip(t, testCase.ProtoReflect().Descriptor(), data)
	}
}

func testValueDynamicRoundTrip(t *testing.T, data string) {
	t.Helper()
	val := &structpb.Value{}
	testDynamicRoundTrip(t, val.ProtoReflect().Descriptor(), []byte(data))
}

func TestValue_Dynamic(t *testing.T) {
	t.Parallel()
	testValueDynamicRoundTrip(t, "null\n")
	testValueDynamicRoundTrip(t, "true\n")
	testValueDynamicRoundTrip(t, "false\n")
	testValueDynamicRoundTrip(t, "1\n")
	testValueDynamicRoundTrip(t, "\"1\"\n")
	testValueDynamicRoundTrip(t, "foo\n")
	testValueDynamicRoundTrip(t, "[]\n")
	testValueDynamicRoundTrip(t, "{}\n")
	testValueDynamicRoundTrip(t, "foo: bar\n")
	testValueDynamicRoundTrip(t, "foo: 1\n")
	testValueDynamicRoundTrip(t, "foo: \"1\"\n")
	testValueDynamicRoundTrip(t, "foo: true\n")
	testValueDynamicRoundTrip(t, "foo: false\n")
	testValueDynamicRoundTrip(t, "foo: null\n")
	testValueDynamicRoundTrip(t, "foo: []\n")
	testValueDynamicRoundTrip(t, "foo: {}\n")
	testValueDynamicRoundTrip(t, "foo:\n    - 1\n")
}

func TestListValue_Dynamic(t *testing.T) {
	t.Parallel()
	val := &structpb.ListValue{
		Values: []*structpb.Value{
			{Kind: &structpb.Value_NullValue{}},
			{Kind: &structpb.Value_NumberValue{NumberValue: 1}},
			{Kind: &structpb.Value_StringValue{StringValue: "foo"}},
		},
	}
	data, err := Marshal(val)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "- null\n- 1\n- foo\n" {
		t.Fatalf("Expected - null\n- 1\n- foo, got %s", string(data))
	}
	actual := &structpb.ListValue{}
	if err := Unmarshal(data, actual); err != nil {
		t.Fatal(err)
	}
	dynval := dynamicpb.NewMessage(val.ProtoReflect().Descriptor())
	if err := Unmarshal(data, dynval); err != nil {
		t.Fatal(err)
	}
	dyndata, err := Marshal(dynval)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(dyndata) {
		t.Fatalf("Expected %s, got %s", string(data), string(dyndata))
	}
}

// See https://github.com/go-yaml/yaml/issues/1004
//
// Update combinatorial tests to include this case again when fixed.
func TestYamlNewLineMaps(t *testing.T) {
	t.Parallel()
	value := map[string]string{
		"\n": "\n",
	}
	data, err := yaml.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	// Should be `"\n": "\n"`, but its garbled.
	if string(data) != "? |4+\n: |4+\n" {
		t.Fatalf("Expected garbled output, got %s", string(data))
	}
}

// Test that non-zero null values don't round trip.
func TestNullValueEnum(t *testing.T) {
	t.Parallel()
	data, err := Marshal(&proto3.TestAllTypes{
		RepeatedNullValue: []structpb.NullValue{
			1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "repeatedNullValue:\n    - null\n" {
		t.Fatalf("Expected %#v, got %#v", "repeatedNullValue:\n    - null\n", string(data))
	}
	actual := &proto3.TestAllTypes{}
	if err := Unmarshal(data, actual); err != nil {
		t.Fatal(err)
	}
	if len(actual.GetRepeatedNullValue()) != 1 {
		t.Fatalf("Expected 1, got %d", len(actual.GetRepeatedNullValue()))
	}
	if actual.GetRepeatedNullValue()[0] != structpb.NullValue_NULL_VALUE {
		t.Fatalf("Expected %v, got %v", structpb.NullValue_NULL_VALUE, actual.GetRepeatedNullValue()[0])
	}
}

func TestInfNanIntegers(t *testing.T) {
	t.Parallel()
	for _, testCase := range []string{
		"single_int32: inf",
		"single_int32: -inf",
		"single_int32: nan",
		"single_int64: Infinity",
		"single_int64: -Infinity",
		"single_int64: NaN",
	} {
		t.Run(testCase, func(t *testing.T) {
			t.Parallel()
			data := []byte(testCase)
			actual := &proto3.TestAllTypes{}
			err := Unmarshal(data, actual)
			require.ErrorContains(t, err, "invalid syntax")
		})
	}
}

func TestAnyValue(t *testing.T) {
	t.Parallel()
	for _, testCase := range []string{
		"{}", "1", "[1, \"hi\"]",
	} {
		t.Run(testCase, func(t *testing.T) {
			t.Parallel()
			data := []byte(`{"@type": "type.googleapis.com/google.protobuf.Value", "value": ` + testCase + `}`)
			yamlAnyVal := &anypb.Any{}
			if err := Unmarshal(data, yamlAnyVal); err != nil {
				t.Fatal(err)
			}
			jsonAnyVal := &anypb.Any{}
			if err := protojson.Unmarshal(data, jsonAnyVal); err != nil {
				t.Fatal(err)
			}
			actualYaml := &structpb.Value{}
			if err := yamlAnyVal.UnmarshalTo(actualYaml); err != nil {
				t.Fatal(err)
			}
			actualJSON := &structpb.Value{}
			if err := jsonAnyVal.UnmarshalTo(actualJSON); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(actualJSON, actualYaml, protocmp.Transform()); diff != "" {
				t.Errorf("Unexpected diff:\n%s", diff)
			}
		})
	}
}

func TestCombinatorial(t *testing.T) {
	t.Parallel()
	cases := protoyamltest.InterestingTestValues()
	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			testRoundTrip(t, c)
		})
	}
}

func TestFuzz(t *testing.T) {
	t.Parallel()
	for i := range int64(100) {
		t.Run(strconv.FormatInt(i, 10), func(t *testing.T) {
			t.Parallel()
			msg := &proto3.TestAllTypes{}
			protoyamltest.PopulateMessage(msg, i)

			data, err := Marshal(msg)
			if err != nil {
				t.Fatal(err)
			}
			roundTrip := &proto3.TestAllTypes{}
			err = Unmarshal(data, roundTrip)
			if err != nil {
				t.Fatal(err)
			}
			diff := cmp.Diff(msg, roundTrip, protocmp.Transform(),
				cmp.Comparer(func(x, y float32) bool {
					return fmt.Sprintf("%f", x) == fmt.Sprintf("%f", y)
				}),
				cmp.Comparer(func(x, y float64) bool {
					return fmt.Sprintf("%f", x) == fmt.Sprintf("%f", y)
				}),
			)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func testRoundTrip(t *testing.T, testCase *proto3.TestAllTypes) {
	t.Helper()
	t.Run("Default", func(t *testing.T) {
		testRoundTripOption(t, testCase, MarshalOptions{})
	})
	t.Run("Alt", func(t *testing.T) {
		testRoundTripOption(t, testCase, MarshalOptions{
			UseProtoNames:  true,
			UseEnumNumbers: true,
		})
	})
}

func testRoundTripOption(t *testing.T, testCase *proto3.TestAllTypes, encOpt MarshalOptions) {
	t.Helper()
	// Encode the message as YAML
	data, err := encOpt.Marshal(testCase)
	if err != nil {
		t.Fatal(err)
	}

	// Decode the message from Yaml
	testUnmarshal(t, testCase, data)

	// Encode the message as JSON

	jsonData, err := protojson.MarshalOptions{
		AllowPartial:    encOpt.AllowPartial,
		UseProtoNames:   encOpt.UseProtoNames,
		UseEnumNumbers:  encOpt.UseEnumNumbers,
		EmitUnpopulated: encOpt.EmitUnpopulated,
		Resolver:        encOpt.Resolver,
	}.Marshal(testCase)
	if err != nil {
		t.Fatal(err)
	}

	// Decode the message from JSON
	testUnmarshal(t, testCase, jsonData)
}

func testUnmarshal(t *testing.T, testCase *proto3.TestAllTypes, data []byte) {
	t.Helper()
	roundTrip := &proto3.TestAllTypes{}
	err := Unmarshal(data, roundTrip)
	if err != nil {
		t.Fatal(err)
	}

	// Compare the two messages
	diff := cmp.Diff(testCase, roundTrip, protocmp.Transform(),
		cmp.Comparer(func(x, y float32) bool {
			return fmt.Sprintf("%f", x) == fmt.Sprintf("%f", y)
		}),
		cmp.Comparer(func(x, y float64) bool {
			return fmt.Sprintf("%f", x) == fmt.Sprintf("%f", y)
		}))
	if diff != "" {
		t.Fatal(diff)
	}
}
