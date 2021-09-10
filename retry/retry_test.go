package retry

import (
	"context"
	"testing"
	"time"
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

func TestLimitAlgorithm(t *testing.T) {
	const (
		duration  = time.Millisecond
		increment = time.Microsecond * 100
		limit     = time.Millisecond * 2
	)

	algorithm := LimitAlgorithm(Incremental(duration, increment), limit)

	for i := uint(0); i < 100; i++ {
		result := algorithm(i)
		expected := duration + (time.Duration(i) * increment)
		if expected > limit {
			expected = limit
		}

		if result != expected {
			t.Errorf("algorithm expected to return a %s duration, but received %s instead", expected, result)
		}
	}
}
