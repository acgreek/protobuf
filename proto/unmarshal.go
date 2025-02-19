package proto

import (
	"fmt"
	"math"
	"os"
	"strings"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type WireType int

const (
	VARINT WireType = iota
	I64
	LEN
	SGROUP
	EGROUP
	I32
)

func (wt WireType) String() string {
	switch wt {
	case VARINT:
		return "VARINT"
	case I64:
		return "I64"
	case LEN:
		return "LEN"
	case I32:
		return "I32"
	}
	return "error"
}
func readType(input []byte) (int, WireType, int, error) {
	offset := 0
	if len(input) == 0 {
		return 0, 0, 0, nil
	}
	num, wireTypeNum, offset := protowire.ConsumeTag(input)
	if int(wireTypeNum) > 5 {
		return 0, 0, 0, fmt.Errorf("%d wire type is invalid: %w", wireTypeNum, os.ErrInvalid)
	}
	if int(wireTypeNum) == 3 {
		return 0, 0, 0, fmt.Errorf("sgroup wire type is depreciated: %w", os.ErrInvalid)
	}
	if int(wireTypeNum) == 4 {
		return 0, 0, 0, fmt.Errorf("egroup wire type is depreciated: %w", os.ErrInvalid)
	}
	return offset, WireType(wireTypeNum), int(num), nil
}
func Unmarshal(input []byte, obj proto.Message) error {
	reflMess := obj.ProtoReflect()
	description := reflMess.Descriptor()
	obj.ProtoReflect()
	refs := &unmarshalObj{
		reflMess:    reflMess,
		description: description,
		fields:      description.Fields(),
	}
	return unmarshalWithFields(refs, input)
}

type unmarshalObj struct {
	reflMess    protoreflect.Message
	description protoreflect.MessageDescriptor
	fields      protoreflect.FieldDescriptors
	path        []string
	unknown     []byte
}

func (obj *unmarshalObj) AddKnown(input []byte) {
	obj.unknown = append(obj.unknown, input...)
}

func (obj *unmarshalObj) PopPath() {
	obj.path = obj.path[:len(obj.path)-1]
}

func (obj *unmarshalObj) PushPath(name protoreflect.Name) {
	obj.path = append(obj.path, string(name))
}

func (obj *unmarshalObj) GetPath(field protoreflect.FieldDescriptor) string {
	return safeGetFieldName(obj.path, field)
}

func unmarshalWithFields(obj *unmarshalObj, input []byte) error {
	startOffset := 0
	defer func() {
		if len(obj.unknown) > 0 {
			obj.reflMess.SetUnknown(protoreflect.RawFields(obj.unknown))
		}
	}()
	for {
		if startOffset > len(input) {
			return fmt.Errorf("%s: next field beyond length of buffer", obj.GetPath(nil))
		}
		offset, wireType, index, err := readType(input[startOffset:])
		if err != nil {
			return fmt.Errorf("%s: error at offset %d: %w", obj.GetPath(nil), startOffset, err)
		}
		if offset < 0 {
			return fmt.Errorf("%s: problem reading tag value at offset %d", obj.GetPath(nil), startOffset)
		}
		if offset == 0 { // done reading
			return nil
		}
		field := obj.fields.ByNumber(protoreflect.FieldNumber(index))
		offset += startOffset
		// fmt.Printf("%d wireTypeNum %s, index %d expected index %d\n", offset, wireType, index, expectedIndex)
		switch wireType {
		case VARINT:
			intValue, valOffset := protowire.ConsumeVarint(input[offset:])
			if valOffset < 0 {
				return fmt.Errorf("%s: invalid varint structure", obj.GetPath(field))
			}
			offset += valOffset
			if field != nil && obj.reflMess != nil {
				switch field.Kind() {
				case protoreflect.Int32Kind:
					setSingleOrRepeated(field, obj, protoreflect.ValueOfInt32(int32(intValue)))
				case protoreflect.Sint32Kind:
					setSingleOrRepeated(field, obj, protoreflect.ValueOfInt32(int32(protowire.DecodeZigZag(intValue))))
				case protoreflect.Sint64Kind:
					setSingleOrRepeated(field, obj, protoreflect.ValueOfInt64(int64(protowire.DecodeZigZag(intValue))))
				case protoreflect.Int64Kind:
					setSingleOrRepeated(field, obj, protoreflect.ValueOfInt64(int64(intValue)))
				case protoreflect.Uint32Kind:
					setSingleOrRepeated(field, obj, protoreflect.ValueOfUint32(uint32(intValue)))
				case protoreflect.Uint64Kind:
					setSingleOrRepeated(field, obj, protoreflect.ValueOfUint64(uint64(intValue)))
				case protoreflect.BoolKind:
					setSingleOrRepeated(field, obj, protoreflect.ValueOfBool(intValue == 1))
				case protoreflect.EnumKind:
					setSingleOrRepeated(field, obj, protoreflect.ValueOfEnum(protoreflect.EnumNumber(intValue)))
				}
			} else {
				obj.AddKnown(input[startOffset:offset])
				// obj.reflMess.SetUnknown(protoreflect.RawFields(input[startOffset:offset]))
			}
		case I32:
			intValue, valOffset := protowire.ConsumeFixed32(input[offset:])
			if valOffset < 0 {
				return fmt.Errorf("%s: invalid fixed32 structure", obj.GetPath(field))
			}
			offset += valOffset
			if field != nil && obj.reflMess != nil {
				switch field.Kind() {
				case protoreflect.FloatKind:
					setSingleOrRepeated(field, obj, protoreflect.ValueOfFloat32((math.Float32frombits(intValue))))
				case protoreflect.Fixed32Kind:
					setSingleOrRepeated(field, obj, protoreflect.ValueOfUint32(uint32(intValue)))
				case protoreflect.Sfixed32Kind:
					setSingleOrRepeated(field, obj, protoreflect.ValueOfInt32(int32(intValue)))
				}
			} else {
				obj.AddKnown(input[startOffset:offset])
				// obj.reflMess.SetUnknown(protoreflect.RawFields(input[startOffset:offset]))
			}
		case I64:
			intValue, valOffset := protowire.ConsumeFixed64(input[offset:])
			if valOffset < 0 {
				return fmt.Errorf("%s: invalid fixed64 structure", obj.GetPath(field))
			}
			offset += valOffset
			if field != nil && obj.reflMess != nil {
				switch field.Kind() {
				case protoreflect.DoubleKind:
					setSingleOrRepeated(field, obj, protoreflect.ValueOfFloat64((math.Float64frombits(intValue))))
				case protoreflect.Fixed64Kind:
					setSingleOrRepeated(field, obj, protoreflect.ValueOfUint64(uint64(intValue)))
				case protoreflect.Sfixed64Kind:
					setSingleOrRepeated(field, obj, protoreflect.ValueOfInt64(int64(intValue)))
				}
			} else {
				obj.AddKnown(input[startOffset:offset])
				// obj.reflMess.SetUnknown(protoreflect.RawFields(input[startOffset:offset]))
			}
		case LEN:
			offset, err = validateWireTypeLen(obj, field, input, startOffset, offset)
			if err != nil {
				return err
			}
		}

		if offset == startOffset {
			break
		}
		startOffset = offset
	}
	return nil
}

func setSingleOrRepeated(field protoreflect.FieldDescriptor, obj *unmarshalObj, setVal protoreflect.Value) {
	if field.IsList() {
		existing := obj.reflMess.Mutable(field)
		if existing.IsValid() {
			vl := existing.List()
			vl.Append(setVal)
		}
	} else if field.IsMap() {
		// Set the key and value for the map entry
		mapEntryDescriptor := field.Message()
		keyField := mapEntryDescriptor.Fields().ByName("key")
		valueField := mapEntryDescriptor.Fields().ByName("value")
		keyValue := setVal.Message().Get(keyField)
		value := setVal.Message().Get(valueField)
		existing := obj.reflMess.Mutable(field)
		existing.Map().Set(protoreflect.MapKey(keyValue), value)
	} else {
		obj.reflMess.Set(field, setVal)
	}
}

func validateWireTypeLen(obj *unmarshalObj, field protoreflect.FieldDescriptor, input []byte,
	startOffset, offset int) (int, error) {
	intValue, valOffset := protowire.ConsumeVarint(input[offset:])
	if valOffset < 0 {
		return 0, fmt.Errorf("%s: LEN varint invalid encoding", obj.GetPath(field))
	}
	offset += valOffset
	end := offset + int(intValue)
	if len(input) < end || end < offset {
		return 0, fmt.Errorf("%s: LEN has length greater than size of input buffer", obj.GetPath(field))
	}
	block := input[offset:end]
	offset += int(intValue)
	if field == nil && obj.reflMess != nil {
		// obj.reflMess.SetUnknown(protoreflect.RawFields(input[startOffset:end]))
		obj.AddKnown(input[startOffset:offset])
		return offset, nil
	}
	kindValue := field.Kind()
	switch kindValue {
	case protoreflect.BytesKind:
		setSingleOrRepeated(field, obj, protoreflect.ValueOfBytes(block))
	case protoreflect.StringKind:
		setSingleOrRepeated(field, obj, protoreflect.ValueOfString(string(block)))
	case protoreflect.MessageKind:
		var reflMess protoreflect.Message
		switch {
		case field.IsList():
			reflMess = obj.reflMess.Get(field).List().NewElement().Message()
		case field.IsMap():
			// Create a dynamic message for the map entry
			mapEntryDescriptor := field.Message()
			reflMess = dynamicpb.NewMessage(mapEntryDescriptor)
		default:
			reflMess = obj.reflMess.Get(field).Message().New()
		}

		refs := &unmarshalObj{
			reflMess:    reflMess,
			description: field.Message(),
			fields:      field.Message().Fields(),
			path:        append(obj.path, string(field.Name())),
		}
		obj.PushPath(field.Name())
		err := unmarshalWithFields(refs, block)
		obj.PopPath()
		setSingleOrRepeated(field, obj, protoreflect.ValueOf(refs.reflMess))
		return offset, err
	case protoreflect.Int32Kind:
		setPacked(field, obj, block, func(input []byte) (protoreflect.Value, int) {
			intValue, valOffset := protowire.ConsumeVarint(input)
			v := protoreflect.ValueOfInt32(int32(intValue))
			return v, valOffset
		})
	case protoreflect.Sint32Kind:
		setPacked(field, obj, block, func(input []byte) (protoreflect.Value, int) {
			intValue, valOffset := protowire.ConsumeVarint(input)
			v := protoreflect.ValueOfInt32(int32(protowire.DecodeZigZag(intValue)))
			return v, valOffset
		})
	case protoreflect.Sint64Kind:
		setPacked(field, obj, block, func(input []byte) (protoreflect.Value, int) {
			intValue, valOffset := protowire.ConsumeVarint(input)
			v := protoreflect.ValueOfInt64(int64(protowire.DecodeZigZag(intValue)))
			return v, valOffset
		})
	case protoreflect.Int64Kind:
		setPacked(field, obj, block, func(input []byte) (protoreflect.Value, int) {
			intValue, valOffset := protowire.ConsumeVarint(input)
			v := protoreflect.ValueOfInt64(int64(intValue))
			return v, valOffset
		})
	case protoreflect.Uint32Kind:
		setPacked(field, obj, block, func(input []byte) (protoreflect.Value, int) {
			intValue, valOffset := protowire.ConsumeVarint(input)
			v := protoreflect.ValueOfUint32(uint32(intValue))
			return v, valOffset
		})
	case protoreflect.Uint64Kind:
		setPacked(field, obj, block, func(input []byte) (protoreflect.Value, int) {
			intValue, valOffset := protowire.ConsumeVarint(input)
			v := protoreflect.ValueOfUint64(uint64(intValue))
			return v, valOffset
		})
	case protoreflect.BoolKind:
		setPacked(field, obj, block, func(input []byte) (protoreflect.Value, int) {
			intValue, valOffset := protowire.ConsumeVarint(input)
			v := protoreflect.ValueOfBool(intValue == 1)
			return v, valOffset
		})
	case protoreflect.EnumKind:
		setPacked(field, obj, block, func(input []byte) (protoreflect.Value, int) {
			intValue, valOffset := protowire.ConsumeVarint(input)
			v := protoreflect.ValueOfEnum(protoreflect.EnumNumber(intValue))
			return v, valOffset
		})
	case protoreflect.DoubleKind:
		setPacked(field, obj, block, func(input []byte) (protoreflect.Value, int) {
			intValue, valOffset := protowire.ConsumeFixed64(input)
			v := protoreflect.ValueOfFloat64((math.Float64frombits(intValue)))
			return v, valOffset
		})
	case protoreflect.Fixed64Kind:
		setPacked(field, obj, block, func(input []byte) (protoreflect.Value, int) {
			intValue, valOffset := protowire.ConsumeFixed64(input)
			v := protoreflect.ValueOfUint64(uint64(intValue))
			return v, valOffset
		})
	case protoreflect.Sfixed64Kind:
		setPacked(field, obj, block, func(input []byte) (protoreflect.Value, int) {
			intValue, valOffset := protowire.ConsumeFixed64(input)
			v := protoreflect.ValueOfInt64(int64(intValue))
			return v, valOffset
		})
	case protoreflect.FloatKind:
		setPacked(field, obj, block, func(input []byte) (protoreflect.Value, int) {
			intValue, valOffset := protowire.ConsumeFixed32(input)
			v := protoreflect.ValueOfFloat32((math.Float32frombits(intValue)))
			return v, valOffset
		})
	case protoreflect.Fixed32Kind:
		setPacked(field, obj, block, func(input []byte) (protoreflect.Value, int) {
			intValue, valOffset := protowire.ConsumeFixed32(input)
			v := protoreflect.ValueOfUint32(uint32(intValue))
			return v, valOffset
		})
	case protoreflect.Sfixed32Kind:
		setPacked(field, obj, block, func(input []byte) (protoreflect.Value, int) {
			intValue, valOffset := protowire.ConsumeFixed32(input)
			v := protoreflect.ValueOfInt32(int32(intValue))
			return v, valOffset
		})
	default:
		return offset, fmt.Errorf("%s: unknown kind %s", obj.GetPath(field), kindValue.String())
	}
	return offset, nil
}

func setPacked(field protoreflect.FieldDescriptor, obj *unmarshalObj,
	block []byte, extractor func(input []byte) (protoreflect.Value, int)) error {
	if !field.IsList() {
		return fmt.Errorf("%s: packed field is not a list", obj.GetPath(field))
	}
	existing := obj.reflMess.Mutable(field)
	vl := existing.List()
	offset := 0
	for {
		if len(block[offset:]) == 0 {
			return nil
		}
		v, o := extractor(block[offset:])
		if o < 0 {
			return fmt.Errorf("%s: packed offset invalid", obj.GetPath(field))
		}
		offset += o
		vl.Append(v)
	}
}

func safeGetFieldName(prefix []string, field protoreflect.FieldDescriptor) string {
	if field == nil {
		return "(nil)"
	}
	return strings.Join(prefix, ".") + "." + string(field.Name())
}
