package retry

import (
	"errors"
	"fmt"
	"strings"
)

// Error implements a circular buffer to store the last N errors.
// Used when WrapErrorsSize option is set to track historical errors.
type Error struct {
	errors   []error
	capacity int
	nextIdx  int
	isFull   bool
}

func NewError(capacity int) *Error {
	return &Error{
		errors:   make([]error, capacity),
		capacity: capacity,
	}
}

func (e *Error) Add(err error) {
	e.errors[e.nextIdx] = err
	e.nextIdx = (e.nextIdx + 1) % e.capacity

	if !e.isFull && e.nextIdx == 0 {
		e.isFull = true
	}
}

// Error method return string representation of Error
// It is an implementation of error interface
func (e *Error) Error() string {
	logWithNumber := make([]string, len(e.errors))
	for i, l := range e.WrappedErrors() {
		if l != nil {
			logWithNumber[i] = fmt.Sprintf("#%d: %s", i+1, l.Error())
		}
	}

	return fmt.Sprintf("all attempts fail:\n%s", strings.Join(logWithNumber, "\n"))
}

func (e *Error) Is(target error) bool {
	for _, v := range e.WrappedErrors() {
		if errors.Is(v, target) {
			return true
		}
	}
	return false
}

func (e *Error) As(target interface{}) bool {
	for _, v := range e.WrappedErrors() {
		if errors.As(v, target) {
			return true
		}
	}
	return false
}

/*
Unwrap the last error for compatibility with `errors.Unwrap()`.
When you need to unwrap all errors, you should use `WrappedErrors()` instead.

	err := retry.Do(context.Background(),
		func(_ context.Context, _ uint) error {
			return errors.New("original error")
		},
		WrapAllErrors(),
	)

	fmt.Println(errors.Unwrap(err)) # "original error" is printed
*/
func (e Error) Unwrap() error {
	lastIdx := e.nextIdx - 1
	if lastIdx < 0 {
		lastIdx += e.capacity
	}
	return e.errors[lastIdx]
}

// WrappedErrors returns the list of errors that this Error is wrapping.
// It is an implementation of the `errwrap.Wrapper` interface
// in package [errwrap](https://github.com/hashicorp/errwrap) so that
// `retry.Error` can be used with that library.
func (e Error) WrappedErrors() []error {
	if e.isFull {
		return append(e.errors[e.nextIdx:], e.errors[:e.nextIdx]...)
	}
	return e.errors[:e.nextIdx]
}

type unrecoverableError struct {
	err error
}

func (e unrecoverableError) Error() string {
	if e.err == nil {
		return "unrecoverable error"
	}
	return e.err.Error()
}

func (e unrecoverableError) Unwrap() error {
	return e.err
}

// Unrecoverable wraps an error in `unrecoverableError` struct
func Unrecoverable(err error) error {
	return unrecoverableError{err}
}

// IsRecoverable checks if the error is recoverable
// (i.e., not wrapped by Unrecoverable).
func IsRecoverable(err error) bool {
	return !errors.Is(err, unrecoverableError{})
}

// Adds support for errors.Is usage on unrecoverableError
func (unrecoverableError) Is(err error) bool {
	_, isUnrecoverable := err.(unrecoverableError)
	return isUnrecoverable
}

func unpackUnrecoverable(err error) error {
	if unrecoverable, isUnrecoverable := err.(unrecoverableError); isUnrecoverable {
		return unrecoverable.err
	}

	return err
}
