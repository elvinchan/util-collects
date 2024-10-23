package trigger

import (
	"sync"
	"time"
)

type DelayTrigger struct {
	fn         func()
	delay      time.Duration
	mu         sync.Mutex
	lastEnter  int64
	isEntering bool
}

// NewDelayTrigger create a non block delay trigger.
// The fn only execute when it is not entering and the duration since last
// entered is larger than delay.
func NewDelayTrigger(fn func(), delay time.Duration) *DelayTrigger {
	return &DelayTrigger{
		fn:    fn,
		delay: delay,
	}
}

func (dt *DelayTrigger) Trigger() {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	now := time.Now()
	if dt.delay != 0 && time.Unix(0, dt.lastEnter).Add(dt.delay).After(now) {
		return
	}
	if dt.isEntering {
		return
	}
	dt.isEntering = true
	dt.lastEnter = now.UnixNano()
	go dt.exec()
}

func (dt *DelayTrigger) exec() {
	defer func() {
		dt.mu.Lock()
		dt.isEntering = false
		dt.mu.Unlock()
	}()
	dt.fn()
}
