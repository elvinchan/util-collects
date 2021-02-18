package ttl

import (
	"sync"
	"time"
)

type Counter struct {
	sync.RWMutex
	ttl      time.Duration
	records  []time.Time
	timer    *time.Timer
	cleaning bool
	shutdown chan struct{}
}

// NewCounter create a counter with TTL records
func NewCounter(d time.Duration) *Counter {
	return &Counter{
		ttl: d,
	}
}

func (c *Counter) Incr() {
	c.Lock()
	defer c.Unlock()
	c.records = append(c.records, time.Now())
	if !c.cleaning {
		if c.timer == nil {
			c.timer = time.NewTimer(c.ttl)
		} else {
			c.timer.Reset(c.ttl)
		}
		c.cleaning = true
		go c.startCleanup()
	}
}

func (t *Counter) Len() int {
	return len(t.records)
}

func (t *Counter) pop() {
	if t.Len() > 0 {
		t.records = t.records[1:]
	}
}

func (t *Counter) get() time.Time {
	var result time.Time
	if t.Len() > 0 {
		result = t.records[0]
	}
	return result
}

func (c *Counter) cleanup() bool {
	for {
		earlyest := c.get()
		if earlyest.IsZero() {
			return true
		}
		d := time.Until(earlyest.Add(c.ttl))
		if d <= 0 {
			c.pop()
			continue
		}
		c.timer.Reset(c.ttl)
		return false
	}
}

func (c *Counter) startCleanup() {
	for {
		select {
		case <-c.shutdown:
			c.timer.Stop()
			return
		case <-c.timer.C:
			c.Lock()
			if c.cleanup() {
				c.Unlock()
				c.cleaning = false
				return
			}
			c.Unlock()
		}
	}
}
