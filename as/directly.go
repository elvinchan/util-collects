package as

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/elvinchan/util-collects/sync/atomic"
)

// directlyAs create a new directly failed As object.
func directlyAs(t testing.TB) *As {
	return &As{
		t,
		4,
		atomic.Bool{},
		strings.Builder{},
	}
}

// Equal asserts that two objects are equal.
func Equal(t testing.TB, want, got interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	directlyAs(t).Equal(want, got, msgAndArgs...)
}

// Equal asserts that two objects are not equal.
func NotEqual(t testing.TB, want, got interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	directlyAs(t).NotEqual(want, got, msgAndArgs...)
}

// Error asserts that a function returned an error (i.e. not `nil`).
func Error(t testing.TB, err error, msgAndArgs ...interface{}) {
	t.Helper()
	directlyAs(t).Error(err, msgAndArgs...)
}

// NoError asserts that a function returned no error (i.e. `nil`).
func NoError(t testing.TB, err error, msgAndArgs ...interface{}) {
	t.Helper()
	directlyAs(t).NoError(err, msgAndArgs...)
}

// Panics asserts that the code inside the specified PanicTestFunc panics.
func Panics(t testing.TB, f PanicTestFunc, msgAndArgs ...interface{}) {
	t.Helper()
	directlyAs(t).Panics(f, msgAndArgs...)
}

// NotPanics asserts that the code inside the specified PanicTestFunc does not panic.
func NotPanics(t testing.TB, f PanicTestFunc, msgAndArgs ...interface{}) {
	t.Helper()
	directlyAs(t).NotPanics(f, msgAndArgs...)
}

// Regexp asserts that a specified regexp matches a string.
func Regexp(t testing.TB, rx interface{}, str string, msgAndArgs ...interface{}) {
	t.Helper()
	directlyAs(t).Regexp(rx, str, msgAndArgs...)
}

// NotRegexp asserts that a specified regexp does not match a string.
func NotRegexp(t testing.TB, rx interface{}, str string, msgAndArgs ...interface{}) {
	t.Helper()
	directlyAs(t).NotRegexp(rx, str, msgAndArgs...)
}

// True asserts that the specified value is true.
func True(t testing.TB, ok bool, msgAndArgs ...interface{}) {
	t.Helper()
	directlyAs(t).True(ok, msgAndArgs...)
}

// False asserts that the specified value is false.
func False(t testing.TB, ok bool, msgAndArgs ...interface{}) {
	t.Helper()
	directlyAs(t).False(ok, msgAndArgs...)
}

// Eventually asserts that given condition will be met in waitFor time,
// periodically checking target function each tick.
func Eventually(t testing.TB, condition func(ctx context.Context) bool,
	waitFor, checkInterval time.Duration, msgAndArgs ...interface{}) {
	t.Helper()
	directlyAs(t).Eventually(condition, waitFor, checkInterval, msgAndArgs...)
}

// Never asserts that the given condition doesn't satisfy in waitFor time,
// periodically checking the target function each tick.
func Never(t testing.TB, condition func(ctx context.Context) bool,
	waitFor, checkInterval time.Duration, msgAndArgs ...interface{}) {
	t.Helper()
	directlyAs(t).Never(condition, waitFor, checkInterval, msgAndArgs...)
}
