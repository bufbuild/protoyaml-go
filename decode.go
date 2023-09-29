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
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

// Validator is an interface for validating a Protobuf message produced from a given YAML node.
type Validator interface {
	// Validate the given message.
	Validate(message proto.Message) error
}

// UnmarshalOptions is a configurable YAML format parser for Protobuf messages.
type UnmarshalOptions struct {
	// The path for the data being unmarshaled.
	//
	// If set, this will be used when producing error messages.
	Path string
	// Validator is a validator to run after unmarshaling a message.
	Validator Validator
	// Resolver is the Protobuf type resolver to use.
	Resolver interface {
		protoregistry.MessageTypeResolver
		protoregistry.ExtensionTypeResolver
	}
}

// Unmarshal a Protobuf message from the given YAML data.
func Unmarshal(data []byte, message proto.Message) error {
	return (UnmarshalOptions{}).Unmarshal(data, message)
}

// Unmarshal a Protobuf message from the given YAML data.
func (o UnmarshalOptions) Unmarshal(data []byte, message proto.Message) error {
	var yamlFile yaml.Node
	if err := yaml.Unmarshal(data, &yamlFile); err != nil {
		return err
	}
	return o.unmarshalNode(&yamlFile, message, data)
}

func (o UnmarshalOptions) unmarshalNode(node *yaml.Node, message proto.Message, data []byte) error {
	if node.Kind == 0 {
		return nil
	}
	unm := &unmarshaler{
		options:   o,
		custom:    make(map[protoreflect.FullName]customUnmarshaler),
		validator: o.Validator,
		lines:     strings.Split(string(data), "\n"),
	}

	addWktUnmarshalers(unm.custom)

	// Unwrap the document node
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) != 1 {
			return errors.New("expected exactly one node in document")
		}
		node = node.Content[0]
	}

	unm.unmarshalMessage(node, message)
	if unm.validator != nil {
		err := unm.validator.Validate(message)
		var verr *protovalidate.ValidationError
		switch {
		case err == nil: // Valid.
		case errors.As(err, &verr):
			for _, violation := range verr.Violations {
				closest := nodeClosestToPath(node, message.ProtoReflect().Descriptor(), violation.GetFieldPath(), violation.GetForKey())
				unm.addError(closest, &violationError{
					Violation: violation,
				})
			}
		default:
			unm.addError(node, err)
		}
	}

	if len(unm.errors) > 0 {
		return unmarshalErrors(unm.errors)
	}
	return nil
}

type unmarshaler struct {
	options   UnmarshalOptions
	errors    []error
	custom    map[protoreflect.FullName]customUnmarshaler
	validator Validator
	lines     []string
}

func (u *unmarshaler) addError(node *yaml.Node, err error) {
	u.errors = append(u.errors, &nodeError{
		Path:  u.options.Path,
		Node:  node,
		cause: err,
		line:  u.lines[node.Line-1],
	})
}
func (u *unmarshaler) addErrorf(node *yaml.Node, format string, args ...interface{}) {
	u.addError(node, fmt.Errorf(format, args...))
}

func (u *unmarshaler) checkKind(node *yaml.Node, expected yaml.Kind) bool {
	if node.Kind != expected {
		u.addErrorf(node, "expected %v, got %v", getNodeKind(expected), getNodeKind(node.Kind))
		return false
	}
	return true
}

func (u *unmarshaler) checkTag(node *yaml.Node, expected string) {
	if node.Tag != "" && node.Tag != expected {
		u.addErrorf(node, "expected tag %v, got %v", expected, node.Tag)
	}
}

func (u *unmarshaler) findAnyType(node *yaml.Node) protoreflect.MessageType {
	if len(node.Content) == 0 {
		return nil
	}
	typeURL := ""
	for i := 1; i < len(node.Content); i += 2 {
		keyNode := node.Content[i-1]
		valueNode := node.Content[i]
		if keyNode.Value == "@type" && u.checkKind(valueNode, yaml.ScalarNode) {
			typeURL = valueNode.Value
			break
		}
	}
	if typeURL == "" {
		return nil
	}

	// Get the message type.
	var msgType protoreflect.MessageType
	var err error
	if u.options.Resolver != nil {
		msgType, err = u.options.Resolver.FindMessageByURL(typeURL)
	} else { // Use the global registry.
		msgType, err = protoregistry.GlobalTypes.FindMessageByURL(typeURL)
	}
	if err != nil {
		u.addErrorf(node, "unknown type %q: %v", typeURL, err)
		return nil
	}
	return msgType
}

func (u *unmarshaler) findType(msgDesc protoreflect.MessageDescriptor) protoreflect.MessageType {
	var msgType protoreflect.MessageType
	var err error
	if u.options.Resolver != nil {
		msgType, err = u.options.Resolver.FindMessageByName(msgDesc.FullName())
	} else { // Use the global registry
		msgType, err = protoregistry.GlobalTypes.FindMessageByName(msgDesc.FullName())
	}
	if err != nil {
		return nil
	}
	return msgType
}

