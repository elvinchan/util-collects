package mutex

import (
	"context"
	"runtime"
	"sync/atomic"
	"time"
)

type CASMutex struct {
	cas uint32
}

// NewCASMutex returns mutex based on CAS.
func NewCASMutex() *CASMutex {
	return &CASMutex{}
}

// Lock locks m.
// If the lock is already in use, the calling goroutine
// blocks until the mutex is available.
func (m *CASMutex) Lock() {
	for !atomic.CompareAndSwapUint32(&m.cas, 0, 1) {
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
	if !atomic.CompareAndSwapUint32(&m.cas, 1, 0) {
		panic("lock: unlock of unlocked mutex")
	}
}

// TryLock attempts to locks m.
// Return false if the lock is in use.
func (m *CASMutex) TryLock() bool {
	return atomic.CompareAndSwapUint32(&m.cas, 0, 1)
}

// TryLockWithContext attempts to lock m, blocking until the mutex is available
// or ctx is done (timeout or cancellation).
func (m *CASMutex) TryLockWithContext(ctx context.Context) bool {
	for !atomic.CompareAndSwapUint32(&m.cas, 0, 1) {
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
