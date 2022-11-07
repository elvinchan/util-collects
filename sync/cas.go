package sync

import (
	"context"
	"runtime"
	"time"

	"github.com/elvinchan/util-collects/sync/atomic"
)

type CASMutex struct {
	b atomic.Bool
}

// NewCASMutex returns mutex based on CAS.
func NewCASMutex() *CASMutex {
	return &CASMutex{}
}

// Lock locks m.
// If the lock is already in use, the calling goroutine
// blocks until the mutex is available.
func (m *CASMutex) Lock() {
	for !m.b.CompareAndSwap(false, true) {
		runtime.Gosched()
	}
}

// Unlock unlocks m.
// Panics if m is not locked on entry to Unlock.
//
// A locked Mutex is not associated with a particular goroutine.
// It is allowed for one goroutine to lock a Mutex and then
// arrange for another goroutine to unlock it.
func (m *CASMutex) Unlock() {
	if !m.b.CompareAndSwap(true, false) {
		panic("sync: unlock of unlocked mutex")
	}
}

// TryLock attempts to locks m.
// Return false if the lock is in use.
func (m *CASMutex) TryLock() bool {
	return m.b.CompareAndSwap(false, true)
}

// TryLockWithContext attempts to lock m, blocking until the mutex is available
// or ctx is done (timeout or cancellation).
func (m *CASMutex) TryLockWithContext(ctx context.Context) bool {
	for !m.b.CompareAndSwap(false, true) {
		select {
		case <-ctx.Done():
			return false
		default:
			runtime.Gosched()
		}
	}
	return true
}

// TryLockWithContext attempts to lock m, blocking until the mutex is available
// or timeout.
func (m *CASMutex) TryLockWithTimeout(d time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()

	return m.TryLockWithContext(ctx)
}
