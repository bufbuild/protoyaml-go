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

package golden

import (
	"fmt"
	"strings"

	"github.com/bufbuild/protovalidate-go"
	"github.com/bufbuild/protoyaml-go"
	testv1 "github.com/bufbuild/protoyaml-go/internal/gen/proto/buf/protoyaml/test/v1"
	"google.golang.org/protobuf/proto"
)

// GenGoldenContent generates golden content for the given file path and data.
func GenGoldenContent(filePath string, data []byte) (string, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return "", err
	}

	options := protoyaml.UnmarshalOptions{
		Validator: validator,
		Path:      filePath,
	}
	var val proto.Message
	switch {
	case strings.HasSuffix(filePath, ".proto2test.yaml"):
		testCase := &testv1.Proto2Test{}
		err = options.Unmarshal(data, testCase)
		val = testCase
	case strings.HasSuffix(filePath, ".proto3test.yaml"):
		testCase := &testv1.Proto3Test{}
		err = options.Unmarshal(data, testCase)
		val = testCase
	case strings.HasSuffix(filePath, ".const.yaml"):
		testCase := &testv1.ConstValues{}
		err = options.Unmarshal(data, testCase)
		val = testCase
	case strings.HasSuffix(filePath, ".validate.yaml"):
		testCase := &testv1.ValidateTest{}
		err = options.Unmarshal(data, testCase)
		val = testCase
	default:
		return "", fmt.Errorf("unknown file type: %s", filePath)
	}
	if err != nil {
		return err.Error(), nil //nolint
	}
	opts := protoyaml.MarshalOptions{
		UseProtoNames: true,
	}
	yamlData, err := opts.Marshal(val)
	if err != nil {
		return "", err
	}
	return string(yamlData), nil
}
