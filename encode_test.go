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

	"github.com/bufbuild/protoyaml-go/internal/gen/proto/buf/protoyaml/test/v1/proto3"
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
	if len(oval.RepeatedDouble) != 3 {
		t.Fatalf("Expected 3 values, got %d", len(oval.RepeatedDouble))
	}
	if !math.IsInf(oval.RepeatedDouble[0], 1) {
		t.Fatalf("Expected +Inf, got %f", oval.RepeatedDouble[0])
	}
	if !math.IsInf(oval.RepeatedDouble[1], -1) {
		t.Fatalf("Expected -Inf, got %f", oval.RepeatedDouble[1])
	}
	if !math.IsNaN(oval.RepeatedDouble[2]) {
		t.Fatalf("Expected NaN, got %f", oval.RepeatedDouble[2])
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
	if len(oval.RepeatedString) != 1 {
		t.Fatalf("Expected 1 value, got %d", len(oval.RepeatedString))
	}
	if oval.RepeatedString[0] != ival.RepeatedString[0] {
		t.Fatalf("Expected %q, got %q", ival.RepeatedString[0], oval.RepeatedString[0])
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
	if len(oval.RepeatedBytes) != 1 {
		t.Fatalf("Expected 1 value, got %d", len(oval.RepeatedBytes))
	}
	if string(oval.RepeatedBytes[0]) != string(ival.RepeatedBytes[0]) {
		t.Fatalf("Expected %q, got %q", ival.RepeatedBytes[0], oval.RepeatedBytes[0])
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

	if len(oval.RepeatedNestedEnum) != 5 {
		t.Fatalf("Expected 5 values, got %d", len(oval.RepeatedNestedEnum))
	}
	if oval.RepeatedNestedEnum[0] != ival.RepeatedNestedEnum[0] {
		t.Fatalf("Expected %v, got %v", ival.RepeatedNestedEnum[0], oval.RepeatedNestedEnum[0])
	}
	if oval.RepeatedNestedEnum[1] != ival.RepeatedNestedEnum[1] {
		t.Fatalf("Expected %v, got %v", ival.RepeatedNestedEnum[1], oval.RepeatedNestedEnum[1])
	}
	if oval.RepeatedNestedEnum[2] != ival.RepeatedNestedEnum[2] {
		t.Fatalf("Expected %v, got %v", ival.RepeatedNestedEnum[2], oval.RepeatedNestedEnum[2])
	}
	if oval.RepeatedNestedEnum[3] != ival.RepeatedNestedEnum[3] {
		t.Fatalf("Expected %v, got %v", ival.RepeatedNestedEnum[3], oval.RepeatedNestedEnum[3])
	}
	if oval.RepeatedNestedEnum[4] != ival.RepeatedNestedEnum[4] {
		t.Fatalf("Expected %v, got %v", ival.RepeatedNestedEnum[4], oval.RepeatedNestedEnum[4])
	}
}
