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
		attemptsForError[err] = attempts
	}

	var (
		n       uint
		emptyT  T
		lastErr error
	)
	shouldRetry := true
	for shouldRetry {
		if n == 0 {
			if config.delay > 0 {
				select {
				case <-ctx.Done():
					return emptyT, ctx.Err()
				case <-config.timer.After(config.delay):
				default:
				}
			} else {
				if err := ctx.Err(); err != nil {
					return emptyT, err
				}
			}
		} else {
			if config.onRetry != nil {
				config.onRetry(n, lastErr)
			}
			if config.backoffFunc != nil {
				select {
				case <-ctx.Done():
					return emptyT, ctx.Err()
				case <-config.timer.After(config.backoffFunc(n)):
				default:
				}
			}
		}

		t, err := action(ctx, n+1)
		if err == nil {
			return t, nil
		}

		lastErr = err
		if config.retryIf != nil && !config.retryIf(lastErr) {
			break
		}
		if config.maxAttempts > 0 && n >= config.maxAttempts {
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
