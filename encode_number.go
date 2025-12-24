package hpack

import (
	"math"
	"reflect"
)

// EncodeInt encodes an int64 in 1, 2, 3, 5, or 9 bytes.
// Type of the number is lost during encoding.
func (e *Encoder) EncodeInt(n int64) error {
	if n >= 0 {
		return e.EncodeUint(uint64(n))
	}
	// negative fixint stores 5-bit negative integer
	if n >= int64(int8(NegFixedNumLow)) {
		return e.writeCode(byte(n))
		// return e.w.WriteByte(byte(n))
	}
	if n >= math.MinInt8 {
		return e.EncodeInt8(int8(n))
	}
	if n >= math.MinInt16 {
		return e.EncodeInt16(int16(n))
	}
	if n >= math.MinInt32 {
		return e.EncodeInt32(int32(n))
	}
	return e.EncodeInt64(n)
}
func (e *Encoder) EncodeInt8(n int8) error {
	return e.write1(Int8, uint8(n)) // 2 bytes (1+1)
}
func (e *Encoder) EncodeInt16(n int16) error {
	return e.write2(Int16, uint16(n)) // 3 bytes (1+2)
}
func (e *Encoder) EncodeInt32(n int32) error {
	return e.write4(Int32, uint32(n)) // 5 bytes (1+4)
}
func (e *Encoder) EncodeInt64(n int64) error {
	return e.write8(Int64, uint64(n)) // 9 bytes (1+8)
}

// EncodeUint encodes an uint64 in 1, 2, 3, 5, or 9 bytes.
// Type of the number is lost during encoding.
func (e *Encoder) EncodeUint(n uint64) error {
	if n <= math.MaxInt8 {
		// return e.w.WriteByte(byte(n))
		return e.writeCode(byte(n))
	}
	if n <= math.MaxUint8 {
		return e.EncodeUint8(uint8(n))
	}
	if n <= math.MaxUint16 {
		return e.EncodeUint16(uint16(n))
	}
	if n <= math.MaxUint32 {
		return e.EncodeUint32(uint32(n))
	}
	return e.EncodeUint64(n)
}
func (e *Encoder) EncodeUint8(n uint8) error {
	return e.write1(Uint8, n) // 2 bytes (1+1)
}
func (e *Encoder) EncodeUint16(n uint16) error {
	return e.write2(Uint16, n) // 3 bytes (1+2)
}
func (e *Encoder) EncodeUint32(n uint32) error {
	return e.write4(Uint32, n) // 5 bytes (1+4)
}
func (e *Encoder) EncodeUint64(n uint64) error {
	return e.write8(Uint64, n) // 9 bytes (1+8)
}

func (e *Encoder) EncodeFloat32(n float32) error {
	if e.flags&useCompactFloatsFlag != 0 {
		if float32(int64(n)) == n {
			return e.EncodeInt(int64(n))
		}
	}
	return e.write4(Float, math.Float32bits(n))
}
func (e *Encoder) EncodeFloat64(n float64) error {
	if e.flags&useCompactFloatsFlag != 0 {
		// Both NaN and Inf convert to int64(-0x8000000000000000)
		// If n is NaN then it never compares true with any other value
		// If n is Inf then it doesn't convert from int64 back to +/-Inf
		// In both cases the comparison works.
		if float64(int64(n)) == n {
			return e.EncodeInt(int64(n))
		}
	}
	return e.write8(Double, math.Float64bits(n))
}

// Int format family stores an integer in 1, 2, 3, 5, or 9 bytes.
func (e *Encoder) write1(code byte, n uint8) error {
	e.buf = e.buf[:2]
	e.buf[0] = code
	e.buf[1] = n
	return e.write(e.buf)
}

func (e *Encoder) write2(code byte, n uint16) error {
	e.buf = e.buf[:3]
	e.buf[0] = code
	e.buf[1] = byte(n >> 8)
	e.buf[2] = byte(n)
	return e.write(e.buf)
}

func (e *Encoder) write4(code byte, n uint32) error {
	e.buf = e.buf[:5]
	e.buf[0] = code
	e.buf[1] = byte(n >> 24)
	e.buf[2] = byte(n >> 16)
	e.buf[3] = byte(n >> 8)
	e.buf[4] = byte(n)
	return e.write(e.buf)
}

func (e *Encoder) write8(code byte, n uint64) error {
	e.buf = e.buf[:9]
	e.buf[0] = code
	e.buf[1] = byte(n >> 56)
	e.buf[2] = byte(n >> 48)
	e.buf[3] = byte(n >> 40)
	e.buf[4] = byte(n >> 32)
	e.buf[5] = byte(n >> 24)
	e.buf[6] = byte(n >> 16)
	e.buf[7] = byte(n >> 8)
	e.buf[8] = byte(n)
	return e.write(e.buf)
}

// encoderFunc
func encodeUintValue(e *Encoder, v reflect.Value) error {
	return e.EncodeUint(v.Uint())
}

func encodeIntValue(e *Encoder, v reflect.Value) error {
	return e.EncodeInt(v.Int())
}

// func encodeUint8CondValue(e *Encoder, v reflect.Value) error {
// 	return e.encodeUint8Cond(uint8(v.Uint()))
// }

// func encodeUint16CondValue(e *Encoder, v reflect.Value) error {
// 	return e.encodeUint16Cond(uint16(v.Uint()))
// }

// func encodeUint32CondValue(e *Encoder, v reflect.Value) error {
// 	return e.encodeUint32Cond(uint32(v.Uint()))
// }

// func encodeUint64CondValue(e *Encoder, v reflect.Value) error {
// 	return e.encodeUint64Cond(v.Uint())
// }

// func encodeInt8CondValue(e *Encoder, v reflect.Value) error {
// 	return e.encodeInt8Cond(int8(v.Int()))
// }

// func encodeInt16CondValue(e *Encoder, v reflect.Value) error {
// 	return e.encodeInt16Cond(int16(v.Int()))
// }

// func encodeInt32CondValue(e *Encoder, v reflect.Value) error {
// 	return e.encodeInt32Cond(int32(v.Int()))
// }

// func encodeInt64CondValue(e *Encoder, v reflect.Value) error {
// 	return e.encodeInt64Cond(v.Int())
// }

func encodeFloat32Value(e *Encoder, v reflect.Value) error {
	return e.EncodeFloat32(float32(v.Float()))
}

func encodeFloat64Value(e *Encoder, v reflect.Value) error {
	return e.EncodeFloat64(v.Float())
}
