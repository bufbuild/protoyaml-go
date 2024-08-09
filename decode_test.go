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
	"testing"

	testv1 "github.com/bufbuild/protoyaml-go/internal/gen/proto/buf/protoyaml/test/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

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
