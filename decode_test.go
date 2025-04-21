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
	"strings"
	"testing"
	"time"

	testv1 "buf.build/go/protoyaml/internal/gen/proto/buf/protoyaml/test/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

type testCustomUnmarshaler struct{}

var _ CustomUnmarshaler = (*testCustomUnmarshaler)(nil)

func (t *testCustomUnmarshaler) Unmarshal(node *yaml.Node, msg proto.Message) (bool, error) {
	if node.Kind != yaml.ScalarNode {
		return false, nil
	}
	ts, ok := msg.(*timestamppb.Timestamp)
	if !ok {
		return false, nil
	}

	switch strings.ToLower(node.Value) {
	case "epoch":
		proto.Reset(ts)
		return true, nil
	case "now":
		proto.Merge(ts, timestamppb.Now())
		return true, nil
	case "tomorrow":
		proto.Merge(ts, timestamppb.New(time.Now().Add(24*time.Hour)))
		return true, nil
	case "yesterday":
		proto.Merge(ts, timestamppb.New(time.Now().Add(-24*time.Hour)))
		return true, nil
	}
	return false, nil
}

func TestCustomUnmarshal(t *testing.T) {
	t.Parallel()
	options := UnmarshalOptions{
		CustomUnmarshaler: &testCustomUnmarshaler{},
	}
	for _, testCase := range []struct {
		Input string
		Time  time.Time
	}{
		{Input: "epoch", Time: time.UnixMicro(0)},
		{Input: "now", Time: time.Now()},
		{Input: "tomorrow", Time: time.Now().Add(24 * time.Hour)},
		{Input: "yesterday", Time: time.Now().Add(-24 * time.Hour)},
		{Input: "2023-10-01T00:00:00Z", Time: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)},
	} {
		t.Run(testCase.Input, func(t *testing.T) {
			t.Parallel()
			actual := &timestamppb.Timestamp{}
			err := options.Unmarshal([]byte(testCase.Input), actual)
			require.NoError(t, err)
			delta := actual.AsTime().Sub(testCase.Time)
			assert.LessOrEqual(t, delta.Seconds(), 10.0)
		})
	}
}

