package mutex

import (
	"context"
	"testing"
	"time"
)

func TestChanMutex(t *testing.T) {
	mu := NewChanMutex()
	mu.Lock()
	defer mu.Unlock()
	if mu.TryLock() {
		t.Errorf("cannot fetch mutex")
	}
}

func TestChanMutexTryLockTimeout(t *testing.T) {
	mu := NewChanMutex()
	mu.Lock()
	go func() {
		time.Sleep(1 * time.Millisecond)
		mu.Unlock()
	}()
	if mu.TryLockWithTimeout(500 * time.Microsecond) {
		t.Errorf("cannot fetch mutex in 500us")
	}
	if !mu.TryLockWithTimeout(5 * time.Millisecond) {
		t.Errorf("should fetch mutex in 5ms")
	}
	mu.Unlock()
}

func TestChanMutexUnlockTwice(t *testing.T) {
	mu := NewChanMutex()
	mu.Lock()
	defer func() {
		if x := recover(); x != nil {
			if x != "lock: unlock of unlocked mutex" {
				t.Errorf("unexpect panic")
			}
		} else {
			t.Errorf("should panic after unlock twice")
		}
	}()
	mu.Unlock()
	mu.Unlock()
}

func TestChanMutexTryLockContext(t *testing.T) {
	mu := NewChanMutex()
	ctx, cancel := context.WithCancel(context.Background())
	mu.Lock()
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	if mu.TryLockWithContext(ctx) {
		t.Errorf("cannot fetch mutex")
	}
}
