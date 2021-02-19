package mutex

import (
	"context"
	"time"
)

type ChanMutex struct {
	ch chan struct{}
}

// NewChanMutex returns mutex based on channel.
func NewChanMutex() *ChanMutex {
	return &ChanMutex{
		ch: make(chan struct{}, 1),
	}
}

// Lock locks m.
// If the lock is already in use, the calling goroutine
// blocks until the mutex is available.
func (m *ChanMutex) Lock() {
	m.ch <- struct{}{}
}

// Unlock unlocks m.
// Panics if m is not locked on entry to Unlock.
//
// A locked Mutex is not associated with a particular goroutine.
// It is allowed for one goroutine to lock a Mutex and then
// arrange for another goroutine to unlock it.
func (m *ChanMutex) Unlock() {
	select {
	case <-m.ch:
	default:
		panic("lock: unlock of unlocked mutex")
	}
}

// TryLock attempts to locks m.
// Return false if the lock is in use.
func (m *ChanMutex) TryLock() bool {
	select {
	case m.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

// TryLockWithContext attempts to lock m, blocking until the mutex is available
// or ctx is done (timeout or cancellation).
func (m *ChanMutex) TryLockWithContext(ctx context.Context) bool {
	select {
	case m.ch <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

// TryLockWithContext attempts to lock m, blocking until the mutex is available
// or timeout.
func (m *ChanMutex) TryLockWithTimeout(d time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()

	return m.TryLockWithContext(ctx)
}
