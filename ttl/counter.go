package ttl

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

type Counter struct {
	mu       sync.RWMutex
	ttl      time.Duration
	ttlList  *list.List
	timer    *time.Timer
	cleaning uint32 // 0 -> false, 1 -> true
	shutdown chan struct{}
}

// NewCounter create a counter with TTL records
func NewCounter(d time.Duration) *Counter {
	return &Counter{
		ttl:     d,
		ttlList: list.New(),
	}
}

func (c *Counter) Incr() {
	c.mu.Lock()
	c.ttlList.PushFront(time.Now().Add(c.ttl))
	c.mu.Unlock()
	if atomic.CompareAndSwapUint32(&c.cleaning, 0, 1) {
		go c.startCleanup()
	}
}

func (c *Counter) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ttlList.Len()
}

// Close remove all data from Counter and exit cleanup.
// Counter cannot be used any more after close
func (c *Counter) Close() {
	close(c.shutdown)
	c.mu.Lock()
	c.ttlList = nil
	c.mu.Unlock()
}

func (t *Counter) pop() {
	t.mu.Lock()
	e := t.ttlList.Back()
	if e != nil {
		t.ttlList.Remove(e)
	}
	t.mu.Unlock()
}

func (t *Counter) get() time.Time {
	var result time.Time
	t.mu.Lock()
	e := t.ttlList.Back()
	if e != nil {
		result = e.Value.(time.Time)
	}
	t.mu.Unlock()
	return result
}

func (c *Counter) cleanup() bool {
	for {
		earlyest := c.get()
		if earlyest.IsZero() {
			return true
		}
		d := time.Until(earlyest)
		if d <= 0 {
			c.pop()
			continue
		}
		c.timer.Reset(d)
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
			if c.cleanup() {
				atomic.StoreUint32(&c.cleaning, 0)
				return
			}
		}
	}
}
