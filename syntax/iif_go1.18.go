//go:build go1.18
// +build go1.18

package syntax

func IIf[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}
