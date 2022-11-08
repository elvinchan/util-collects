package ttl

import (
	"testing"
	"time"

	"github.com/elvinchan/util-collects/as"
)

func TestCounter(t *testing.T) {
	c := NewCounter(time.Millisecond * 5)
	c.Incr()
	time.Sleep(time.Millisecond * 2)
	c.Incr()
	as.Equal(t, c.Len(), 2)
	time.Sleep(time.Millisecond * 5) // not 3 because some latency
	as.Equal(t, c.Len(), 1)
	time.Sleep(time.Millisecond * 5)
	as.Equal(t, c.Len(), 0)
}
