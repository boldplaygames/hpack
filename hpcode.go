package hpack

var (
	PosFixedNumHigh byte = 0x7f // 01111111, 127
	NegFixedNumLow  byte = 0xe0 // 11100000, 224

	Nil byte = 0xc0 // 11000000, 192

	False byte = 0xc2 // 11000010, 194
	True  byte = 0xc3 // 11000011, 195

	Float  byte = 0xca // 11001010, 202
	Double byte = 0xcb // 11001011, 203

	Uint8  byte = 0xcc // 11001100, 204
	Uint16 byte = 0xcd // 11001101, 205
	Uint32 byte = 0xce // 11001110, 206
	Uint64 byte = 0xcf // 11001111, 207

	Int8  byte = 0xd0 // 11010000, 208
	Int16 byte = 0xd1 // 11010001, 209
	Int32 byte = 0xd2 // 11010010, 210
	Int64 byte = 0xd3 // 11010011, 211

	FixedStrLow  byte = 0xa0 // 10100000, 160
	FixedStrHigh byte = 0xbf // 10111111, 191
	FixedStrMask byte = 0x1f // 00011111, 31
	Str8         byte = 0xd9 // 11011001, 217
	Str16        byte = 0xda // 11011010, 218
	Str32        byte = 0xdb // 11011011, 219

	Bin8  byte = 0xc4 // 11000100, 196
	Bin16 byte = 0xc5 // 11000101, 197
	Bin32 byte = 0xc6 // 11000110, 198

	FixedArrayLow  byte = 0x90 // 10010000, 144
	FixedArrayHigh byte = 0x9f // 10011111, 159
	FixedArrayMask byte = 0xf  // 00001111, 15
	Array16        byte = 0xdc // 11011100, 220
	Array32        byte = 0xdd // 11011101, 221

	FixedMapLow  byte = 0x80 // 10000000, 128
	FixedMapHigh byte = 0x8f // 10001111, 143
	FixedMapMask byte = 0xf  // 00001111, 15
	Map16        byte = 0xde // 11011110, 222
	Map32        byte = 0xdf // 11011111, 223

	FixExt1  byte = 0xd4 // 11010100, 212
	FixExt2  byte = 0xd5 // 11010101, 213
	FixExt4  byte = 0xd6 // 11010110, 214
	FixExt8  byte = 0xd7 // 11010111, 215
	FixExt16 byte = 0xd8 // 11011000, 216
	Ext8     byte = 0xc7 // 11000111, 199
	Ext16    byte = 0xc8 // 11001000, 200
	Ext32    byte = 0xc9 // 11001001, 201
)

func IsFixedNum(c byte) bool {
	return c <= PosFixedNumHigh || c >= NegFixedNumLow
}

func IsFixedMap(c byte) bool {
	return c >= FixedMapLow && c <= FixedMapHigh
}

func IsFixedArray(c byte) bool {
	return c >= FixedArrayLow && c <= FixedArrayHigh
}

func IsFixedString(c byte) bool {
	return c >= FixedStrLow && c <= FixedStrHigh
}

func IsString(c byte) bool {
	return IsFixedString(c) || c == Str8 || c == Str16 || c == Str32
}

func IsBin(c byte) bool {
	return c == Bin8 || c == Bin16 || c == Bin32
}

func IsFixedExt(c byte) bool {
	return c >= FixExt1 && c <= FixExt16
}

func IsExt(c byte) bool {
	return IsFixedExt(c) || c == Ext8 || c == Ext16 || c == Ext32
}