// Unmarshal the field based on the field kind, ignoring IsList and IsMap,
// which are handled by the caller.
func (u *unmarshaler) unmarshalScalar(
	node *yaml.Node,
	field protoreflect.FieldDescriptor,
	forKey bool,
) protoreflect.Value {
	switch field.Kind() {
	case protoreflect.BoolKind:
		return protoreflect.ValueOfBool(u.unmarshalBool(node, forKey))
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return protoreflect.ValueOfInt32(int32(u.unmarshalInteger(node, 32)))
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return protoreflect.ValueOfInt64(u.unmarshalInteger(node, 64))
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return protoreflect.ValueOfUint32(uint32(u.unmarshalUnsigned(node, 32)))
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return protoreflect.ValueOfUint64(u.unmarshalUnsigned(node, 64))
	case protoreflect.FloatKind:
		return protoreflect.ValueOfFloat32(float32(u.unmarshalFloat(node, 32)))
	case protoreflect.DoubleKind:
		return protoreflect.ValueOfFloat64(u.unmarshalFloat(node, 64))
	case protoreflect.StringKind:
		u.checkKind(node, yaml.ScalarNode)
		return protoreflect.ValueOfString(node.Value)
	case protoreflect.BytesKind:
		return protoreflect.ValueOfBytes(u.unmarshalBytes(node))
	case protoreflect.EnumKind:
		return protoreflect.ValueOfEnum(u.unmarshalEnum(node, field))
	default:
		u.addErrorf(node, "unimplemented scalar type %v", field.Kind())
		return protoreflect.Value{}
	}
}

// Base64 decodes the given node value.
func (u *unmarshaler) unmarshalBytes(node *yaml.Node) []byte {
	if !u.checkKind(node, yaml.ScalarNode) {
		return nil
	}

	// base64 decode the value.
	data, err := base64.StdEncoding.DecodeString(node.Value)
	if err != nil {
		u.addErrorf(node, "invalid base64: %v", err)
	}
	return data
}

// Unmarshal raw `true` or `false` values, only allowing for strings for keys.
func (u *unmarshaler) unmarshalBool(node *yaml.Node, forKey bool) bool {
	if u.checkKind(node, yaml.ScalarNode) {
		switch node.Value {
		case "true":
			if !forKey {
				u.checkTag(node, "!!bool")
			}
			return true
		case "false":
			if !forKey {
				u.checkTag(node, "!!bool")
			}
			return false
		default:
			u.addErrorf(node, "expected bool, got %#v", node.Value)
		}
	}
	return false
}

// Unmarshal the given node into an enum value.
//
// Accepts either the enum name or number.
func (u *unmarshaler) unmarshalEnum(node *yaml.Node, field protoreflect.FieldDescriptor) protoreflect.EnumNumber {
	u.checkKind(node, yaml.ScalarNode)
	// Get the enum descriptor.
	enumDesc := field.Enum()
	if enumDesc.FullName() == "google.protobuf.NullValue" {
		return 0
	}
	// Get the enum value.
	enumVal := enumDesc.Values().ByName(protoreflect.Name(node.Value))
	if enumVal == nil {
		neg, parsed, err := parseSignedLit(node.Value)
		if err != nil || parsed > 1<<32 {
			u.addErrorf(node, "unknown enum value %#v, expected one of %v", node.Value,
				getEnumValueNames(enumDesc.Values()))
		}
		num := protoreflect.EnumNumber(parsed)
		if neg {
			num = -num
		}
		return num
	}
	return enumVal.Number()
}

// Unmarshal the given node into a float with the given bits.
func (u *unmarshaler) unmarshalFloat(node *yaml.Node, bits int) float64 {
	if !u.checkKind(node, yaml.ScalarNode) {
		return 0
	}

	parsed, err := strconv.ParseFloat(node.Value, bits)
	if err != nil {
		u.addErrorf(node, "invalid float: %v", err)
	}
	return parsed
}

// Unmarshal the given node into an unsigned integer with the given bits.
func (u *unmarshaler) unmarshalUnsigned(node *yaml.Node, bits int) uint64 {
	if !u.checkKind(node, yaml.ScalarNode) {
		return 0
	}

	parsed, err := parseUnsignedLit(node.Value)
	if err != nil {
		u.addErrorf(node, "invalid integer: %v", err)
	}
	if bits < 64 && parsed >= 1<<bits {
		u.addErrorf(node, "integer is too large: > %v", 1<<bits-1)
	}
	return parsed
}

