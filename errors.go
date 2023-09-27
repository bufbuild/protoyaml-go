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
	"fmt"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"gopkg.in/yaml.v3"
)

// unmarshalError is returned when unmarshaling fails.
type unmarshalError struct {
	Errors []nodeError
	path   string
	data   []byte
}

func (u *unmarshalError) Error() string {
	var result strings.Builder
	lines := strings.Split(string(u.data), "\n")
	for _, err := range u.Errors {
		result.WriteString(err.DetailedError(u.path, lines))
	}
	return result.String()
}

// nodeError is an error that occurred while processing a specific yaml.Node.
type nodeError struct {
	Node  *yaml.Node
	Cause error
}

// DetailedError returns an error message that includes the path and a code snippet, if
// the lines of the source code are provided.
func (n *nodeError) DetailedError(path string, lines []string) string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("%s:%d:%d %s\n", path, n.Node.Line, n.Node.Column, n.Cause.Error()))
	if n.Node.Line > 0 && n.Node.Line <= len(lines) {
		lineNum := fmt.Sprintf("%4d", n.Node.Line)
		result.WriteString(fmt.Sprintf("%s | %s\n", lineNum, lines[n.Node.Line-1]))
		tailLen := len(lines[n.Node.Line-1])
		if tailLen < 40 {
			tailLen = 40
		}
		tailLen -= n.Node.Column
		if tailLen < 1 {
			tailLen = 1
		}
		marker := strings.Repeat(" ", n.Node.Column-1) + "^" + strings.Repeat(".", tailLen)
		result.WriteString(fmt.Sprintf("%s | %s %s\n", strings.Repeat(" ", len(lineNum)), marker, n.Cause.Error()))
	}
	return result.String()
}

// violationError is singe validation violation.
type violationError struct {
	Violation *validate.Violation
}

// Error prints the field path, message, and constraint ID.
func (v *violationError) Error() string {
	return v.Violation.GetFieldPath() + ": " + v.Violation.GetMessage() + " (" + v.Violation.GetConstraintId() + ")"
}
