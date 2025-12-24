package hpack

import (
	"encoding"
	"fmt"
	"reflect"
)

var valueEncoders []encoderFunc

func init() {
	valueEncoders = []encoderFunc{
		reflect.Bool:  encodeBoolValue,
		reflect.Int:   encodeIntValue,
		reflect.Int8:  encodeIntValue,
		reflect.Int16: encodeIntValue,
		reflect.Int32: encodeIntValue,
		reflect.Int64: encodeIntValue,
		// reflect.Int8:          encodeInt8CondValue,
		// reflect.Int16:         encodeInt16CondValue,
		// reflect.Int32:         encodeInt32CondValue,
		// reflect.Int64:         encodeInt64CondValue,
		reflect.Uint:   encodeUintValue,
		reflect.Uint8:  encodeUintValue,
		reflect.Uint16: encodeUintValue,
		reflect.Uint32: encodeUintValue,
		reflect.Uint64: encodeUintValue,
		// reflect.Uint8:         encodeUint8CondValue,
		// reflect.Uint16:        encodeUint16CondValue,
		// reflect.Uint32:        encodeUint32CondValue,
		// reflect.Uint64:        encodeUint64CondValue,
		reflect.Float32:       encodeFloat32Value,
		reflect.Float64:       encodeFloat64Value,
		reflect.Complex64:     encodeUnsupportedValue,
		reflect.Complex128:    encodeUnsupportedValue,
		reflect.Array:         encodeArrayValue,
		reflect.Chan:          encodeUnsupportedValue,
		reflect.Func:          encodeUnsupportedValue,
		reflect.Interface:     encodeInterfaceValue,
		reflect.Map:           encodeMapValue,
		reflect.Pointer:       encodeUnsupportedValue, // 여기서는 처리안함
		reflect.Slice:         encodeSliceValue,
		reflect.String:        encodeStringValue,
		reflect.Struct:        encodeStructValue,
		reflect.UnsafePointer: encodeUnsupportedValue,
	}
}

func getEncoder(typ reflect.Type) encoderFunc {
	if v, ok := typeEncMap.Load(typ); ok {
		return v.(encoderFunc)
	}
	fn := _getEncoder(typ)
	typeEncMap.Store(typ, fn)
	return fn
}

// TODO: custom marshal/unmarshal 적용
func _getEncoder(typ reflect.Type) encoderFunc {
	kind := typ.Kind()

	if kind == reflect.Pointer {
		if _, ok := typeEncMap.Load(typ.Elem()); ok {
			return ptrEncoderFunc(typ)
		}
	}

	if typ.Implements(customEncoderType) {
		return encodeCustomValue
	}
	if typ.Implements(marshalerType) {
		return marshalValue
	}
	if typ.Implements(binaryMarshalerType) {
		return marshalBinaryValue
	}
	if typ.Implements(textMarshalerType) {
		return marshalTextValue
	}

	// Addressable struct field value.

	if kind != reflect.Pointer {
		ptr := reflect.PointerTo(typ)
		if ptr.Implements(customEncoderType) {
			return encodeCustomValuePtr
		}
		if ptr.Implements(marshalerType) {
			return marshalValuePtr
		}
		if ptr.Implements(binaryMarshalerType) {
			return marshalBinaryValueAddr
		}
		if ptr.Implements(textMarshalerType) {
			return marshalTextValueAddr
		}
	}

	if typ == errorType {
		return encodeErrorValue
	}

	switch kind {
	case reflect.Pointer:
		return ptrEncoderFunc(typ)
	case reflect.Slice:
		elem := typ.Elem()
		if elem.Kind() == reflect.Uint8 {
			return encodeByteSliceValue
		} else if elem == stringType {
			return encodeStringSliceValue
		}
	case reflect.Array:
		if typ.Elem().Kind() == reflect.Uint8 {
			return encodeByteArrayValue
		}
	case reflect.Map:
		if typ.Key() == stringType {
			switch typ.Elem() {
			case stringType: // [string]string
				return encodeMapStringStringValue
			case boolType: // [string]bool
				return encodeMapStringBoolValue
			case interfaceType: //[string]interface
				return encodeMapStringInterfaceValue
			}
		}
	}

	return valueEncoders[kind]
}

