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

package golden

import (
	"fmt"
	"strings"

	"buf.build/go/protoyaml"
	testv1 "buf.build/go/protoyaml/internal/gen/proto/buf/protoyaml/test/v1"
	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/protobuf/proto"
)

// protovalidateValidator is a temporary adapter that works with protovalidate-go v0.9.2 and later versions.
// Newer versions (v0.9.3+) take functional options to the validate method so are incompatible w/ protoyaml's Validator.
type protovalidateValidator struct {
	validator protovalidate.Validator
}

func (p *protovalidateValidator) Validate(message proto.Message) error {
	return p.validator.Validate(message)
}

// GenGoldenContent generates golden content for the given file path and data.
//
// If the data is invalid, the error message is returned as the golden content.
// Otherwise, the golden content is the marshaled YAML data.
func GenGoldenContent(filePath string, data []byte) (string, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return "", err
	}

	options := protoyaml.UnmarshalOptions{
		Validator: &protovalidateValidator{validator: validator},
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
	case strings.HasSuffix(filePath, ".test.yaml"):
		testCase := &testv1.Test{}
		options.VersionKind = protoyaml.OptionalVersion
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
