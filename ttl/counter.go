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
	c.records = append(c.records, time.Now())
	if !c.cleaning {
		c.cleaning = true
		go c.startCleanup()
	}
	c.Unlock()
}

func (c *Counter) Len() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.records)
}

// Close remove all data from Counter and exit cleanup.
// Counter cannot use any more after close
func (c *Counter) Close() {
	close(c.shutdown)
	c.Lock()
	c.records = nil
	c.Unlock()
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
	if c.timer == nil {
		c.timer = time.NewTimer(c.ttl)
	} else {
		c.timer.Reset(c.ttl)
	}
	for {
		select {
		case <-c.shutdown:
			c.timer.Stop()
			return
		case <-c.timer.C:
			c.Lock()
			if c.cleanup() {
				c.cleaning = false
				c.Unlock()
				return
			}
			c.Unlock()
		}
	}
}
