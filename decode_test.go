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
	"github.com/sergi/go-diff/diffmatchpatch"
	"google.golang.org/protobuf/proto"
)

func TestTestData(t *testing.T) {
	t.Parallel()
	// Walk the test data directory for .yaml files
	if err := filepath.Walk("internal/testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".yaml") {
			testRunYAMLFile(t, path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
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
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
	} else {
		defer expectedFile.Close()
		expectedData, err = io.ReadAll(expectedFile)
		if err != nil {
			t.Fatal(err)
		}
	}
	expectedText := string(expectedData)
	if expectedText != errorText {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(expectedText, errorText, false)
		t.Errorf("%s: Test %s failed:\n%s", expectedFileName, testFile, dmp.DiffPrettyText(diffs))
	}
}
