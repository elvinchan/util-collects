package retry

import (
	"time"

	"github.com/elvinchan/util-collects/retry/backoff"
)

type Timer interface {
	After(time.Duration) <-chan time.Time
}

type Config struct {
	retryIf          RetryIfFunc
	onRetry          OnRetryFunc
	maxAttempts      uint
	delay            time.Duration
	attemptsForError map[error]uint
	timer            Timer
	backoffFunc      backoff.BackoffFunc
}

type Option func(*Config)

type RetryIfFunc func(error) bool

// RetryIf controls whether an action should be executed after an error
// (assuming there are any retry attempts remaining)
//
// skip retry if special error example:
//
//	retry.Do(
//		func() error {
//			return errors.New("special error")
//		},
//		retry.RetryIf(func(err error) bool {
//			if err.Error() == "special error" {
//				return false
//			}
//			return true
//		})
//	)
//
// By default RetryIf stops execution if the error is wrapped using `retry.Unrecoverable`,
// so above example may also be shortened to:
//
//	retry.Do(
//		func() error {
//			return retry.Unrecoverable(errors.New("special error"))
//		}
//	)
func RetryIf(f RetryIfFunc) Option {
	return func(c *Config) {
		c.retryIf = f
	}
}

type OnRetryFunc func(retries uint, err error)

// OnRetry function callback are called each retry
//
// log each retry example:
//
//	retry.Do(
//		func() error {
//			return errors.New("some error")
//		},
//		retry.OnRetry(func(retries uint, err error) {
//			log.Printf("#%d: %s\n", retries, err)
//		}),
//	)

func OnRetry(f OnRetryFunc) Option {
	return func(c *Config) {
		c.onRetry = f
	}
}

// MaxRetries set max retries to execute action.
// If set to 0, no retries will be executed.
// If want to retry forever, use `MaxAttempts(0)` instead.
func MaxRetries(m uint) Option {
	return func(c *Config) {
		c.maxAttempts = m + 1
	}
}

// MaxAttempts set max attempts to execute action.
// If set to 0, it will retry forever.
func MaxAttempts(m uint) Option {
	return func(c *Config) {
		c.maxAttempts = m
	}
}

func Delay(d time.Duration) Option {
	return func(c *Config) {
		c.delay = d
	}
}

func AttemptsForError(err error, attempts uint) Option {
	return func(c *Config) {
		if c.attemptsForError == nil {
			c.attemptsForError = make(map[error]uint)
		}
		c.attemptsForError[err] = attempts
	}
}

func IsEventuallyError(err error) Option {
	return AttemptsForError(err, 0)
}

func WithTimer(timer Timer) Option {
	return func(c *Config) {
		c.timer = timer
	}
}

// BackoffSum returns a option that sum the value of all backoff strategies.
func BackoffSum(bfs ...backoff.BackoffFunc) Option {
	return func(c *Config) {
		c.backoffFunc = func(retries uint) time.Duration {
			var sum time.Duration
			for _, bf := range bfs {
				sum += bf(retries)
			}
			return sum
		}
	}
}

// BackoffMin returns a option picks the max value of all backoff strategies.
func BackoffMax(bfs ...backoff.BackoffFunc) Option {
	return func(c *Config) {
		c.backoffFunc = func(retries uint) time.Duration {
			var max time.Duration
			for _, bf := range bfs {
				if d := bf(retries); d > max {
					max = d
				}
			}
			return max
		}
	}
}

// Backoff returns a option that set a single backoff strategy.
func Backoff(bf backoff.BackoffFunc) Option {
	return func(c *Config) {
		c.backoffFunc = bf
	}
}

// BackoffWithJitter returns a option that set a single backoff with jitter strategy.
// It is the shortcut of `Backoff(backoff.WithJitter(bf, jf))`.
func BackoffJitter(bf backoff.BackoffFunc, jf backoff.JitterFunc) Option {
	return func(c *Config) {
		c.backoffFunc = backoff.WithJitter(bf, jf)
	}
}

// BackoffLimit returns a option that set a single backoff with limit strategy.
// It is the shortcut of `Backoff(backoff.WithLimit(bf, limit))`.
func BackoffLimit(bf backoff.BackoffFunc, limit time.Duration) Option {
	return func(c *Config) {
		c.backoffFunc = backoff.WithLimit(bf, limit)
	}
}

// BackoffLimitJitter returns a option that set a single backoff with limit and jitter strategy.
// It is the shortcut of `Backoff(backoff.WithLimitJitter(bf, limit, jf))`.
func BackoffLimitJitter(bf backoff.BackoffFunc, limit time.Duration,
	jf backoff.JitterFunc) Option {
	return func(c *Config) {
		c.backoffFunc = backoff.WithLimitJitter(bf, limit, jf)
	}
}
