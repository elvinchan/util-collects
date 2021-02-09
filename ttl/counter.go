package ttl

import (
	"sync"
	"time"
)

type Counter struct {
	duration time.Duration
	records  []time.Time
	sync.RWMutex
}

func NewCounter(duration time.Duration) *Counter {
	return &Counter{
		duration: duration,
	}
}

func (t *Counter) Incr() {
	t.Lock()
	var start bool
	if t.Len() == 0 {
		start = true
	}
	t.records = append(t.records, time.Now())
	t.Unlock()
	if start {
		go t.cleanup()
	}
}

func (t *Counter) Len() int {
	return len(t.records)
}

func (t *Counter) pop() {
	t.Lock()
	if t.Len() > 0 {
		t.records = t.records[1:]
	}
	t.Unlock()
}

func (t *Counter) get() time.Time {
	var result time.Time
	t.RLock()
	if t.Len() > 0 {
		result = t.records[0]
	}
	t.RUnlock()
	return result
}

func (t *Counter) cleanup() {
	earlyest := t.get()
	if earlyest.IsZero() {
		return
	}
	duration := time.Until(earlyest.Add(t.duration))
	if duration > 0 {
		timer := time.NewTimer(duration)
		<-timer.C
	}
	t.pop()
	go t.cleanup()
}
