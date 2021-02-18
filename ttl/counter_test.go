package ttl

import (
	"testing"
	"time"

	"github.com/elvinchan/util-collects/testkit"
)

func TestCounter(t *testing.T) {
	c := NewCounter(time.Millisecond * 5)
	c.Incr()
	time.Sleep(time.Millisecond * 2)
	c.Incr()
	testkit.Assert(t, c.Len() == 2)
	time.Sleep(time.Millisecond * 5) // not 3 because some latency
	testkit.Assert(t, c.Len() == 1)
	time.Sleep(time.Millisecond * 5)
	testkit.Assert(t, c.Len() == 0)
}
