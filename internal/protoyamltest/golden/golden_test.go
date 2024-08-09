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
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestGoldenFiles(t *testing.T) {
	t.Parallel()
	// Walk the test data directory for .yaml files
	fsys := os.DirFS("../../..")
	if err := fs.WalkDir(fsys, "internal/testdata", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".yaml") {
			t.Run(path, func(t *testing.T) {
				t.Parallel()
				testRunYAMLFile(t, fsys, path)
			})
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func testRunYAMLFile(t *testing.T, fsys fs.FS, testFile string) {
	t.Helper()

	file, err := fsys.Open(testFile)
	require.NoError(t, err)
	defer file.Close()

	data, err := io.ReadAll(file)
	require.NoError(t, err)

	actualText, err := GenGoldenContent(testFile, data)
	require.NoError(t, err)

	// Read the expected file
	expectedFileName := strings.TrimSuffix(testFile, ".yaml") + ".txt"
	expectedFile, err := fsys.Open(expectedFileName)
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
	if expectedText != actualText {
		diff := cmp.Diff(expectedText, actualText)
		t.Errorf("%s: Test %s failed:\nExpected:\n%s\nActual:\n%s\nDiff:\n%s", expectedFileName, testFile, expectedText, actualText, diff)
	}
}
