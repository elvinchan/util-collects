package retry

import (
	"context"
	"testing"
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