// Unmarshal the given node into a signed integer with the given bits.
func (u *unmarshaler) unmarshalInteger(node *yaml.Node, bits int) int64 {
	if !u.checkKind(node, yaml.ScalarNode) {
		return 0
	}

	neg, parsed, err := parseSignedLit(node.Value)
	if err != nil {
		u.addErrorf(node, "invalid integer: %v", err)
	}
	if neg {
		if parsed <= 1<<(bits-1) {
			return -int64(parsed)
		}
		u.addErrorf(node, "integer is too small: < %v", -(1 << (bits - 1)))
	} else if parsed >= 1<<(bits-1) {
		u.addErrorf(node, "integer is too large: > %v", 1<<(bits-1)-1)
	}
	return int64(parsed)
}

func getFieldNames(fields protoreflect.FieldDescriptors) []protoreflect.Name {
	names := make([]protoreflect.Name, 0, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		names = append(names, fields.Get(i).Name())
		if i > 5 {
			names = append(names, protoreflect.Name("..."))
			break
		}
	}
	return names
}

func getEnumValueNames(values protoreflect.EnumValueDescriptors) []protoreflect.Name {
	names := make([]protoreflect.Name, 0, values.Len())
	for i := 0; i < values.Len(); i++ {
		names = append(names, values.Get(i).Name())
		if i > 5 {
			names = append(names, protoreflect.Name("..."))
			break
		}
	}
	return names
}

func getNodeKind(kind yaml.Kind) string {
	switch kind {
	case yaml.DocumentNode:
		return "document"
	case yaml.SequenceNode:
		return "sequence"
	case yaml.MappingNode:
		return "mapping"
	case yaml.ScalarNode:
		return "scalar"
	case yaml.AliasNode:
		return "alias"
	}
	return fmt.Sprintf("unknown(%d)", kind)
}

func getExpectedNodeKind(field protoreflect.FieldDescriptor, forList bool) string {
	if field.IsList() && !forList {
		return "sequence"
	}
	switch field.Kind() {
	case protoreflect.EnumKind:
		return "enum"
	case protoreflect.BoolKind:
		return "boolean"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float"
	case protoreflect.DoubleKind:
		return "double"
	case protoreflect.MessageKind:
		return "mapping"
	default:
		return "scalar"
	}
}

// Parses Octal, Hex, Binary, Decimal, and Unsigned Integer Float literals.
//
// Conversion through JSON/YAML may have converted integers into floats, including
// exponential notation. This function will parse those values back into integers
// if possible.
func parseUnsignedLit(value string) (uint64, error) {
	base := 10
	if len(value) >= 2 && strings.HasPrefix(value, "0") {
		switch value[1] {
		case 'x', 'X':
			base = 16
			value = value[2:]
		case 'o', 'O':
			base = 8
			value = value[2:]
		case 'b', 'B':
			base = 2
			value = value[2:]
		}
	}

	parsed, err := strconv.ParseUint(value, base, 64)
	if err != nil {
		parsedFloat, floatErr := strconv.ParseFloat(value, 64)
		if floatErr != nil || parsedFloat < 0 {
			return 0, err
		}
		// See if it's actually an integer.
		parsed = uint64(parsedFloat)
		if float64(parsed) != parsedFloat || parsed >= (1<<53) {
			return parsed, errors.New("precision loss")
		}
	}
	return parsed, nil
}

func parseSignedLit(value string) (bool, uint64, error) {
	var negative bool
	if strings.HasPrefix(value, "-") {
		negative = true
		value = value[1:]
	}
	parsed, err := parseUnsignedLit(value)
	return negative, parsed, err
}

// Searches for the field with the given 'key' first by Name, then by JSONName,
// and finally by Number.
func findField(key string, fields protoreflect.FieldDescriptors) protoreflect.FieldDescriptor {
	if field := fields.ByName(protoreflect.Name(key)); field != nil {
		return field
	}
	if field := fields.ByJSONName(key); field != nil {
		return field
	}
	num, err := strconv.ParseInt(key, 10, 32)
	if err == nil {
		if field := fields.ByNumber(protoreflect.FieldNumber(num)); field != nil {
			return field
		}
	}
	return nil
}

// Unmarshal a field, handling isList/isMap.
func (u *unmarshaler) unmarshalField(node *yaml.Node, field protoreflect.FieldDescriptor, message proto.Message) {
	switch {
	case field.IsList():
		u.unmarshalList(node, field, message.ProtoReflect().Mutable(field).List())
	case field.IsMap():
		u.unmarshalMap(node, field, message.ProtoReflect().Mutable(field).Map())
	case field.Kind() == protoreflect.MessageKind:
		u.unmarshalMessage(node, message.ProtoReflect().Mutable(field).Message().Interface())
	default:
		message.ProtoReflect().Set(field, u.unmarshalScalar(node, field, false))
	}
}

// Unmarshal the list, with explicit handling for lists of messages.
func (u *unmarshaler) unmarshalList(node *yaml.Node, field protoreflect.FieldDescriptor, list protoreflect.List) {
	if u.checkKind(node, yaml.SequenceNode) {
		switch field.Kind() {
		case protoreflect.MessageKind, protoreflect.GroupKind:
			for _, itemNode := range node.Content {
				msgVal := list.NewElement()
				u.unmarshalMessage(itemNode, msgVal.Message().Interface())
				list.Append(msgVal)
			}
		default:
			for _, itemNode := range node.Content {
				list.Append(u.unmarshalScalar(itemNode, field, false))
			}
		}
	}
}

