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

package protoyaml

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/bufbuild/protoyaml-go/internal/gen/proto/bufext/cel/expr/conformance/proto3"
	"github.com/bufbuild/protoyaml-go/internal/protoyamltest"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
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
		testCase := testCase
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
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

func TestDuration(t *testing.T) {
	val := durationpb.Duration{
		Seconds: 3600,
	}
	data, err := Marshal(&val)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "seconds: 3600\n" {
		t.Fatalf("Expected seconds: 3600, got %s", string(data))
	}
	actual := &durationpb.Duration{}
	if err := Unmarshal(data, actual); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(&val, actual) {
		t.Fatalf("Expected %v, got %v", &val, actual)
	}

}

func TestCombinatorial(t *testing.T) {
	t.Parallel()
	cases := protoyamltest.InterestingTestValues()
	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			testRoundTrip(t, c)
		})
	}
}

func TestFuzz(t *testing.T) {
	t.Parallel()
	for i := int64(0); i < 100; i++ {
		i := i
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
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
			cmp.Diff(msg, roundTrip, protocmp.Transform(),
				cmp.Comparer(func(x, y float32) bool {
					return fmt.Sprintf("%f", x) == fmt.Sprintf("%f", y)
				}),
				cmp.Comparer(func(x, y float64) bool {
					return fmt.Sprintf("%f", x) == fmt.Sprintf("%f", y)
				}),
			)
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
			UseProtoNames:   true,
			UseEnumNumbers:  true,
			EmitUnpopulated: true,
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
	cmp.Diff(testCase, roundTrip, protocmp.Transform(),
		cmp.Comparer(func(x, y float32) bool {
			return fmt.Sprintf("%f", x) == fmt.Sprintf("%f", y)
		}),
		cmp.Comparer(func(x, y float64) bool {
			return fmt.Sprintf("%f", x) == fmt.Sprintf("%f", y)
		}))
}
