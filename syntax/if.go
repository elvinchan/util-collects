package syntax

// ref: https://golang.org/doc/faq#Does_Go_have_a_ternary_form

// IIf Immediate if.
// This function should not be used in production since
// interface convertion cost a lot.
//
// Better solution is like:
// func MinInt(x, y int) int { if x < y { return x } return y }
func IIf(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}
