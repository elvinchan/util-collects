package retry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Function signature of retry action function
type Action func(ctx context.Context, attempt uint) error

// Function signature of retry action function with data
type ActionWithData[T any] func(ctx context.Context, attempt uint) (T, error)

// Default timer is a wrapper around time.After
type timerImpl struct{}

func (t *timerImpl) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func newDefaultConfig() *Config {
	return &Config{
		retryIf:          IsRecoverable,
		attemptsForError: make(map[error]uint),
		timer:            &timerImpl{},
	}
}

// Do takes an action and performs it, repetitively, until successful.
func Do(ctx context.Context, action Action, options ...Option) error {
	actionWithData := func(ctx context.Context, attempt uint) (any, error) {
		return nil, action(ctx, attempt)
	}

	_, err := DoWithData(ctx, actionWithData, options...)
	return err
}

// DoWithData takes an action that returns some value and performs it,
// repetitively, until successful.
func DoWithData[T any](ctx context.Context, action ActionWithData[T], options ...Option) (T, error) {
	config := newDefaultConfig()
	for _, opt := range options {
		opt(config)
	}

	attemptsForError := make(map[error]uint, len(config.attemptsForError))
	for err, attempts := range config.attemptsForError {
		if attempts > 0 {
			attemptsForError[err] = attempts
		}
	}

	var (
		n       uint
		emptyT  T
		lastErr error
		wrapErr *Error
	)
	if config.wrapErrorsSize > 0 {
		wrapErr = NewError(config.wrapErrorsSize)
	}
	shouldRetry := true
	for shouldRetry {
		// prepare exec action
		var delay time.Duration
		if n == 0 && config.delay > 0 {
			delay = config.delay
		} else if n > 0 && config.backoffFunc != nil {
			delay = config.backoffFunc(n)
		}
		if delay > 0 {
			select {
			case <-ctx.Done():
				if wrapErr != nil {
					wrapErr.Add(ctx.Err())
					return emptyT, wrapErr
				}
				return emptyT, ctx.Err()
			case <-config.timer.After(delay):
			}
		} else {
			if err := ctx.Err(); err != nil {
				if wrapErr != nil {
					wrapErr.Add(err)
					return emptyT, wrapErr
				}
				return emptyT, err
			}
		}
		if n > 0 && config.onRetry != nil {
			config.onRetry(n, lastErr)
		}

		// exec action
		t, err := action(ctx, n+1)
		if err == nil {
			return t, nil
		}

		// handle error, prepare next iteration
		lastErr = unpackUnrecoverable(err)
		if wrapErr != nil {
			wrapErr.Add(lastErr)
		}

		if !config.retryIf(err) {
			break
		}
		if config.maxAttempts > 0 && n+1 >= config.maxAttempts {
			break
		}
		for errToCheck, attempts := range attemptsForError {
			if errors.Is(lastErr, errToCheck) {
				attempts--
				attemptsForError[errToCheck] = attempts
				shouldRetry = shouldRetry && attempts > 0
			}
		}
		n++
	}
	if wrapErr != nil {
		return emptyT, wrapErr
	}
	return emptyT, lastErr
}

// Go takes an action and performs it asynchronous, repetitively, until successful.
func Go(ctx context.Context, action Action, options ...Option) <-chan error {
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("retry: unexpected panic: %#v", r)
				}
				done <- err
			}
			close(done)
		}()
		done <- Do(ctx, action, options...)
	}()
	return done
}
