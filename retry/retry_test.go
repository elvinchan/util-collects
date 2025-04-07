package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/elvinchan/util-collects/retry/backoff"
)

func TestRetry(t *testing.T) {
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
}

var (
	errMock        = errors.New("mock error")
	errUnrecover   = Unrecoverable(errors.New("unrecoverable error"))
	errAnotherMock = errors.New("another mock error")
)

func TestDo_Success(t *testing.T) {
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

func TestDoWithData_ErrorHandling(t *testing.T) {
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
		attemps uint
		lastErr error
	)

	action := func(ctx context.Context, n uint) error {
		return errMock
	}

	_ = Do(context.Background(), action,
		MaxRetries(3),
		OnRetry(func(n uint, err error) {
			attemps = n
			lastErr = err
		}),
	)

	if attemps != 4 {
		t.Errorf("OnRetry() called %d times, want 4", attemps)
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
				MaxRetries(2),
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

func TestGo_PanicHandling(t *testing.T) {
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

type mockTimer struct {
	ch        chan time.Time
	lastDelay time.Duration
}

func (m *mockTimer) After(d time.Duration) <-chan time.Time {
	m.lastDelay = d
	return m.ch
}
