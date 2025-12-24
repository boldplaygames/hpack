package hpack

import "unsafe"

// bytesToString converts byte slice to string.
// WARNING: The returned string shares the same memory with the input byte slice.
// Modifying the input byte slice may lead to undefined behavior.
//
//	func bytesToString(b []byte) string {
//		return *(*string)(unsafe.Pointer(&b))
//	}
func bytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

// StringToBytes converts string to byte slice.
// WARNING: The returned byte slice shares the same memory with the input string.
// Modifying the returned byte slice may lead to undefined behavior.
//
// func StringToBytes(s string) []byte {
// 	return *(*[]byte)(unsafe.Pointer(
// 		&struct {
// 			string
// 			Cap int
// 		}{s, len(s)},
// 	))
// }

func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
