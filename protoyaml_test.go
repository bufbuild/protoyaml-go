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

	"github.com/bufbuild/protoyaml-go/internal/gen/proto/buf/protoyaml/test/v1/proto3"
	"github.com/bufbuild/protoyaml-go/internal/protoyamltest"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestParseFieldPath(t *testing.T) {
	t.Parallel()
	for i, testCase := range []struct {
		Path   string
		Expect []string
	}{
		{"", nil},
		{"foo", []string{"foo"}},
		{"foo.bar", []string{"foo", "bar"}},
		{"foo[0]", []string{"foo", "0"}},
		{"foo[0].bar", []string{"foo", "0", "bar"}},
		{"foo[0][1]", []string{"foo", "0", "1"}},
		{"foo.0.1.bar", []string{"foo", "0", "1", "bar"}},
		{"foo.bar[0]", []string{"foo", "bar", "0"}},
		{"foo.bar[0].baz", []string{"foo", "bar", "0", "baz"}},
		{`foo["bar"].baz`, []string{"foo", "bar", "baz"}},
		{`foo["bar"].baz[0]`, []string{"foo", "bar", "baz", "0"}},
		{`foo["b\"ar"].baz[0]`, []string{"foo", "b\"ar", "baz", "0"}},
		{`foo["b.ar"].baz`, []string{"foo", "b.ar", "baz"}},
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
