package hpack

import (
	"math"
	"reflect"
	"sort"
)

func encodeMapStringInterfaceValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}
	m := v.Convert(mapStringInterfaceType).Interface().(map[string]interface{})
	if e.flags&sortMapKeysFlag != 0 {
		return e.EncodeMapSorted(m)
	}
	return e.EncodeMap(m)
}

func (e *Encoder) EncodeMap(m map[string]interface{}) error {
	if m == nil {
		return e.EncodeNil()
	}
	if err := e.encodeMapLen(len(m)); err != nil {
		return err
	}
	for mk, mv := range m {
		if err := e.EncodeString(mk); err != nil {
			return err
		}
		if err := e.Encode(mv); err != nil {
			return err
		}
	}
	return nil
}

func encodeMapStringStringValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	if err := e.encodeMapLen(v.Len()); err != nil {
		return err
	}

	m := v.Convert(mapStringStringType).Interface().(map[string]string)
	if e.flags&sortMapKeysFlag != 0 {
		return e.encodeSortedMapStringString(m)
	}

	for mk, mv := range m {
		if err := e.EncodeString(mk); err != nil {
			return err
		}
		if err := e.EncodeString(mv); err != nil {
			return err
		}
	}

	return nil
}

func encodeMapStringBoolValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	if err := e.encodeMapLen(v.Len()); err != nil {
		return err
	}

	m := v.Convert(mapStringBoolType).Interface().(map[string]bool)
	if e.flags&sortMapKeysFlag != 0 {
		return e.encodeSortedMapStringBool(m)
	}

	for mk, mv := range m {
		if err := e.EncodeString(mk); err != nil {
			return err
		}
		if err := e.EncodeBool(mv); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encodeMapLen(l int) error {
	if l < 16 {
		return e.writeCode(FixedMapLow | byte(l))
	}
	if l <= math.MaxUint16 {
		return e.write2(Map16, uint16(l))
	}
	return e.write4(Map32, uint32(l))
}

func encodeMapValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	if err := e.encodeMapLen(v.Len()); err != nil {
		return err
	}

	iter := v.MapRange()
	for iter.Next() {
		if err := e.EncodeValue(iter.Key()); err != nil {
			return err
		}
		if err := e.EncodeValue(iter.Value()); err != nil {
			return err
		}
	}

	return nil
}

func encodeStructValue(e *Encoder, strct reflect.Value) error {
	structFields := structs.Fields(strct.Type())

	fields := structFields.OmitEmpty(e, strct)
	// logc.Trace().Msgf(" <<<<<<<<<<<<<<<< encodeStruct  %s  >>>>>>>>>>>>>>>>>>>> %d/%d", strct.Type().Name(), len(fields), len(structFields.List))

	// logc.Trace().Msgf("getEncoder fields length: %d, struct %s", len(fields), strct.Type().Name())

	if err := e.encodeMapLen(len(fields)); err != nil {
		return err
	}

	for _, f := range fields {
		if err := e.encodeFieldName(f.fieldName); err != nil {
			return err
		}
		// beforeLen := e.w.Len()
		if err := f.EncodeValue(e, strct); err != nil {
			return err
		}
		// encodedSize := e.w.Len() - beforeLen
		// logc.Info().Msgf("%s(%b) => %d bytes [Total: %d]", f.fieldName.name, f.fieldName.hash32, encodedSize, e.w.Len())
	}
	// logc.Trace().Msgf(" <<<<<<<<<<<<<<<< ended %s  >>>>>>>>>>>>>>>>>>>>", strct.Type().Name())

	return nil
}

func (e *Encoder) EncodeMapSorted(m map[string]interface{}) error {
	if m == nil {
		return e.EncodeNil()
	}
	if err := e.encodeMapLen(len(m)); err != nil {
		return err
	}

	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		if err := e.EncodeString(k); err != nil {
			return err
		}
		if err := e.Encode(m[k]); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encodeSortedMapStringBool(m map[string]bool) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		err := e.EncodeString(k)
		if err != nil {
			return err
		}
		if err = e.EncodeBool(m[k]); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encodeSortedMapStringString(m map[string]string) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		err := e.EncodeString(k)
		if err != nil {
			return err
		}
		if err = e.EncodeString(m[k]); err != nil {
			return err
		}
	}

	return nil
}
