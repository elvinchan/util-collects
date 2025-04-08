package retry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/elvinchan/util-collects/retry/backoff"
)

func TestRetry(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		action := func(ctx context.Context, attempt uint) error {
			return nil
		}

		err := Do(context.Background(), action)
		if err != nil {
			t.Error("expected a nil error")
		}

		err = <-Go(context.Background(), action)
		if err != nil {
			t.Error("expected a nil error")
		}
	})

	t.Run("WithData", func(t *testing.T) {
		action := func(ctx context.Context, n uint) (string, error) {
			return fmt.Sprintf("attempt-%d", n), nil
		}

		result, err := DoWithData(context.Background(), action)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(result, "attempt-") {
			t.Errorf("result = %q, want prefix 'attempt-'", result)
		}
	})
}

var (
	errMock        = errors.New("mock error")
	errUnrecover   = Unrecoverable(errors.New("unrecoverable error"))
	errAnotherMock = errors.New("another mock error")
)

func TestMaxRetries(t *testing.T) {
	attempt := 0
	action := func(ctx context.Context, n uint) error {
		attempt++
		if attempt < 3 {
			return errMock
		}
		return nil
	}

	err := Do(context.Background(), action, MaxRetries(3))
	if err != nil {
		t.Errorf("Do() unexpected error: %v", err)
	}
	if attempt != 3 {
		t.Errorf("Do() attempts = %d, want 3", attempt)
	}
}

func TestErrorHandling(t *testing.T) {
	t.Run("MaxAttempts", func(t *testing.T) {
		action := func(ctx context.Context, attempt uint) (int, error) {
			return 0, errMock
		}

		_, err := DoWithData(context.Background(), action, MaxAttempts(3))
		if !errors.Is(err, errMock) {
			t.Errorf("DoWithData() error = %v, want %v", err, errMock)
		}
	})

	t.Run("UnrecoverableError", func(t *testing.T) {
		attempts := 0
		action := func(ctx context.Context, n uint) (int, error) {
			attempts++
			return 0, errUnrecover
		}

		_, err := DoWithData(context.Background(), action, MaxRetries(3))
		if err == nil {
			t.Errorf("DoWithData() error = %v, want not nil", err)
		}
		if attempts != 1 {
			t.Errorf("DoWithData() attempts = %d, want 1", attempts)
		}
	})
}

func TestRetryIf(t *testing.T) {
	attempts := 0
	action := func(ctx context.Context, n uint) error {
		attempts++
		return errAnotherMock
	}

	err := Do(context.Background(), action,
		MaxRetries(3),
		RetryIf(func(err error) bool {
			return !errors.Is(err, errAnotherMock)
		}),
	)

	if attempts != 1 {
		t.Errorf("RetryIf() attempts = %d, want 1", attempts)
	}
	if !errors.Is(err, errAnotherMock) {
		t.Errorf("RetryIf() error = %v, want %v", err, errAnotherMock)
	}
}

func TestOnRetry(t *testing.T) {
	var (
		retries uint
		lastErr error
	)

	action := func(ctx context.Context, n uint) error {
		return errMock
	}

	_ = Do(context.Background(), action,
		MaxRetries(3),
		OnRetry(func(n uint, err error) {
			retries = n
			lastErr = err
		}),
	)

	if retries != 3 {
		t.Errorf("OnRetry() called %d times, want 3", retries)
	}
	if !errors.Is(lastErr, errMock) {
		t.Errorf("OnRetry() last error = %v, want %v", lastErr, errMock)
	}
}

func TestBackoffStrategies(t *testing.T) {
	mockTimer := &mockTimer{ch: make(chan time.Time)}
	baseDelay := 10 * time.Millisecond

	testCases := []struct {
		name       string
		option     Option
		wantDelays []time.Duration
	}{
		{
			name:   "ConstantBackoff",
			option: Backoff(backoff.Constant(baseDelay)),
			wantDelays: []time.Duration{
				baseDelay, baseDelay, baseDelay,
			},
		},
		{
			name: "BackoffSum",
			option: BackoffSum(
				backoff.Constant(5*time.Millisecond),
				backoff.Constant(5*time.Millisecond),
			),
			wantDelays: []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond},
		},
		{
			name: "BackoffWithJitter",
			option: BackoffJitter(
				backoff.Constant(baseDelay),
				backoff.Full(nil),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var delays []time.Duration
			action := func(ctx context.Context, n uint) error {
				return errMock
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				// Simulate timer advances
				for {
					mockTimer.ch <- time.Time{}
				}
			}()

			_ = Do(ctx, action,
				MaxRetries(3),
				tc.option,
				WithTimer(mockTimer),
				OnRetry(func(n uint, err error) {
					delays = append(delays, mockTimer.lastDelay)
				}),
			)

			if tc.wantDelays != nil {
				if len(delays) != len(tc.wantDelays) {
					t.Fatalf("got %d delays, want %d", len(delays), len(tc.wantDelays))
				}
				for i := range delays {
					if delays[i] != tc.wantDelays[i] {
						t.Errorf("delay[%d] = %v, want %v", i, delays[i], tc.wantDelays[i])
					}
				}
			} else {
				// For jitter test just check range
				for _, d := range delays {
					if d < 0 || d >= baseDelay {
						t.Errorf("jitter delay %v out of [0, %v) range", d, baseDelay)
					}
				}
			}
		})
	}
}

func TestContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	action := func(ctx context.Context, n uint) error {
		if n == 1 {
			cancel()
		}
		return errMock
	}

	err := Do(ctx, action, MaxRetries(3))
	if !errors.Is(err, context.Canceled) {
		t.Errorf("ContextCancel() error = %v, want %v", err, context.Canceled)
	}
}

func TestGoPanicHandling(t *testing.T) {
	action := func(ctx context.Context, n uint) error {
		panic("test panic")
	}

	ch := Go(context.Background(), action)
	err := <-ch

	if err == nil || err.Error() != "retry: unexpected panic: \"test panic\"" {
		t.Errorf("Go() panic handling failed, got: %v", err)
	}
}

func TestAttemptsForError(t *testing.T) {
	attempt := 0
	action := func(ctx context.Context, n uint) error {
		attempt++
		if attempt%2 == 1 {
			return errMock
		} else if attempt%2 == 0 {
			return errAnotherMock
		} else {
			return nil
		}
	}

	err := Do(context.Background(), action,
		AttemptsForError(errMock, 2),
		AttemptsForError(errAnotherMock, 3),
		MaxAttempts(5),
	)

	if attempt != 3 {
		t.Errorf("AttemptsForError() attempts = %d, want 3", attempt)
	}
	if !errors.Is(err, errMock) {
		t.Errorf("AttemptsForError() error = %v, want %v", err, errMock)
	}
}

func TestWrapErrors(t *testing.T) {
	action := func(ctx context.Context, n uint) error {
		return fmt.Errorf("attempt %d error", n)
	}

	err := Do(context.Background(), action,
		MaxRetries(3),
		WrapErrorsSize(2),
	)

	var retryErr *Error
	if !errors.As(err, &retryErr) {
		t.Fatal("expected wrapped errors")
	}

	wrapped := retryErr.WrappedErrors()
	if len(wrapped) != 2 {
		t.Fatalf("got %d wrapped errors, want 2", len(wrapped))
	}

	expected := []string{
		"attempt 3 error",
		"attempt 4 error",
	}
	for i, err := range wrapped {
		if !strings.Contains(err.Error(), expected[i]) {
			t.Errorf("wrapped[%d] = %v, want contains %q", i, err, expected[i])
		}
	}
}

func TestInfiniteRetry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	attempts := 0
	action := func(ctx context.Context, n uint) error {
		attempts++
		return errors.New("transient error")
	}

	err := Do(ctx, action, MaxAttempts(0))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, want %v", err, context.DeadlineExceeded)
	}
	if attempts < 5 {
		t.Errorf("attempts = %d, expected >5 under time constraint", attempts)
	}
}

func TestMultipleErrorAttempts(t *testing.T) {
	var (
		errA = errors.New("error A")
		errB = errors.New("error B")
		errC = errors.New("error C")
	)

	testCases := []struct {
		errSequence  []error
		wantAttempts int
	}{
		{
			errSequence:  []error{errA, errA, errB},
			wantAttempts: 2, // errA:2 attempts, errB:0 attempt
		},
		{
			errSequence:  []error{errB, errB, errB},
			wantAttempts: 1, // errB:1 attempts
		},
		{
			errSequence:  []error{errC, errC, errC},
			wantAttempts: 3, // errC:3 attempts (no limits)
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			attempt := 0
			action := func(ctx context.Context, n uint) error {
				if attempt >= len(tc.errSequence) {
					return nil
				}
				err := tc.errSequence[attempt]
				attempt++
				return err
			}

			_ = Do(context.Background(), action,
				AttemptsForError(errA, 2),
				AttemptsForError(errB, 1),
				AttemptsForError(errC, 0),
				MaxAttempts(3),
			)

			if attempt != tc.wantAttempts {
				t.Errorf("attempts = %d, want %d", attempt, tc.wantAttempts)
			}
		})
	}
}

func TestBackoffLimit(t *testing.T) {
	mockTimer := &mockTimer{ch: make(chan time.Time)}
	maxDelay := 20 * time.Millisecond

	var delays []time.Duration
	action := func(ctx context.Context, n uint) error {
		return errMock
	}

	go func() {
		// Simulate timer advances
		for {
			mockTimer.ch <- time.Time{}
		}
	}()

	_ = Do(context.Background(), action,
		MaxRetries(3),
		BackoffLimit(
			backoff.Linear(10*time.Millisecond),
			maxDelay,
		),
		WithTimer(mockTimer),
		OnRetry(func(n uint, err error) {
			delays = append(delays, mockTimer.lastDelay)
		}),
	)

	for i, d := range delays {
		if d > maxDelay {
			t.Errorf("delay[%d] = %v exceeds limit %v", i, d, maxDelay)
		}
		if i > 0 && d < delays[i-1] {
			t.Errorf("delay[%d] = %v should be larger than previous %v",
				i, d, delays[i-1])
		}
	}
}

type mockTimer struct {
	ch        chan time.Time
	lastDelay time.Duration
}

func (m *mockTimer) After(d time.Duration) <-chan time.Time {
	m.lastDelay = d
	return m.ch
}
