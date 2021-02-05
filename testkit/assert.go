package testkit

import (
	"fmt"
	"testing"
)

type Helper interface {
	Helper()
}

func Assert(tb testing.TB, condition bool, args ...interface{}) {
	if x, ok := tb.(Helper); ok {
		x.Helper() // Go1.9+
	}
	if !condition {
		if msg := fmt.Sprint(args...); msg != "" {
			tb.Fatalf("Assert failed, %s", msg)
		} else {
			tb.Fatalf("Assert failed")
		}
	}
}

func Assertf(tb testing.TB, condition bool, format string, a ...interface{}) {
	if x, ok := tb.(Helper); ok {
		x.Helper() // Go1.9+
	}
	if !condition {
		if msg := fmt.Sprintf(format, a...); msg != "" {
			tb.Fatalf("Assertf failed, %s", msg)
		} else {
			tb.Fatalf("Assertf failed")
		}
	}
}

func AssertFunc(tb testing.TB, fn func() error) {
	if x, ok := tb.(Helper); ok {
		x.Helper() // Go1.9+
	}
	if err := fn(); err != nil {
		tb.Fatalf("AssertFunc failed, %v", err)
	}
}