// Unmarshal the map, with explicit handling for maps to messages.
func (u *unmarshaler) unmarshalMap(node *yaml.Node, field protoreflect.FieldDescriptor, mapVal protoreflect.Map) {
	if u.checkKind(node, yaml.MappingNode) {
		mapKeyField := field.MapKey()
		mapValueField := field.MapValue()
		for i := 1; i < len(node.Content); i += 2 {
			keyNode := node.Content[i-1]
			valueNode := node.Content[i]
			mapKey := u.unmarshalScalar(keyNode, mapKeyField, true)
			switch mapValueField.Kind() {
			case protoreflect.MessageKind, protoreflect.GroupKind:
				mapValue := mapVal.NewValue()
				u.unmarshalMessage(valueNode, mapValue.Message().Interface())
				mapVal.Set(mapKey.MapKey(), mapValue)
			default:
				mapVal.Set(mapKey.MapKey(), u.unmarshalScalar(valueNode, mapValueField, false))
			}
		}
	}
}

// Unmarshal the given yaml node into the given proto.Message.
func (u *unmarshaler) unmarshalMessage(node *yaml.Node, message proto.Message) {
	if node.Tag == "!!null" {
		return // Null is always allowed for messages
	}

	// Check for a custom unmarshaler
	custom, ok := u.custom[message.ProtoReflect().Descriptor().FullName()]
	if ok && custom(u, node, message) {
		return // Custom unmarshaler handled the decoding
	}
	if node.Kind != yaml.MappingNode {
		u.addErrorf(node, "expected fields for %v, got %v",
			message.ProtoReflect().Descriptor().FullName(), getNodeKind(node.Kind))
		return
	}
	// Decode the fields
	fields := message.ProtoReflect().Descriptor().Fields()
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		if u.checkKind(keyNode, yaml.ScalarNode) && keyNode.Value != "@type" {
			// Get the field Name, JSONName, or Number
			if field := findField(keyNode.Value, fields); field != nil {
				valueNode := node.Content[i+1]
				u.unmarshalField(valueNode, field, message)
			} else {
				u.addErrorf(keyNode, "unknown field %#v, expended one of %v", keyNode.Value, getFieldNames(fields))
			}
		}
	}
}

type customUnmarshaler func(u *unmarshaler, node *yaml.Node, message proto.Message) bool

// Add all well-known type unmarshalers to the given map (including struct unmarshalers).
func addWktUnmarshalers(custom map[protoreflect.FullName]customUnmarshaler) {
	custom["google.protobuf.Any"] = unmarshalAnyMsg

	custom["google.protobuf.Duration"] = unmarshalDurationMsg
	custom["google.protobuf.Timestamp"] = unmarshalTimestampMsg

	custom["google.protobuf.BoolValue"] = unmarshalWrapperMsg
	custom["google.protobuf.BytesValue"] = unmarshalWrapperMsg
	custom["google.protobuf.DoubleValue"] = unmarshalWrapperMsg
	custom["google.protobuf.FloatValue"] = unmarshalWrapperMsg
	custom["google.protobuf.Int32Value"] = unmarshalWrapperMsg
	custom["google.protobuf.Int64Value"] = unmarshalWrapperMsg
	custom["google.protobuf.UInt32Value"] = unmarshalWrapperMsg
	custom["google.protobuf.UInt64Value"] = unmarshalWrapperMsg
	custom["google.protobuf.StringValue"] = unmarshalWrapperMsg

	custom["google.protobuf.Value"] = unmarshalValueMsg
	custom["google.protobuf.ListValue"] = unmarshalListValueMsg
	custom["google.protobuf.Struct"] = unmarshalStructMsg
}

func unmarshalAnyMsg(unm *unmarshaler, node *yaml.Node, message proto.Message) bool {
	anyVal, ok := message.(*anypb.Any)
	if !ok || !unm.checkKind(node, yaml.MappingNode) || len(node.Content) == 0 {
		return false
	}

	// Get the message type
	msgType := unm.findAnyType(node)
	if msgType != nil {
		protoVal := msgType.New()
		unm.unmarshalMessage(node, protoVal.Interface())
		err := anyVal.MarshalFrom(protoVal.Interface())
		if err != nil {
			unm.addErrorf(node, "failed to marshal %v: %v", msgType.Descriptor().FullName(), err)
		}
	}

	return true
}

const (
	maxTimestampSeconds = 253402300799
	minTimestampSeconds = -62135596800
)

