package backoff

import (
	"math"
	"time"
)

type BackoffFunc func(retries uint) time.Duration

// Constant creates a BackoffFunc that returns the same duration for every retries.
func Constant(duration time.Duration) BackoffFunc {
	return func(_ uint) time.Duration {
		return duration
	}
}

// Exponential creates a BackoffFunc that exponentially increases the base
func Explicit(durations ...time.Duration) BackoffFunc {
	return func(retries uint) time.Duration {
		if len(durations)-1 < int(retries) {
			return durations[len(durations)-1]
		}
		return durations[retries]
	}
}

// Incremental creates a BackoffFunc that increments the initial duration
// by the given increment for each retries.
func Incremental(initial, increment time.Duration) BackoffFunc {
	return func(retries uint) time.Duration {
		return initial + (increment * time.Duration(retries))
	}
}

// Linear creates a BackoffFunc that linearly multiplies the factor
// duration by the retries number for each retries.
func Linear(factor time.Duration) BackoffFunc {
	return func(retries uint) time.Duration {
		return (factor * time.Duration(retries))
	}
}

// Exponential creates a BackoffFunc that multiplies the factor duration by
// an exponentially increasing factor for each retries, where the factor is
// calculated as the given base raised to the retries number.
func Exponential(factor time.Duration, base float64) BackoffFunc {
	return func(retries uint) time.Duration {
		return (factor * time.Duration(math.Pow(base, float64(retries))))
	}
}

// BinaryExponential creates a BackoffFunc that multiplies the factor
// duration by an exponentially increasing factor for each retries, where the
// factor is calculated as `2` raised to the retries number (2^retries).
func BinaryExponential(factor time.Duration) BackoffFunc {
	return Exponential(factor, 2)
}

// Fibonacci creates a BackoffFunc that multiplies the factor duration by
// an increasing factor for each retries, where the factor is the Nth number in
// the Fibonacci sequence.
func Fibonacci(factor time.Duration) BackoffFunc {
	return func(retries uint) time.Duration {
		return (factor * time.Duration(fibonacciNumber(retries)))
	}
}

func fibonacciNumber(n uint) uint {
	if n <= 1 {
		return n
	}
	a, b := uint(0), uint(1)
	for i := uint(2); i <= n; i++ {
		a, b = b, a+b
	}
	return b
}
