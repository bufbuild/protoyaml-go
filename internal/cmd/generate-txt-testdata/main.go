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

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protovalidate-go"
	"github.com/bufbuild/protoyaml-go"
	testv1 "github.com/bufbuild/protoyaml-go/internal/gen/proto/buf/protoyaml/test/v1"
	"google.golang.org/protobuf/proto"
)

func main() {
	if err := run(); err != nil {
		if errString := err.Error(); errString != "" {
			_, _ = fmt.Fprintln(os.Stderr, errString)
		}
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: %s [file or directory]", os.Args[0])
	}
	filePath := os.Args[1]

	// If the file is a directory, recurse
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !strings.HasSuffix(path, ".yaml") {
				return nil
			}
			actualText, err := tryParse(path)
			if err != nil {
				return err
			}
			// Replace the {type}.yaml extension with .expected.txt
			expectedPath := strings.TrimSuffix(path, ".yaml") + ".txt"
			// Write the actual text to the expected file
			if err := os.WriteFile(expectedPath, []byte(actualText), 0600); err != nil {
				return err
			}
			return nil
		})
	}
	actualText, err := tryParse(filePath)
	if err != nil {
		return err
	}
	fmt.Print(actualText)
	return nil
}

func tryParse(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

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
		return err.Error(), nil
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
