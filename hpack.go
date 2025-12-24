// Package hpack
// 필드명을 CRC32Hash로 대체한 Serializer
// Hash의 크기는 1B(중복 시 2B -> 4B 순서로 해시값 재할당 시도)
package hpack

import (
	"fmt"
	"hash/crc32"
	"reflect"
	"sync"
)

// FieldNameSizeFlag : 필드명 크기 플래그
type FieldNameSizeFlag byte

const (
	FieldNameSizeFlag1Byte FieldNameSizeFlag = 0x00 // (00000000)
	FieldNameSizeFlag2Byte FieldNameSizeFlag = 0x40 // (01000000)
	FieldNameSizeFlag4Byte FieldNameSizeFlag = 0x80 // (10000000)
)

func (f FieldNameSizeFlag) ToString() string {
	switch f {
	case FieldNameSizeFlag1Byte:
		return "1Byte"
	case FieldNameSizeFlag2Byte:
		return "2Byte"
	case FieldNameSizeFlag4Byte:
		return "4Byte"
	}

	fmt.Errorf("unknown FieldNameSizeFlag: %v", f)
	return "Unknown"
}
func (f FieldNameSizeFlag) ToSize() int {
	switch f {
	case FieldNameSizeFlag1Byte:
		return 1
	case FieldNameSizeFlag2Byte:
		return 2
	case FieldNameSizeFlag4Byte:
		return 4
	}

	fmt.Errorf("unknown FieldNameSizeFlag: %v", f)
	return 0
}

var fieldNameSizeFlagValues = []FieldNameSizeFlag{
	FieldNameSizeFlag1Byte,
	FieldNameSizeFlag2Byte,
	FieldNameSizeFlag4Byte,
}

const (
	Hex1Byte uint32 = 0xFF
	// Hex2Byte uint32 = 0xFFFF
	// Hex4Byte uint32 = 0xFFFFFFFF
)

const defaultStructTag = "msgpack"

var structs = newStructCache()

type structCache struct {
	m sync.Map // fields
}

type structCacheKey struct {
	typ reflect.Type
}

func newStructCache() *structCache {
	return new(structCache)
}

func (m *structCache) Fields(typ reflect.Type) *fields {
	key := structCacheKey{typ: typ}

	if v, ok := m.m.Load(key); ok {
		return v.(*fields)
	}

	fs := getFields(typ)
	m.m.Store(key, fs)

	return fs
}

type fields struct {
	Type reflect.Type
	Map  map[uint32]*Field
	List []*Field

	hasOmitEmpty bool
}

// IEEE: 가장 흔한 CRC-32 (0x04C11DB7)
var crc32Table = crc32.MakeTable(crc32.IEEE)

// CRC32Hash : 문자열을 IEEE 802.3 표준 CRC-32 해시로 변환
func CRC32Hash(s string) uint32 {
	return crc32.Checksum([]byte(s), crc32Table)
}

// ------------------------------------------------------------------------------
type Marshaler interface {
	MarshalMsgpack() ([]byte, error)
}

type Unmarshaler interface {
	UnmarshalMsgpack([]byte) error
}

type CustomEncoder interface {
	EncodeMsgpack(*Encoder) error
}

type CustomDecoder interface {
	DecodeMsgpack(*Decoder) error
}

// ------------------------------------------------------------------------------
type RawMessage []byte

var (
	_ CustomEncoder = (RawMessage)(nil)
	_ CustomDecoder = (*RawMessage)(nil)
)

func (m RawMessage) EncodeMsgpack(enc *Encoder) error {
	return enc.write(m)
}

func (m *RawMessage) DecodeMsgpack(dec *Decoder) error {
	msg, err := dec.DecodeRaw()
	if err != nil {
		return err
	}
	*m = msg
	return nil
}

//------------------------------------------------------------------------------

type unexpectedCodeError struct {
	hint string
	code byte
}

func (err unexpectedCodeError) Error() string {
	return fmt.Sprintf("hpack: unexpected code=%x decoding %s", err.code, err.hint)
}
