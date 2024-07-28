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
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bufbuild/protovalidate-go"
	testv1 "github.com/bufbuild/protoyaml-go/internal/gen/proto/buf/protoyaml/test/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
	"gotest.tools/v3/assert"
)

func TestGoldenFiles(t *testing.T) {
	t.Parallel()
	// Walk the test data directory for .yaml files
	if err := filepath.Walk("internal/testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".yaml") {
			t.Run(path, func(t *testing.T) {
				t.Parallel()
				testRunYAMLFile(t, path)
			})
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestParseDuration(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		Input    string
		Expected *durationpb.Duration
	}{
		{Input: "", Expected: nil},
		{Input: "-", Expected: nil},
		{Input: "-s", Expected: nil},
		{Input: "0s", Expected: &durationpb.Duration{}},
		{Input: "-0s", Expected: &durationpb.Duration{}},
		{Input: "1s", Expected: &durationpb.Duration{Seconds: 1}},
		{Input: "-1s", Expected: &durationpb.Duration{Seconds: -1}},
		{Input: "--1s", Expected: nil},
		{Input: "1.5s", Expected: &durationpb.Duration{Seconds: 1, Nanos: 500000000}},
		{Input: "-1.5s", Expected: &durationpb.Duration{Seconds: -1, Nanos: -500000000}},
		{Input: "1.000000001s", Expected: &durationpb.Duration{Seconds: 1, Nanos: 1}},
		{Input: "1.0000000001s", Expected: nil},
		{Input: "1.000000000s", Expected: &durationpb.Duration{Seconds: 1}},
		{Input: "1.0000000010s", Expected: nil},
		{Input: "-1.000000001s", Expected: &durationpb.Duration{Seconds: -1, Nanos: -1}},
		{Input: "0s", Expected: &durationpb.Duration{}},
		{Input: "-0s", Expected: &durationpb.Duration{}},
		{Input: "0.1s", Expected: &durationpb.Duration{Nanos: 100000000}},
		{Input: "-0.1s", Expected: &durationpb.Duration{Nanos: -100000000}},
		{Input: "0.000000001s", Expected: &durationpb.Duration{Nanos: 1}},
		{Input: "0.0000000001s", Expected: nil},
		{Input: "0.000000000s", Expected: &durationpb.Duration{}},
		{Input: "0.0000000010s", Expected: nil},
		{Input: "-0.000000001s", Expected: &durationpb.Duration{Nanos: -1}},
	} {
		testCase := testCase
		t.Run(testCase.Input, func(t *testing.T) {
			t.Parallel()
			actual := &durationpb.Duration{}
			err := parseDuration(testCase.Input, actual)
			if testCase.Expected == nil {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(testCase.Expected, actual, protocmp.Transform()); diff != "" {
					t.Errorf("Unexpected diff:\n%s", diff)
				}
			}
		})
	}
}

func TestExtension(t *testing.T) {
	t.Parallel()

	actual := &testv1.Proto2Test{}
	err := Unmarshal([]byte(`[buf.protoyaml.test.v1.p2t_string_ext]: hi`), actual)
	require.NoError(t, err)
	require.Equal(t, "hi", proto.GetExtension(actual, testv1.E_P2TStringExt))
}

func testRunYAML(path string, msg proto.Message) error {
	// Read the test file
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	validator, err := protovalidate.New()
	if err != nil {
		return err
	}

	return UnmarshalOptions{
		Validator: validator,
		Path:      path,
	}.Unmarshal(data, msg)
}

func testRunYAMLFile(t *testing.T, testFile string) {
	t.Helper()

	var err error
	switch {
	case strings.HasSuffix(testFile, ".proto2test.yaml"):
		err = testRunYAML(testFile, &testv1.Proto2Test{})
	case strings.HasSuffix(testFile, ".proto3test.yaml"):
		err = testRunYAML(testFile, &testv1.Proto3Test{})
	case strings.HasSuffix(testFile, ".const.yaml"):
		err = testRunYAML(testFile, &testv1.ConstValues{})
	case strings.HasSuffix(testFile, ".validate.yaml"):
		err = testRunYAML(testFile, &testv1.ValidateTest{})
	default:
		t.Fatalf("Unknown test file extension: %s", testFile)
	}
	var errorText string
	if err != nil {
		errorText = err.Error()
	}
	// Read the expected file
	expectedFileName := strings.TrimSuffix(testFile, ".yaml") + ".txt"
	expectedFile, err := os.Open(expectedFileName)
	var expectedData []byte
	if err != nil {
		t.Fatal(err)
	} else {
		defer expectedFile.Close()
		expectedData, err = io.ReadAll(expectedFile)
		if err != nil {
			t.Fatal(err)
		}
	}
	expectedText := string(expectedData)
	if expectedText != errorText {
		diff := cmp.Diff(expectedText, errorText)
		t.Errorf("%s: Test %s failed:\nExpected:\n%s\nActual:\n%s\nDiff:\n%s", expectedFileName, testFile, expectedText, errorText, diff)
	}
}

func TestToDuration(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		Literal  string
		Expected *durationpb.Duration
		ErrMsg   string
	}{
		{Literal: "", Expected: nil, ErrMsg: "invalid duration"},
		{Literal: "-", Expected: nil, ErrMsg: "invalid duration"},
		{Literal: "s", Expected: nil, ErrMsg: "invalid duration"},
		{Literal: ".", Expected: nil, ErrMsg: "invalid duration"},
		{Literal: "-s", Expected: nil, ErrMsg: "invalid duration"},
		{Literal: ".s", Expected: nil, ErrMsg: "invalid duration"},
		{Literal: "-.", Expected: nil, ErrMsg: "invalid duration"},
		{Literal: "-.s", Expected: nil, ErrMsg: "invalid duration"},
		{Literal: "0y", Expected: nil, ErrMsg: "unknown unit"},
		{Literal: "0so", Expected: nil, ErrMsg: "unknown unit"},
		{Literal: "0os", Expected: nil, ErrMsg: "unknown unit"},
		{Literal: "0s-0ms", Expected: nil, ErrMsg: "invalid duration"},
		{Literal: "0.5ns", Expected: nil, ErrMsg: "fractional nanos"},
		{Literal: "0.0005us", Expected: nil, ErrMsg: "fractional nanos"},
		{Literal: "0.0000005μs", Expected: nil, ErrMsg: "fractional nanos"},
		{Literal: "0.0000000005ms", Expected: nil, ErrMsg: "fractional nanos"},
		{Literal: "9223372036854775807s", Expected: &durationpb.Duration{Seconds: 9223372036854775807}},
		{Literal: "9223372036854775808s", ErrMsg: "out of range"},
		{Literal: "-9223372036854775808s", Expected: &durationpb.Duration{Seconds: -9223372036854775808}},
		{Literal: "-9223372036854775809s", ErrMsg: "out of range"},
		{Literal: "18446744073709551615s", ErrMsg: "out of range"},
		{Literal: "18446744073709551616s", ErrMsg: "overflow"},
		{Literal: "0"},
		{Literal: "0s"},
		{Literal: "-0s"},
		{Literal: "1s", Expected: &durationpb.Duration{Seconds: 1}},
		{Literal: "-1s", Expected: &durationpb.Duration{Seconds: -1}},
		{Literal: "1.5s", Expected: &durationpb.Duration{Seconds: 1, Nanos: 500000000}},
		{Literal: "-1.5s", Expected: &durationpb.Duration{Seconds: -1, Nanos: -500000000}},
		{Literal: "1.000000001s", Expected: &durationpb.Duration{Seconds: 1, Nanos: 1}},
		{Literal: "1.0000000001s", ErrMsg: "fractional nanos"},
		{Literal: "1.000000000s", Expected: &durationpb.Duration{Seconds: 1}},
		{Literal: "1.0000000010s", Expected: &durationpb.Duration{Seconds: 1, Nanos: 1}},
		{Literal: "1h", Expected: &durationpb.Duration{Seconds: 3600}},
		{Literal: "1m", Expected: &durationpb.Duration{Seconds: 60}},
		{Literal: "1h1m", Expected: &durationpb.Duration{Seconds: 3660}},
		{Literal: "1h1m1s", Expected: &durationpb.Duration{Seconds: 3661}},
		{Literal: "1h1m1.5s", Expected: &durationpb.Duration{Seconds: 3661, Nanos: 500000000}},
		{Literal: "1.5h1m1.5s", Expected: &durationpb.Duration{Seconds: 5461, Nanos: 500000000}},
		{Literal: "1.5h1m1.5s1.5h1m1.5s", Expected: &durationpb.Duration{Seconds: 10923}},
		{Literal: "1h1m1s1ms1us1μs1µs1ns", Expected: &durationpb.Duration{Seconds: 3661, Nanos: 1003001}},
	} {
		testCase := testCase
		t.Run("", func(t *testing.T) {
			t.Parallel()
			actual, err := ParseDuration(testCase.Literal)
			if testCase.ErrMsg != "" {
				require.ErrorContains(t, err, testCase.ErrMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.Expected.GetSeconds(), actual.GetSeconds())
			assert.Equal(t, testCase.Expected.GetNanos(), actual.GetNanos())
		})
	}
}
