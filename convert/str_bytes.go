package convert

import "unsafe"

// ref: https://github.com/golang/go/issues/25484

// StrToBytes convert string to byte slice.
// note: this should be used only when result byte slice
// not change while s is in used
func StrToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	b := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&b))
}

// BytesToStr convert byte slice to string.
// note: this should be used only when b not change while
// result string is in used
// ref: https://golang.org/src/strings/builder.go#L46
func BytesToStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