func TestParseDuration(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		Input    string
		Expected *durationpb.Duration
		ErrMsg   string
	}{
		{Input: "", Expected: nil, ErrMsg: "invalid duration"},
		{Input: "-", Expected: nil, ErrMsg: "invalid duration"},
		{Input: "s", Expected: nil, ErrMsg: "invalid duration"},
		{Input: ".", Expected: nil, ErrMsg: "invalid duration"},
		{Input: "-s", Expected: nil, ErrMsg: "invalid duration"},
		{Input: ".s", Expected: nil, ErrMsg: "invalid duration"},
		{Input: "-.", Expected: nil, ErrMsg: "invalid duration"},
		{Input: "-.s", Expected: nil, ErrMsg: "invalid duration"},
		{Input: "--0s", Expected: nil, ErrMsg: "invalid duration"},
		{Input: "0y", Expected: nil, ErrMsg: "unknown unit"},
		{Input: "0so", Expected: nil, ErrMsg: "unknown unit"},
		{Input: "0os", Expected: nil, ErrMsg: "unknown unit"},
		{Input: "0s-0ms", Expected: nil, ErrMsg: "invalid duration"},
		{Input: "0.5ns", Expected: nil, ErrMsg: "fractional nanos"},
		{Input: "0.0005us", Expected: nil, ErrMsg: "fractional nanos"},
		{Input: "0.0000005μs", Expected: nil, ErrMsg: "fractional nanos"},
		{Input: "0.0000000005ms", Expected: nil, ErrMsg: "fractional nanos"},
		{Input: "9223372036854775807s", Expected: &durationpb.Duration{Seconds: 9223372036854775807}},
		{Input: "9223372036854775808s", ErrMsg: "out of range"},
		{Input: "-9223372036854775808s", Expected: &durationpb.Duration{Seconds: -9223372036854775808}},
		{Input: "-9223372036854775809s", ErrMsg: "out of range"},
		{Input: "18446744073709551615s", ErrMsg: "out of range"},
		{Input: "18446744073709551616s", ErrMsg: "overflow"},
		{Input: "0"},
		{Input: "0s"},
		{Input: "-0s"},
		{Input: "1s", Expected: &durationpb.Duration{Seconds: 1}},
		{Input: "-1s", Expected: &durationpb.Duration{Seconds: -1}},
		{Input: "1.5s", Expected: &durationpb.Duration{Seconds: 1, Nanos: 500000000}},
		{Input: "-1.5s", Expected: &durationpb.Duration{Seconds: -1, Nanos: -500000000}},
		{Input: "1.000000001s", Expected: &durationpb.Duration{Seconds: 1, Nanos: 1}},
		{Input: "1.0000000001s", ErrMsg: "fractional nanos"},
		{Input: "1.000000000s", Expected: &durationpb.Duration{Seconds: 1}},
		{Input: "1.0000000010s", Expected: &durationpb.Duration{Seconds: 1, Nanos: 1}},
		{Input: "1h", Expected: &durationpb.Duration{Seconds: 3600}},
		{Input: "1m", Expected: &durationpb.Duration{Seconds: 60}},
		{Input: "1h1m", Expected: &durationpb.Duration{Seconds: 3660}},
		{Input: "1h1m1s", Expected: &durationpb.Duration{Seconds: 3661}},
		{Input: "1h1m1.5s", Expected: &durationpb.Duration{Seconds: 3661, Nanos: 500000000}},
		{Input: "1.5h1m1.5s", Expected: &durationpb.Duration{Seconds: 5461, Nanos: 500000000}},
		{Input: "1.5h1m1.5s1.5h1m1.5s", Expected: &durationpb.Duration{Seconds: 10923}},
		{Input: "1h1m1s1ms1us1μs1µs1ns", Expected: &durationpb.Duration{Seconds: 3661, Nanos: 1003001}},
	} {
		t.Run(testCase.Input, func(t *testing.T) {
			t.Parallel()
			actual, err := ParseDuration(testCase.Input)
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

func TestExtension(t *testing.T) {
	t.Parallel()

	actual := &testv1.Proto2Test{}
	err := Unmarshal([]byte(`[buf.protoyaml.test.v1.p2t_string_ext]: hi`), actual)
	require.NoError(t, err)
	require.Equal(t, "hi", proto.GetExtension(actual, testv1.E_P2TStringExt))
}

func TestEditions(t *testing.T) {
	t.Parallel()

	expected := &testv1.EditionsTest{
		Name: proto.String("foobar"),
		Nested: &testv1.EditionsTest_Nested{
			Ids: []int64{0, 1, 1, 2, 3, 5, 8},
		},
		Enum: testv1.OpenEnum_OPEN_ENUM_UNSPECIFIED,
	}
	actual := &testv1.EditionsTest{}
	data := []byte(`
name: "foobar"
enum: OPEN_ENUM_UNSPECIFIED
nested:
  ids: [0, 1, 1, 2, 3, 5, 8]`)
	err := Unmarshal(data, actual)
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(expected, actual, protocmp.Transform()))
}

func TestRequiredFields(t *testing.T) {
	t.Parallel()

	actual := &testv1.EditionsTest{}
	err := Unmarshal([]byte(`enum: OPEN_ENUM_UNSPECIFIED`), actual)
	require.ErrorContains(t, err, "required field buf.protoyaml.test.v1.EditionsTest.name not set")

	err = UnmarshalOptions{AllowPartial: true}.Unmarshal([]byte(`enum: OPEN_ENUM_UNSPECIFIED`), actual)
	require.NoError(t, err)
	expected := &testv1.EditionsTest{
		Enum: testv1.OpenEnum_OPEN_ENUM_UNSPECIFIED,
	}
	require.Empty(t, cmp.Diff(expected, actual, protocmp.Transform()))
}

func TestDiscardUnknown(t *testing.T) {
	t.Parallel()

	data := []byte(`
unknown: hi
values:
  - oneof_string_value: hi
`)

	actual := &testv1.Proto2Test{}
	err := Unmarshal(data, actual)
	require.Error(t, err)

	err = UnmarshalOptions{
		DiscardUnknown: true,
	}.Unmarshal(data, actual)
	require.NoError(t, err)
	require.Equal(t, "hi", actual.GetValues()[0].GetOneofStringValue())
}
