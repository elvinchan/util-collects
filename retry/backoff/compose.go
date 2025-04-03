package backoff

import (
	"time"
)

// WithJitter composes BackoffFunc and JitterFunc into a new BackoffFunc
func WithJitter(backoffFunc BackoffFunc, jitterFunc JitterFunc) BackoffFunc {
	return func(retries uint) time.Duration {
		return jitterFunc(backoffFunc(retries))
	}
}

// WithLimit caps the maximum value of the given BackoffFunc as a new BackoffFunc
func WithLimit(backoffFunc BackoffFunc, limit time.Duration) BackoffFunc {
	return WithLimitJitter(backoffFunc, limit,
		func(duration time.Duration) time.Duration {
			return duration
		},
	)
}

// WithLimitJitter composes BackoffFunc with maximum value of backoff
// and JitterFunc into a new BackoffFunc
func WithLimitJitter(backoffFunc BackoffFunc, limit time.Duration,
	jitterFunc JitterFunc) BackoffFunc {
	return func(retries uint) time.Duration {
		t := backoffFunc(retries)
		if t > limit {
			t = limit
		}
		return jitterFunc(t)
	}
}
