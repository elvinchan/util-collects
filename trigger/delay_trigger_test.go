package trigger

import (
	"context"
	"testing"
	"time"

	"github.com/elvinchan/util-collects/as"
)

func TestDelayTrigger(t *testing.T) {
	cnt := 0
	dt := NewDelayTrigger(func() {
		cnt++
	}, time.Millisecond*200)

	// first time, enter without condition
	dt.Trigger()
	as.Eventually(t, func(ctx context.Context) bool {
		return cnt == 1
	}, time.Millisecond*100, time.Millisecond*50)

	// not reach delay
	dt.Trigger()
	as.Never(t, func(ctx context.Context) bool {
		return cnt == 2
	}, time.Millisecond*300, time.Millisecond*50)

	// reach delay
	dt.Trigger()
	as.Eventually(t, func(ctx context.Context) bool {
		return cnt == 2
	}, time.Millisecond*200, time.Millisecond*50)
}