// Format is decimal seconds with up to 9 fractional digits, followed by an 's'.
func parseDuration(txt string, duration *durationpb.Duration) error {
	// Remove trailing s.
	if txt[len(txt)-1] != 's' {
		return errors.New("missing trailing 's'")
	}
	value := txt[:len(txt)-1]

	// Split into seconds and nanos.
	parts := strings.Split(value, ".")
	switch len(parts) {
	case 1:
		// seconds only
		seconds, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return err
		}
		duration.Seconds = seconds
		duration.Nanos = 0
	case 2:
		// seconds and up to 9 digits of fractional seconds
		seconds, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return err
		}
		duration.Seconds = seconds
		nanos, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return err
		}
		power := 9 - len(parts[1])
		if power < 0 {
			return errors.New("too many fractional second digits")
		}
		nanos *= 10 ^ int64(power)
		duration.Nanos = int32(nanos)
	default:
		return errors.New("invalid duration: too many '.' characters")
	}
	return nil
}

// Format is RFC3339Nano, limited to the range 0001-01-01T00:00:00Z to
// 9999-12-31T23:59:59Z inclusive.
func parseTimestamp(txt string, timestamp *timestamppb.Timestamp) error {
	parsed, err := time.Parse(time.RFC3339Nano, txt)
	if err != nil {
		return err
	}
	// Validate seconds.
	secs := parsed.Unix()
	if secs < minTimestampSeconds {
		return errors.New("before 0001-01-01T00:00:00Z")
	} else if secs > maxTimestampSeconds {
		return errors.New("after 9999-12-31T23:59:59Z")
	}
	// Validate nanos.
	subsecond := strings.LastIndexByte(txt, '.')
	timezone := strings.LastIndexAny(txt, "Z-+")
	if subsecond >= 0 && timezone >= subsecond && timezone-subsecond > len(".999999999") {
		return errors.New("too many fractional second digits")
	}

	timestamp.Seconds = secs
	timestamp.Nanos = int32(parsed.Nanosecond())
	return nil
}

func unmarshalDurationMsg(unm *unmarshaler, node *yaml.Node, message proto.Message) bool {
	duration, ok := message.(*durationpb.Duration)
	if node.Kind != yaml.ScalarNode || len(node.Value) == 0 || !ok {
		return false
	}

	err := parseDuration(node.Value, duration)
	if err != nil {
		unm.addErrorf(node, "invalid duration: %v", err)
	}
	return true
}

func unmarshalTimestampMsg(unm *unmarshaler, node *yaml.Node, message proto.Message) bool {
	timestamp, ok := message.(*timestamppb.Timestamp)
	if node.Kind != yaml.ScalarNode || len(node.Value) == 0 || !ok {
		return false
	}
	err := parseTimestamp(node.Value, timestamp)
	if err != nil {
		unm.addErrorf(node, "invalid timestamp: %v", err)
	}
	return true
}

// Forwards unmarshaling to the "value" field of the given wrapper message.
func unmarshalWrapperMsg(unm *unmarshaler, node *yaml.Node, message proto.Message) bool {
	valueField := message.ProtoReflect().Descriptor().Fields().ByName("value")
	if node.Kind == yaml.MappingNode || valueField == nil {
		return false
	}
	unm.unmarshalField(node, valueField, message)
	return true
}

func unmarshalValueMsg(unm *unmarshaler, node *yaml.Node, message proto.Message) bool {
	if value, ok := message.(*structpb.Value); ok {
		unmarshalValue(unm, node, value, nil, false)
		return true
	}
	return false
}

func unmarshalListValueMsg(unm *unmarshaler, node *yaml.Node, message proto.Message) bool {
	listValue, ok := message.(*structpb.ListValue)
	if !ok || node.Kind != yaml.SequenceNode {
		return false
	}

	unmarshalListValue(unm, node, listValue, nil)
	return true
}

func unmarshalStructMsg(unm *unmarshaler, node *yaml.Node, message proto.Message) bool {
	structVal, ok := message.(*structpb.Struct)
	if !ok || node.Kind != yaml.MappingNode {
		return false
	}
	unmarshalStruct(unm, node, structVal, nil, nil)
	return true
}

// Unmarshal the given yaml node into a structpb.Value, using the given
// field descriptor to validate the type, if non-nil.
func unmarshalValue(
	unm *unmarshaler,
	node *yaml.Node,
	value *structpb.Value,
	field protoreflect.FieldDescriptor, forList bool,
) {
	// Check for a custom unmarshaler
	if field != nil && field.Kind() == protoreflect.MessageKind && !field.IsMap() {
		if custom, ok := unm.custom[field.Message().FullName()]; ok {
			msgType := unm.findType(field.Message())
			if msgType != nil && custom(unm, node, msgType.New().Interface()) {
				field = nil // Custom unmarshaler handled the errors.
			}
		}
	}

	// Unmarshal the value.
	switch node.Kind {
	case yaml.SequenceNode: // A list.
		listValue := &structpb.ListValue{}
		unmarshalListValue(unm, node, listValue, field)
		value.Kind = &structpb.Value_ListValue{ListValue: listValue}
	case yaml.MappingNode: // A message or map.
		structVal := &structpb.Struct{}
		unmarshalStruct(unm, node, structVal, nil, field)
		value.Kind = &structpb.Value_StructValue{StructValue: structVal}
	case yaml.ScalarNode:
		if field != nil && field.IsList() && !forList {
			unm.addErrorf(node, "expected %s for %v, got scalar", getExpectedNodeKind(field, forList), field.FullName())
		}
		unmarshalScalarValue(unm, node, value, field)
	case 0:
		value.Kind = &structpb.Value_NullValue{}
	default:
		unm.addErrorf(node, "unimplemented value kind: %v", getNodeKind(node.Kind))
	}
}

