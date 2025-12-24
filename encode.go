package hpack

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sync"
	"time"
)

const (
	sortMapKeysFlag uint32 = 1 << iota
	arrayEncodedStructsFlag
	useCompactIntsFlag
	useCompactFloatsFlag
	useInternedStringsFlag
	omitEmptyFlag
)

func Marshal(v interface{}) ([]byte, error) {
	enc := GetEncoder()

	var buf bytes.Buffer
	enc.Reset(&buf)

	err := enc.Encode(v)
	b := buf.Bytes()

	PutEncoder(enc)

	if err != nil {
		return nil, err
	}

	return b, err
}

var encPool = sync.Pool{
	New: func() interface{} {
		return NewEncoder(nil)
	},
}

func GetEncoder() *Encoder {
	return encPool.Get().(*Encoder)
}

func PutEncoder(enc *Encoder) {
	enc.w = nil
	encPool.Put(enc)
}

type writer interface {
	io.Writer
	WriteByte(byte) error
	Len() int
}
type byteWriter struct {
	io.Writer
}

func (bw byteWriter) WriteByte(c byte) error {
	_, err := bw.Write([]byte{c})
	return err
}
func (bw byteWriter) Len() int {
	return bw.Writer.(*bytes.Buffer).Len()
}

type Encoder struct {
	w         writer
	dict      map[string]int
	structTag string
	buf       []byte
	timeBuf   []byte
	flags     uint32
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	e := &Encoder{
		buf: make([]byte, 9),
	}
	e.Reset(w)
	return e
}

// Writer returns the Encoder's writer.
func (e *Encoder) Writer() io.Writer {
	return e.w
}

// Reset discards any buffered data, resets all state, and switches the writer to write to w.
func (e *Encoder) Reset(w io.Writer) {
	e.ResetDict(w, nil)
}
func (e *Encoder) ResetDict(w io.Writer, dict map[string]int) {
	e.ResetWriter(w)
	e.flags = 0
	e.structTag = ""
	e.dict = dict
}
func (e *Encoder) ResetWriter(w io.Writer) {
	// e.dict = nil
	if bw, ok := w.(writer); ok {
		e.w = bw
	} else if w == nil {
		e.w = nil
	} else {
		e.w = newByteWriter(w)
	}
}
func newByteWriter(w io.Writer) byteWriter {
	return byteWriter{
		Writer: w,
	}
}

func (e *Encoder) Encode(v interface{}) error {
	switch v := v.(type) {
	case nil:
		return e.EncodeNil()
	case string:
		return e.EncodeString(v)
	case []byte:
		return e.EncodeBytes(v)
	case int:
		return e.EncodeInt(int64(v))
	case int64:
		// return e.encodeInt64Cond(v) use compact?
		return e.EncodeInt(v)
	case uint:
		return e.EncodeUint(uint64(v))
	case uint64:
		// return e.encodeUint64Cond(v) use compact?
		return e.EncodeUint(v)
	case bool:
		return e.EncodeBool(v)
	case float32:
		return e.EncodeFloat32(v) // use compact?
	case float64:
		return e.EncodeFloat64(v) // use compact?
	case time.Duration:
		// return e.encodeInt64Cond(int64(v)) // use compact?
		return e.EncodeInt(int64(v))
	case time.Time:
		return e.EncodeTime(v)
	}
	return e.EncodeValue(reflect.ValueOf(v))
}

func (e *Encoder) EncodeValue(v reflect.Value) error {
	fn := getEncoder(v.Type())
	return fn(e, v)
}

func (e *Encoder) EncodeNil() error {
	return e.writeCode(Nil)
}
func (e *Encoder) EncodeBool(value bool) error {
	if value {
		return e.writeCode(True)
	}
	return e.writeCode(False)
}

// encodeFieldName: 문자열 필드명을 Header를 붙인 Hashcode로 변환
func (e *Encoder) encodeFieldName(fieldName FieldName) error {
	szFlag := fieldName.GetSizeFlag()
	h32 := fieldName.GetHash32()

	// header(8bit): 필드명size(2bit), 빈 공간(6bit)
	e.writeCode(byte(szFlag))

	buf := make([]byte, 0, szFlag.ToSize())
	// buf := make([]byte, 0, szFlag.ToSize())

	switch szFlag {
	case FieldNameSizeFlag1Byte:
		buf = append(buf, byte(h32&Hex1Byte))

	case FieldNameSizeFlag2Byte:
		buf = append(buf,
			byte((h32>>8)&Hex1Byte),
			byte(h32&Hex1Byte),
		)

	case FieldNameSizeFlag4Byte:
		buf = append(buf,
			byte(h32>>24),
			byte(h32>>16),
			byte(h32>>8),
			byte(h32),
		)
	default:
		return fmt.Errorf("invalid hash size: %d", szFlag)
	}

	// logc.Trace().Msgf("encodeFieldName(%s):: size(%s) hash(%b) %d", fieldName.name, szFlag.ToString(), buf[:], len(buf))

	e.write(buf)

	return nil
}

//

func (e *Encoder) writeCode(c byte) error {
	// logc.Trace().Msgf("Code:: %b(%x)", c, c)
	return e.w.WriteByte(c)
}
func (e *Encoder) write(b []byte) error {
	// logc.Trace().Msgf("write:: %d bytes: %b(%x)", len(b), b, b)
	_, err := e.w.Write(b)
	return err
}
func (e *Encoder) writeString(s string) error {
	// b := StringToBytes(s)
	// logc.Trace().Msgf("writeString:: write %d bytes: %b(%x)", len(b), b, b)

	_, err := e.w.Write(StringToBytes(s))
	return err
}
