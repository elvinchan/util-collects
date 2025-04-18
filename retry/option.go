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
	attemptsForError map[error]uint
	delay            time.Duration
	wrapErrorsSize   int
	timer            Timer
	backoffFunc      backoff.BackoffFunc
}

type Option func(*Config)

func emptyOption(c *Config) {}

type RetryIfFunc func(error) bool

// RetryIf controls whether an action should be executed after an error
// (assuming there are any retry attempts remaining)
//
// skip retry if special error example:
//
//	retry.Do(context.Background(),
//		func(_ context.Context, _ uint) error {
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
//	retry.Do(context.Background(),
//		func(_ context.Context, _ uint) error {
//			return retry.Unrecoverable(errors.New("special error"))
//		}
//	)
func RetryIf(f RetryIfFunc) Option {
	if f == nil {
		return emptyOption
	}
	return func(c *Config) {
		c.retryIf = f
	}
}

type OnRetryFunc func(retries uint, err error)

// OnRetry function callback are called each retry after first attempt
//
// log each retry example:
//
//	retry.Do(context.Background(),
//		func(_ context.Context, _ uint) error {
//			return errors.New("some error")
//		},
//		retry.OnRetry(func(retries uint, err error) {
//			log.Printf("#%d: %s\n", retries, err)
//		}),
//	)

func OnRetry(f OnRetryFunc) Option {
	if f == nil {
		return emptyOption
	}
	return func(c *Config) {
		c.onRetry = f
	}
}

// MaxRetries sets the maximum number of retries (excluding initial attempt).
// Example: MaxRetries(3) allows 1 initial attempt + 3 retries = 4 total attempts.
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

// AttemptsForError sets count of attempts in case execution results in given `err`.
// Attempts for the given `err` are also counted against total attempts.
// The retry will stop if any of given attempts is exhausted.
func AttemptsForError(err error, attempts uint) Option {
	return func(c *Config) {
		c.attemptsForError[err] = attempts
	}
}

// Delay set delay for first attempt.
// If you want dalay for following attempts, see `Backoff`.
func Delay(d time.Duration) Option {
	return func(c *Config) {
		c.delay = d
	}
}

// WrapErrorsSize set the size of error wrapping stack.
// If set to 0, it will not wrap errors, or it will wrap errors with `retry.Error`.
// Default is 0.
func WrapErrorsSize(s int) Option {
	return func(c *Config) {
		c.wrapErrorsSize = s
	}
}

// WithTimer provides a way to swap out timer module implementations.
// This primarily is useful for mocking/testing, where you may not want to explicitly wait for a set duration
// for retries.
//
// example of augmenting time.After with a print statement
//
//	type struct MyTimer {}
//
//	func (t *MyTimer) After(d time.Duration) <- chan time.Time {
//	    fmt.Print("Timer called!")
//	    return time.After(d)
//	}
//
//	retry.Do(context.Background(),
//	    func(_ context.Context, _ uint) error { ... },
//		   retry.WithTimer(&MyTimer{})
//	)
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