// Unmarshal the given yaml node into a structpb.ListValue, using the given field
// descriptor to validate each item, if non-nil.
func unmarshalListValue(
	unm *unmarshaler,
	node *yaml.Node,
	list *structpb.ListValue,
	field protoreflect.FieldDescriptor,
) {
	if field != nil && !field.IsList() {
		unm.addErrorf(node, "expected %s for %v, got sequence", getExpectedNodeKind(field, false), field.FullName())
	}
	for _, itemNode := range node.Content {
		itemValue := &structpb.Value{}
		unmarshalValue(unm, itemNode, itemValue, field, true)
		list.Values = append(list.Values, itemValue)
	}
}

// Unmarshal the given yaml node into a structpb.Struct
//
// Structs can represent either a message or a map.
// For messages, the message descriptor can be provided or inferred from the node.
// For maps, the field descriptor can be provided to validate the map keys/values.
func unmarshalStruct(
	unm *unmarshaler,
	node *yaml.Node,
	message *structpb.Struct,
	msgDesc protoreflect.MessageDescriptor,
	field protoreflect.FieldDescriptor,
) {
	if field != nil {
		// Validate the field is a message or map.
		if !field.IsMap() && field.Kind() != protoreflect.MessageKind {
			unm.addErrorf(node, "expected %s for %v, got mapping", getExpectedNodeKind(field, false), field.FullName())
		}
	} else if msgDesc == nil {
		// Try to find the message descriptor.
		msgType := unm.findAnyType(node)
		if msgType != nil {
			msgDesc = msgType.Descriptor()
		}
	}

	for i := 1; i < len(node.Content); i += 2 {
		keyNode := node.Content[i-1]
		// Validate the key.
		if !unm.checkKind(keyNode, yaml.ScalarNode) {
			continue
		} else if field != nil && field.MapKey() != nil {
			// Try to unmarshal the key, to validate it.
			tmp := &structpb.Value{}
			unmarshalValue(unm, keyNode, tmp, field.MapKey(), false)
		}

		// Try to resolve the value descriptor.
		var valueDesc protoreflect.FieldDescriptor
		if msgDesc != nil {
			valueDesc = findField(keyNode.Value, msgDesc.Fields())
		} else if field != nil && field.MapValue() != nil {
			valueDesc = field.MapValue()
		}

		// Unmarshal the value.
		valueNode := node.Content[i]
		value := &structpb.Value{}
		unmarshalValue(unm, valueNode, value, valueDesc, false)
		if message.Fields == nil {
			message.Fields = make(map[string]*structpb.Value)
		}
		message.Fields[keyNode.Value] = value
	}
}

func unmarshalScalarValue(
	unm *unmarshaler,
	node *yaml.Node,
	value *structpb.Value,
	field protoreflect.FieldDescriptor,
) {
	switch node.Tag {
	case "!!null":
		unmarshalScalarNull(unm, node, value, field)
	case "!!bool":
		unmarshalScalarBool(unm, node, value, field)
	default:
		unmarshalScalarString(unm, node, value, field)
	}
}

// Can be a Message, NULL_VALUE enum, or a string.
func unmarshalScalarNull(
	unm *unmarshaler,
	node *yaml.Node,
	value *structpb.Value,
	field protoreflect.FieldDescriptor,
) {
	if field != nil {
		switch field.Kind() {
		case protoreflect.BytesKind:
			checkBytes(unm, node, field)
		case protoreflect.MessageKind, protoreflect.StringKind:
		default:
			unm.addErrorf(node, "expected %s for %v, got null", getExpectedNodeKind(field, true), field.FullName())
		}
	}
	value.Kind = &structpb.Value_NullValue{}
}

