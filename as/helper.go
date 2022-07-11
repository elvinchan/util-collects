package as

import (
	"bytes"
	"errors"
	"reflect"
	"regexp"
)

func validateEqualArgs(want, got interface{}) error {
	if want == nil && got == nil {
		return nil
	}

	if isFunction(want) || isFunction(got) {
		return errors.New("cannot take func type as argument")
	}
	return nil
}

func isFunction(arg interface{}) bool {
	if arg == nil {
		return false
	}
	return reflect.TypeOf(arg).Kind() == reflect.Func
}

func objectsAreEqual(want, got interface{}) bool {
	if want == nil || got == nil {
		return want == got
	}

	exp, ok := want.([]byte)
	if !ok {
		return reflect.DeepEqual(want, got)
	}

	act, ok := got.([]byte)
	if !ok {
		return false
	}
	if exp == nil || act == nil {
		return exp == nil && act == nil
	}
	return bytes.Equal(exp, act)
}

// PanicTestFunc defines a func that should be passed to the as.Panics and
// as.NotPanics methods, and represents a simple func that takes no arguments,
// and returns nothing.
type PanicTestFunc func()

// didPanic returns true if the function passed to it panics. Otherwise, it
// returns false.
func didPanic(f PanicTestFunc) (didPanic bool, message interface{}) {
	didPanic = true

	defer func() {
		message = recover()
	}()

	// call the target function
	f()
	didPanic = false

	return
}

func regexMatches(regex interface{}, value string) (bool, error) {
	r, ok := regex.(*regexp.Regexp)
	if !ok {
		var err error
		if r, err = regexp.Compile(regex.(string)); err != nil {
			return false, err
		}
	}
	return r.MatchString(value), nil
}
