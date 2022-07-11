package as_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/elvinchan/util-collects/as"
)

func TestBasic(t *testing.T) {
	as.Equal(t, 1, 1)
	as.NotEqual(t, 1, 2)
	as.Error(t, errors.New("sth wrong"))
	as.NoError(t, nil)
	as.Panics(t, func() {
		panic(0)
	})
	as.NotPanics(t, func() {})
	as.Regexp(t, "^start", "start of the line")
	as.NotRegexp(t, "^asdfastart", "Not the start of the line")
	as.True(t, true)
	as.False(t, false)
	as.Eventually(t, func(ctx context.Context) bool {
		time.Sleep(time.Millisecond * 500)
		return true
	}, time.Second, time.Millisecond*100)
	as.Never(t, func(ctx context.Context) bool {
		time.Sleep(time.Millisecond * 500)
		return false
	}, time.Second, time.Millisecond*100)
}