// bool, string, or bytes.
func unmarshalScalarBool(
	unm *unmarshaler,
	node *yaml.Node,
	value *structpb.Value,
	field protoreflect.FieldDescriptor,
) {
	if field != nil {
		switch field.Kind() {
		case protoreflect.BoolKind, protoreflect.StringKind:
		case protoreflect.BytesKind:
			checkBytes(unm, node, field)
		default:
			unm.addErrorf(node, "expected %s for %v, got bool", getExpectedNodeKind(field, true), field.FullName())
		}
	}
	switch node.Value {
	case "true":
		value.Kind = &structpb.Value_BoolValue{BoolValue: true}
	case "false":
		value.Kind = &structpb.Value_BoolValue{BoolValue: false}
	default: // This is a string, not a bool.
		if field != nil && field.Kind() != protoreflect.StringKind {
			unm.addErrorf(node, "expected %s for %v, got string", getExpectedNodeKind(field, true), field.FullName())
		}
		value.Kind = &structpb.Value_StringValue{StringValue: node.Value}
	}
}

// Can be string, bytes, float, or int.
func unmarshalScalarString(unm *unmarshaler, node *yaml.Node, value *structpb.Value, field protoreflect.FieldDescriptor) {
	floatVal, err := strconv.ParseFloat(node.Value, 64)
	if err != nil {
		// String or bytes.
		if field != nil {
			switch field.Kind() {
			case protoreflect.StringKind:
			case protoreflect.BytesKind:
				checkBytes(unm, node, field)
			default:
				unm.addErrorf(node, "expected %s for %v, got string", getExpectedNodeKind(field, true), field.FullName())
			}
		}
		value.Kind = &structpb.Value_StringValue{StringValue: node.Value}
		return
	}

	if math.IsInf(floatVal, 0) || math.IsNaN(floatVal) {
		// String or float.
		if field != nil {
			switch field.Kind() {
			case protoreflect.StringKind, protoreflect.FloatKind, protoreflect.DoubleKind:
			case protoreflect.BytesKind:
				checkBytes(unm, node, field)
			default:
				unm.addErrorf(node, "expected %s for %v, got float", getExpectedNodeKind(field, true), field.FullName())
			}
		}
		value.Kind = &structpb.Value_StringValue{StringValue: node.Value}
		return
	}

	// String, float, or int.
	unmarshalScalarFloat(unm, node, value, field, floatVal)
}

func unmarshalScalarFloat(
	unm *unmarshaler,
	node *yaml.Node,
	value *structpb.Value,
	field protoreflect.FieldDescriptor,
	floatVal float64,
) {
	// Try to parse it as in integer, to see if the float representation is lossy.
	neg, uintVal, uintErr := parseSignedLit(node.Value)

	// Check if we can represent this as a number.
	floatUintVal := uint64(math.Abs(floatVal))     // The uint64 representation of the float.
	if uintErr != nil || floatUintVal == uintVal { // Safe to represent as a number.
		value.Kind = &structpb.Value_NumberValue{NumberValue: floatVal}
	} else { // Keep string representation.
		value.Kind = &structpb.Value_StringValue{StringValue: node.Value}
	}

	// Check for type/precision errors.
	if field != nil {
		switch field.Kind() {
		case protoreflect.StringKind:
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
			switch {
			case uintErr != nil:
				unm.addErrorf(node, "expected int32 for %v, but: %v", field.FullName(), uintErr)
			case neg && uintVal > 1<<31: // Underflow.
				unm.addErrorf(node, "expected int32 for %v, got int64", field.FullName())
			case !neg && uintVal >= 1<<31: // Overflow.
				unm.addErrorf(node, "expected int32 for %v, got int64", field.FullName())
			}
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			switch {
			case uintErr != nil:
				unm.addErrorf(node, "expected int64 for %v, but: %v", field.FullName(), uintErr)
			case neg && uintVal > 1<<63: // Underflow.
				unm.addErrorf(node, "expected int64 for %v, but: out of range", field.FullName())
			case !neg && uintVal >= 1<<63: // Overflow.
				unm.addErrorf(node, "expected int64 for %v, got uint64", field.FullName())
			}
		case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
			switch {
			case uintErr != nil:
				unm.addErrorf(node, "expected uint32 for %v, but: %v", field.FullName(), uintErr)
			case neg: // Underflow.
				unm.addErrorf(node, "expected uint32 for %v, got negative", field.FullName())
			case uintVal >= 1<<32: // Overflow.
				unm.addErrorf(node, "expected uint32 for %v, got uint64", field.FullName())
			}
		case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			switch {
			case uintErr != nil:
				unm.addErrorf(node, "expected uint64 for %v, but: %v", field.FullName(), uintErr)
			case neg: // Underflow.
				unm.addErrorf(node, "expected uint64 for %v, got negative", field.FullName())
			}
		case protoreflect.FloatKind:
			if math.Abs(floatVal) > math.MaxFloat32 {
				unm.addErrorf(node, "expected %s for %v, got float64", getExpectedNodeKind(field, true), field.FullName())
			} else if uintErr == nil && uint64(float32(math.Abs(floatVal))) != uintVal {
				// Loss of precision.
				unm.addErrorf(node, "expected %s for %v, got integer", getExpectedNodeKind(field, true), field.FullName())
			}
		case protoreflect.DoubleKind:
			if uintErr == nil && floatUintVal != uintVal {
				// Loss of precision.
				unm.addErrorf(node, "expected %s for %v, got integer", getExpectedNodeKind(field, true), field.FullName())
			}
		case protoreflect.BytesKind:
			checkBytes(unm, node, field)
		default:
			unm.addErrorf(node, "expected %s for %v, got number", getExpectedNodeKind(field, true), field.FullName())
		}
	}
}

