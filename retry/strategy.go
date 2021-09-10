package retry

import (
	"context"
	"time"
)

type Strategy func(ctx context.Context, attempt uint) bool

// Limit creates a Strategy that limits the number of attempts that Retry will
// make.
func Limit(attemptLimit uint) Strategy {
	return func(ctx context.Context, attempt uint) bool {
		return (attempt < attemptLimit)
	}
}

// Delay creates a Strategy that waits the given duration before the first
// attempt is made.
func Delay(duration time.Duration) Strategy {
	return func(ctx context.Context, attempt uint) bool {
		keep := true
		if attempt == 0 {
			timer := time.NewTimer(duration)
			select {
			case <-timer.C:
			case <-ctx.Done():
				keep = false
			}
			stopTimer(timer)
		}
		return keep
	}
}

// Wait creates a Strategy that waits the given durations for each attempt after
// the first. If the number of attempts is greater than the number of durations
// provided, then the strategy uses the last duration provided.
func Wait(durations ...time.Duration) Strategy {
	return func(ctx context.Context, attempt uint) bool {
		keep := true
		if attempt > 0 && len(durations) > 0 {
			durationIndex := int(attempt - 1)
			if len(durations) <= durationIndex {
				durationIndex = len(durations) - 1
			}
			timer := time.NewTimer(durations[durationIndex])
			select {
			case <-timer.C:
			case <-ctx.Done():
				keep = false
			}
			stopTimer(timer)
		}
		return keep
	}
}

// Backoff creates a Strategy that waits before each attempt, with a duration as
// defined by the given Algorithm.
func Backoff(algorithm Algorithm) Strategy {
	return BackoffJitter(algorithm, func(duration time.Duration) time.Duration {
		return duration
	})
}

// BackoffWithJitter creates a Strategy that waits before each attempt, with a
// duration as defined by the given Algorithm and Transformation.
func BackoffJitter(algorithm Algorithm, transformation Transformation) Strategy {
	return func(ctx context.Context, attempt uint) bool {
		keep := true
		if attempt > 0 {
			timer := time.NewTimer(transformation(algorithm(attempt)))
			select {
			case <-timer.C:
			case <-ctx.Done():
				keep = false
			}
		}
		return keep
	}
}

func BackoffLimit(algorithm Algorithm, limit time.Duration) Strategy {
	return BackoffLimitJitter(algorithm, limit,
		func(duration time.Duration) time.Duration {
			return duration
		},
	)
}

func BackoffLimitJitter(algorithm Algorithm, limit time.Duration,
	transformation Transformation) Strategy {
	return func(ctx context.Context, attempt uint) bool {
		keep := true
		if attempt > 0 {
			t := algorithm(attempt)
			if t > limit {
				t = limit
			}
			timer := time.NewTimer(transformation(t))
			select {
			case <-timer.C:
			case <-ctx.Done():
				keep = false
			}
		}
		return keep
	}
}

func stopTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}
