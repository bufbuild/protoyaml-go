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
	"math"
	"testing"

	"github.com/bufbuild/protoyaml-go/internal/gen/proto/bufext/cel/expr/conformance/proto3"
)

func TestFloatJsonEncoding(t *testing.T) {
	t.Parallel()
	ival := &proto3.TestAllTypes{
		// inf, -inf, and nan are not valid JSON values, so we expect them to be encoded as strings.
		RepeatedDouble: []float64{math.Inf(1), math.Inf(-1), math.NaN()},
	}
	// Encode the message as YAML
	data, err := Marshal(ival)
	if err != nil {
		t.Fatal(err)
	}

	// Decode the message from Yaml
	oval := &proto3.TestAllTypes{}
	if err := Unmarshal(data, oval); err != nil {
		t.Fatal(err)
	}
	if len(oval.GetRepeatedDouble()) != 3 {
		t.Fatalf("Expected 3 values, got %d", len(oval.GetRepeatedDouble()))
	}
	if !math.IsInf(oval.GetRepeatedDouble()[0], 1) {
		t.Fatalf("Expected +Inf, got %f", oval.GetRepeatedDouble()[0])
	}
	if !math.IsInf(oval.GetRepeatedDouble()[1], -1) {
		t.Fatalf("Expected -Inf, got %f", oval.GetRepeatedDouble()[1])
	}
	if !math.IsNaN(oval.GetRepeatedDouble()[2]) {
		t.Fatalf("Expected NaN, got %f", oval.GetRepeatedDouble()[2])
	}
}

func TestStringEscaping(t *testing.T) {
	t.Parallel()
	ival := &proto3.TestAllTypes{
		RepeatedString: []string{"\a\b\f\n\r\t\v\\\"\\'\x7f"},
	}
	// Encode the message as YAML
	data, err := Marshal(ival)
	if err != nil {
		t.Fatal(err)
	}

	// Decode the message from Yaml
	oval := &proto3.TestAllTypes{}
	if err := Unmarshal(data, oval); err != nil {
		t.Fatal(err)
	}
	if len(oval.GetRepeatedString()) != 1 {
		t.Fatalf("Expected 1 value, got %d", len(oval.GetRepeatedString()))
	}
	if oval.GetRepeatedString()[0] != ival.GetRepeatedString()[0] {
		t.Fatalf("Expected %q, got %q", ival.GetRepeatedString()[0], oval.GetRepeatedString()[0])
	}
}

func TestBytesEncoding(t *testing.T) {
	t.Parallel()
	ival := &proto3.TestAllTypes{
		RepeatedBytes: [][]byte{{0x00, 0x01, 0x02}},
	}
	// Encode the message as YAML
	data, err := Marshal(ival)
	if err != nil {
		t.Fatal(err)
	}

	// Decode the message from Yaml
	oval := &proto3.TestAllTypes{}
	if err := Unmarshal(data, oval); err != nil {
		t.Fatal(err)
	}
	if len(oval.GetRepeatedBytes()) != 1 {
		t.Fatalf("Expected 1 value, got %d", len(oval.GetRepeatedBytes()))
	}
	if string(oval.GetRepeatedBytes()[0]) != string(ival.GetRepeatedBytes()[0]) {
		t.Fatalf("Expected %q, got %q", ival.GetRepeatedBytes()[0], oval.GetRepeatedBytes()[0])
	}
}

func TestEnumEncoding(t *testing.T) {
	t.Parallel()
	ival := &proto3.TestAllTypes{
		RepeatedNestedEnum: []proto3.TestAllTypes_NestedEnum{proto3.TestAllTypes_BAR, -1, 0, 1, 100},
	}
	// Encode the message as YAML
	data, err := Marshal(ival)
	if err != nil {
		t.Fatal(err)
	}

	// Decode the message from Yaml
	oval := &proto3.TestAllTypes{}
	if err := Unmarshal(data, oval); err != nil {
		t.Fatal(err)
	}

	if len(oval.GetRepeatedNestedEnum()) != 5 {
		t.Fatalf("Expected 5 values, got %d", len(oval.GetRepeatedNestedEnum()))
	}
	if oval.GetRepeatedNestedEnum()[0] != ival.GetRepeatedNestedEnum()[0] {
		t.Fatalf("Expected %v, got %v", ival.GetRepeatedNestedEnum()[0], oval.GetRepeatedNestedEnum()[0])
	}
	if oval.GetRepeatedNestedEnum()[1] != ival.GetRepeatedNestedEnum()[1] {
		t.Fatalf("Expected %v, got %v", ival.GetRepeatedNestedEnum()[1], oval.GetRepeatedNestedEnum()[1])
	}
	if oval.GetRepeatedNestedEnum()[2] != ival.GetRepeatedNestedEnum()[2] {
		t.Fatalf("Expected %v, got %v", ival.GetRepeatedNestedEnum()[2], oval.GetRepeatedNestedEnum()[2])
	}
	if oval.GetRepeatedNestedEnum()[3] != ival.GetRepeatedNestedEnum()[3] {
		t.Fatalf("Expected %v, got %v", ival.GetRepeatedNestedEnum()[3], oval.GetRepeatedNestedEnum()[3])
	}
	if oval.GetRepeatedNestedEnum()[4] != ival.GetRepeatedNestedEnum()[4] {
		t.Fatalf("Expected %v, got %v", ival.GetRepeatedNestedEnum()[4], oval.GetRepeatedNestedEnum()[4])
	}
}
