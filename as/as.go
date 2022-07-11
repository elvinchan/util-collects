package as

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// As provides check methods around the testing.TB interface.
type As struct {
	testing.TB

	skipCallers  int
	failDirectly bool
	info         strings.Builder
}

// New create a new As object for the specified testing.TB.
func New(t testing.TB, opts ...Option) *As {
	a := &As{
		t,
		3,
		false,
		strings.Builder{},
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Equal checks that two objects are equal.
func (c *As) Equal(want, got interface{}, msgAndArgs ...interface{}) {
	c.Helper()
	if err := validateEqualArgs(want, got); err != nil {
		c.fail("invalid operation", []string{
			fmt.Sprintf("%#v == %#v (%s)", want, got, err),
		}, msgAndArgs...)
		return
	}
	if !objectsAreEqual(want, got) {
		c.fail("should equal", []string{
			fmt.Sprintf("want: %v", want),
			fmt.Sprintf("got: %v", got),
		}, msgAndArgs...)
	}
}

// Equal checks that two objects are not equal.
func (c *As) NotEqual(want, got interface{}, msgAndArgs ...interface{}) {
	c.Helper()
	if err := validateEqualArgs(want, got); err != nil {
		c.fail("invalid operation", []string{
			fmt.Sprintf("%#v == %#v (%s)", want, got, err),
		}, msgAndArgs...)
		return
	}

	if objectsAreEqual(want, got) {
		c.fail("should not equal", []string{
			fmt.Sprintf("got: %v", got),
		}, msgAndArgs...)
	}
}

// Error checks that a function returned an error (i.e. not `nil`).
func (c *As) Error(err error, msgAndArgs ...interface{}) {
	if err != nil {
		return
	}
	c.Helper()
	c.fail("want an error", nil, msgAndArgs...)
}

// NoError checks that a function returned no error (i.e. `nil`).
func (c *As) NoError(err error, msgAndArgs ...interface{}) {
	if err == nil {
		return
	}
	c.Helper()
	c.fail("did not want an error", []string{
		fmt.Sprintf("got: %+v", err),
	}, msgAndArgs...)
}

// Panics checks that the code inside the specified PanicTestFunc panics.
func (c *As) Panics(f PanicTestFunc, msgAndArgs ...interface{}) {
	c.Helper()
	didPanic, panicValue := didPanic(f)
	if !didPanic {
		c.fail("should panic", []string{
			fmt.Sprintf("func: %#v", f),
			fmt.Sprintf("panic value: %#v", panicValue),
		}, msgAndArgs...)
	}
}

// NotPanics checks that the code inside the specified PanicTestFunc does not panic.
func (c *As) NotPanics(f PanicTestFunc, msgAndArgs ...interface{}) {
	c.Helper()
	didPanic, panicValue := didPanic(f)
	if didPanic {
		c.fail("should not panic", []string{
			fmt.Sprintf("func: %#v", f),
			fmt.Sprintf("panic value: %#v", panicValue),
		}, msgAndArgs...)
	}
}

// Regexp checks that a specified regexp matches a string.
func (c *As) Regexp(rx interface{}, str string, msgAndArgs ...interface{}) {
	if ok, err := regexMatches(rx, str); !ok {
		c.Helper()
		if err != nil {
			c.fail("failed compiling regex", []string{
				fmt.Sprintf("%v", rx),
			}, msgAndArgs...)
		} else {
			c.fail("should match", []string{
				fmt.Sprintf("regex: %v", rx),
				fmt.Sprintf("str: %s", str),
			}, msgAndArgs...)
		}
	}
}

// NotRegexp checks that a specified regexp does not match a string.
func (c *As) NotRegexp(rx interface{}, str string, msgAndArgs ...interface{}) {
	if ok, err := regexMatches(rx, str); ok || err != nil {
		c.Helper()
		if err != nil {
			c.fail("failed compiling regex", []string{
				fmt.Sprintf("%v", rx),
			}, msgAndArgs...)
		} else {
			c.fail("should not match", []string{
				fmt.Sprintf("regex: %v", rx),
				fmt.Sprintf("str: %s", str),
			}, msgAndArgs...)
		}
	}
}

// True checks that the specified value is true.
func (c *As) True(ok bool, msgAndArgs ...interface{}) {
	if ok {
		return
	}
	c.Helper()
	c.fail("should be true", nil, msgAndArgs...)
}

// False checks that the specified value is false.
func (c *As) False(ok bool, msgAndArgs ...interface{}) {
	if !ok {
		return
	}
	c.Helper()
	c.fail("should be false", nil, msgAndArgs...)
}

// Eventually checks that given condition will be met in waitFor time,
// periodically checking target function each tick.
func (c *As) Eventually(condition func(ctx context.Context) bool,
	waitFor, tick time.Duration, msgAndArgs ...interface{}) {
	c.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()

	tk := time.NewTimer(tick)
	for {
		select {
		case <-ctx.Done():
			tk.Stop()
			c.fail("condition not satisfied", nil, msgAndArgs...)
			return
		case <-tk.C:
			if condition(ctx) {
				return
			}
			tk.Reset(tick)
		}
	}
}

// Never checks that the given condition doesn't satisfy in waitFor time,
// periodically checking the target function each tick.
func (c *As) Never(condition func(ctx context.Context) bool,
	waitFor, tick time.Duration, msgAndArgs ...interface{}) {
	c.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()

	tk := time.NewTimer(tick)
	for {
		select {
		case <-ctx.Done():
			tk.Stop()
			return
		case <-tk.C:
			if condition(ctx) {
				c.fail("condition satisfied", nil, msgAndArgs...)
				return
			}
			tk.Reset(tick)
		}
	}
}
