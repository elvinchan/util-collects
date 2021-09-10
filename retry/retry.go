package retry

import (
	"context"
	"fmt"
)

type Action func(ctx context.Context, attempt uint) error

// Do takes an action and performs it, repetitively, until successful.
//
// Optionally, strategies may be passed that assess whether or not an attempt
// should be made.
func Do(ctx context.Context, action Action, strategies ...Strategy) error {
	var err error
	for attempt := uint(0); (attempt == 0 || err != nil) &&
		shouldAttempt(ctx, attempt, strategies...); attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		err = action(ctx, attempt+1)
	}
	return err
}

func shouldAttempt(ctx context.Context, attempt uint, strategies ...Strategy) bool {
	for i := range strategies {
		if !strategies[i](ctx, attempt) {
			return false
		}
	}
	return true
}

// Go takes an action and performs it asynchronous, repetitively, until successful.
//
// Optionally, strategies may be passed that assess whether or not an attempt
// should be made.
func Go(ctx context.Context, action Action, strategies ...Strategy) <-chan error {
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
		done <- Do(ctx, action, strategies...)
	}()
	return done
}