func encodeByteArrayValue(e *Encoder, v reflect.Value) error {
	if err := e.encodeBytesLen(v.Len()); err != nil {
		return err
	}

	if v.CanAddr() {
		// 슬라이스로 변환하여 Byte로 직접 접근하여 인코딩
		b := v.Slice(0, v.Len()).Bytes()
		return e.write(b) // no copy
	}

	// 버퍼로 복사 후 인코딩
	e.buf = grow(e.buf, v.Len())
	reflect.Copy(reflect.ValueOf(e.buf), v)
	return e.write(e.buf)
}

func ptrEncoderFunc(typ reflect.Type) encoderFunc {
	encoder := getEncoder(typ.Elem())
	return func(e *Encoder, v reflect.Value) error {
		if v.IsNil() {
			return e.EncodeNil()
		}
		return encoder(e, v.Elem())
	}
}

func encodeErrorValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}
	return e.EncodeString(v.Interface().(error).Error())
}

func encodeByteSliceValue(e *Encoder, v reflect.Value) error {
	return e.EncodeBytes(v.Bytes())
}

func encodeStringSliceValue(e *Encoder, v reflect.Value) error {
	ss := v.Convert(stringSliceType).Interface().([]string)
	return e.encodeStringSlice(ss)
}

func (e *Encoder) encodeStringSlice(s []string) error {
	if s == nil {
		return e.EncodeNil()
	}
	if err := e.encodeArrayLen(len(s)); err != nil {
		return err
	}
	for _, v := range s {
		if err := e.EncodeString(v); err != nil {
			return err
		}
	}
	return nil
}

func encodeBoolValue(e *Encoder, v reflect.Value) error {
	return e.EncodeBool(v.Bool())
}
func encodeUnsupportedValue(e *Encoder, v reflect.Value) error {
	return fmt.Errorf("hpack: Encode(unsupported %s)", v.Type())
}

func encodeInterfaceValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}
	return e.EncodeValue(v.Elem())
}

func nilable(kind reflect.Kind) bool {
	switch kind {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return true
	}
	return false
}

func encodeCustomValuePtr(e *Encoder, v reflect.Value) error {
	if !v.CanAddr() {
		return fmt.Errorf("hpack: Encode(non-addressable %T)", v.Interface())
	}
	encoder := v.Addr().Interface().(CustomEncoder)
	return encoder.EncodeMsgpack(e)
}

func encodeCustomValue(e *Encoder, v reflect.Value) error {
	if nilable(v.Kind()) && v.IsNil() {
		return e.EncodeNil()
	}

	encoder := v.Interface().(CustomEncoder)
	return encoder.EncodeMsgpack(e)
}

func marshalValuePtr(e *Encoder, v reflect.Value) error {
	if !v.CanAddr() {
		return fmt.Errorf("hpack: Encode(non-addressable %T)", v.Interface())
	}
	return marshalValue(e, v.Addr())
}

func marshalValue(e *Encoder, v reflect.Value) error {
	if nilable(v.Kind()) && v.IsNil() {
		return e.EncodeNil()
	}

	marshaler := v.Interface().(Marshaler)
	b, err := marshaler.MarshalMsgpack()
	if err != nil {
		return err
	}
	_, err = e.w.Write(b)
	return err
}

//------------------------------------------------------------------------------

func marshalBinaryValueAddr(e *Encoder, v reflect.Value) error {
	if !v.CanAddr() {
		return fmt.Errorf("hpack: Encode(non-addressable %T)", v.Interface())
	}
	return marshalBinaryValue(e, v.Addr())
}

func marshalBinaryValue(e *Encoder, v reflect.Value) error {
	if nilable(v.Kind()) && v.IsNil() {
		return e.EncodeNil()
	}

	marshaler := v.Interface().(encoding.BinaryMarshaler)
	data, err := marshaler.MarshalBinary()
	if err != nil {
		return err
	}

	return e.EncodeBytes(data)
}

//------------------------------------------------------------------------------

func marshalTextValueAddr(e *Encoder, v reflect.Value) error {
	if !v.CanAddr() {
		return fmt.Errorf("hpack: Encode(non-addressable %T)", v.Interface())
	}
	return marshalTextValue(e, v.Addr())
}

func marshalTextValue(e *Encoder, v reflect.Value) error {
	if nilable(v.Kind()) && v.IsNil() {
		return e.EncodeNil()
	}

	marshaler := v.Interface().(encoding.TextMarshaler)
	data, err := marshaler.MarshalText()
	if err != nil {
		return err
	}

	return e.EncodeBytes(data)
}