func checkBytes(unm *unmarshaler, node *yaml.Node, field protoreflect.FieldDescriptor) {
	if _, err := base64.StdEncoding.DecodeString(node.Value); err != nil {
		unm.addErrorf(node, "expected base64 for %v, but %v", field.FullName(), err)
	}
}

// NodeClosestToPath returns the node closest to the given field path.
//
// If toKey is true, the key node is returned if the path points to a map entry.
//
// Example field paths:
//   - 'foo' -> the field foo
//   - 'foo[0]' -> the first element of the repeated field foo or the map entry with key '0'
//   - 'foo.bar' -> the field bar in the message field foo
//   - 'foo["bar"]' -> the map entry with key 'bar' in the map field foo
func nodeClosestToPath(root *yaml.Node, msgDesc protoreflect.MessageDescriptor, path string, toKey bool) *yaml.Node {
	parsedPath, err := parseFieldPath(path)
	if err != nil {
		return root
	}
	return findNodeByPath(root, msgDesc, parsedPath, toKey)
}

func parseFieldPath(path string) ([]string, error) {
	if len(path) == 0 {
		return nil, nil
	}
	next, path := parseNextFieldName(path)
	result := []string{next}
	for len(path) > 0 {
		switch path[0] {
		case '[': // Parse array index or map key.
			next, path = parseNextValue(path[1:])
		case '.': // Parse field name.
			next, path = parseNextFieldName(path[1:])
		default:
			return nil, errors.New("invalid path")
		}
		result = append(result, next)
	}
	return result, nil
}

func parseNextFieldName(path string) (string, string) {
	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '.':
			return path[:i], path[i:]
		case '[':
			return path[:i], path[i:]
		}
	}
	return path, ""
}

func parseNextValue(path string) (string, string) {
	if len(path) == 0 {
		return "", ""
	}
	if path[0] == '"' {
		// Parse string.
		for i := 1; i < len(path); i++ {
			if path[i] == '\\' {
				i++ // Skip escaped character.
			} else if path[i] == '"' {
				result, err := strconv.Unquote(path[:i+1])
				if err != nil {
					return "", ""
				}
				return result, path[i+2:]
			}
		}
		return path, ""
	}
	// Go til the trailing ']'
	for i := 0; i < len(path); i++ {
		if path[i] == ']' {
			return path[:i], path[i+1:]
		}
	}
	return path, ""
}

// Returns the node as close to the given path as possible.
func findNodeByPath(root *yaml.Node, msgDesc protoreflect.MessageDescriptor, path []string, toKey bool) *yaml.Node {
	cur := root
	curMsg := msgDesc
	var curMap protoreflect.FieldDescriptor
	for i, key := range path {
		switch cur.Kind {
		case yaml.MappingNode:
			found := false
			if curMsg != nil {
				field := findField(key, curMsg.Fields())
				if field == nil {
					return cur
				}
				cur, found = findNodeByField(cur, field)
				if found {
					if field.IsMap() {
						curMap = field
						curMsg = nil
					} else {
						curMap = nil
						curMsg = field.Message()
					}
				}
			} else if curMap != nil {
				var keyNode *yaml.Node
				keyNode, cur, found = findEntryByKey(cur, key)
				if found {
					if i == len(path)-1 && toKey {
						return keyNode
					}
					curMsg = curMap.MapValue().Message()
					curMap = nil
				}
			}
			if !found {
				return cur
			}
		case yaml.SequenceNode:
			idx, err := strconv.Atoi(key)
			if err != nil || idx < 0 || idx >= len(cur.Content) {
				return cur
			}
			cur = cur.Content[idx]
		default:
			return cur
		}
	}
	return cur
}

func findNodeByField(cur *yaml.Node, field protoreflect.FieldDescriptor) (*yaml.Node, bool) {
	fieldNum := fmt.Sprintf("%d", field.Number())
	for i := 1; i < len(cur.Content); i += 2 {
		keyNode := cur.Content[i-1]
		if keyNode.Value == string(field.Name()) ||
			keyNode.Value == field.JSONName() ||
			keyNode.Value == fieldNum {
			return cur.Content[i], true
		}
	}
	return cur, false
}

func findEntryByKey(cur *yaml.Node, key string) (*yaml.Node, *yaml.Node, bool) {
	for i := 1; i < len(cur.Content); i += 2 {
		keyNode := cur.Content[i-1]
		if keyNode.Value == key {
			return keyNode, cur.Content[i], true
		}
	}
	return nil, cur, false
}
