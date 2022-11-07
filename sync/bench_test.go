package sync

import (
	"sync"
	"testing"
)

func BenchmarkMutexLock(b *testing.B) {
	mu := sync.Mutex{}

	for i := 0; i < b.N; i++ {
		mu.Lock()
		mu.Unlock()
	}
}

func BenchmarkConcurrentMutexLock(b *testing.B) {
	mu := sync.Mutex{}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			mu.Unlock()
		}
	})
}

func BenchmarkCASMutexLock(b *testing.B) {
	mu := NewCASMutex()

	for i := 0; i < b.N; i++ {
		mu.Lock()
		mu.Unlock()
	}
}

func BenchmarkConcurrentCASMutexLock(b *testing.B) {
	mu := NewCASMutex()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			mu.Unlock()
		}
	})
}

func BenchmarkCASMutexTryLock(b *testing.B) {
	mu := NewCASMutex()

	for i := 0; i < b.N; i++ {
		if mu.TryLock() {
			mu.Unlock()
		}
	}
}

func BenchmarkConcurrentCASMutexTryLock(b *testing.B) {
	mu := NewCASMutex()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if mu.TryLock() {
				mu.Unlock()
			}
		}
	})
}

func BenchmarkChanMutexLock(b *testing.B) {
	mu := NewChanMutex()

	for i := 0; i < b.N; i++ {
		mu.Lock()
		mu.Unlock()
	}
}

func BenchmarkConcurrentChanMutexLock(b *testing.B) {
	mu := NewChanMutex()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			mu.Unlock()
		}
	})
}

func BenchmarkChanMutexTryLock(b *testing.B) {
	mu := NewChanMutex()

	for i := 0; i < b.N; i++ {
		if mu.TryLock() {
			mu.Unlock()
		}
	}
}

func BenchmarkConcurrentChanMutexTryLock(b *testing.B) {
	mu := NewChanMutex()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if mu.TryLock() {
				mu.Unlock()
			}
		}
	})
}
