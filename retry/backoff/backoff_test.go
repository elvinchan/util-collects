package backoff

import (
	"testing"
	"time"
)

func TestConstant(t *testing.T) {
	duration := 100 * time.Millisecond
	b := Constant(duration)

	tests := []struct {
		retries uint
		want    time.Duration
	}{
		{0, duration},
		{1, duration},
		{5, duration},
	}

	for _, tt := range tests {
		if got := b(tt.retries); got != tt.want {
			t.Errorf("Constant()(%d) = %v, want %v", tt.retries, got, tt.want)
		}
	}
}

func TestExplicit(t *testing.T) {
	durations := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		500 * time.Millisecond,
	}
	b := Explicit(durations...)

	tests := []struct {
		retries uint
		want    time.Duration
	}{
		{0, durations[0]},
		{1, durations[1]},
		{2, durations[2]},
		{3, durations[2]}, // Exceed index
		{5, durations[2]}, // Exceed index
	}

	for _, tt := range tests {
		if got := b(tt.retries); got != tt.want {
			t.Errorf("Explicit()(%d) = %v, want %v", tt.retries, got, tt.want)
		}
	}
}

func TestIncremental(t *testing.T) {
	initial := 100 * time.Millisecond
	increment := 50 * time.Millisecond
	b := Incremental(initial, increment)

	tests := []struct {
		retries uint
		want    time.Duration
	}{
		{0, initial + increment*0},
		{1, initial + increment*1},
		{3, initial + increment*3},
		{5, initial + increment*5},
	}

	for _, tt := range tests {
		if got := b(tt.retries); got != tt.want {
			t.Errorf("Incremental()(%d) = %v, want %v", tt.retries, got, tt.want)
		}
	}
}

func TestLinear(t *testing.T) {
	factor := 50 * time.Millisecond
	b := Linear(factor)

	tests := []struct {
		retries uint
		want    time.Duration
	}{
		{0, 0}, // 50ms * 0
		{1, factor * 1},
		{4, factor * 4},
		{7, factor * 7},
	}

	for _, tt := range tests {
		if got := b(tt.retries); got != tt.want {
			t.Errorf("Linear()(%d) = %v, want %v", tt.retries, got, tt.want)
		}
	}
}

func TestExponential(t *testing.T) {
	factor := 100 * time.Millisecond
	base := 2.0
	b := Exponential(factor, base)

	tests := []struct {
		retries uint
		want    time.Duration
	}{
		{0, factor * 1}, // 2^0=1
		{1, factor * 2}, // 2^1=2
		{3, factor * 8}, // 2^3=8
		{5, factor * 32},
	}

	for _, tt := range tests {
		if got := b(tt.retries); got != tt.want {
			t.Errorf("Exponential()(%d) = %v, want %v", tt.retries, got, tt.want)
		}
	}
}

func TestBinaryExponential(t *testing.T) {
	factor := 100 * time.Millisecond
	b := BinaryExponential(factor)

	// Should behave same as Exponential with base=2
	tests := []struct {
		retries uint
		want    time.Duration
	}{
		{0, factor * 1},
		{2, factor * 4},
		{4, factor * 16},
	}

	for _, tt := range tests {
		if got := b(tt.retries); got != tt.want {
			t.Errorf("BinaryExponential()(%d) = %v, want %v", tt.retries, got, tt.want)
		}
	}
}

func TestFibonacci(t *testing.T) {
	factor := 100 * time.Millisecond
	b := Fibonacci(factor)

	tests := []struct {
		retries uint
		want    time.Duration
	}{
		{0, 0 * factor},  // fib(0)=0
		{1, 1 * factor},  // fib(1)=1
		{2, 1 * factor},  // fib(2)=1
		{3, 2 * factor},  // fib(3)=2
		{5, 5 * factor},  // fib(5)=5
		{7, 13 * factor}, // fib(7)=13
	}

	for _, tt := range tests {
		if got := b(tt.retries); got != tt.want {
			t.Errorf("Fibonacci()(%d) = %v, want %v", tt.retries, got, tt.want)
		}
	}
}

// TestFibonacciNumberBoundaryConditions tests edge cases
func TestFibonacciNumberBoundaryConditions(t *testing.T) {
	tests := []struct {
		name     string
		input    uint
		expected uint
	}{
		{"n=0 should return 0", 0, 0},
		{"n=1 should return 1", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fibonacciNumber(tt.input); got != tt.expected {
				t.Errorf("fibonacciNumber(%d) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

// TestFibonacciNumberNormalValues tests typical Fibonacci sequence values
func TestFibonacciNumberNormalValues(t *testing.T) {
	tests := []struct {
		name     string
		input    uint
		expected uint
	}{
		{"n=2 should return 1", 2, 1},
		{"n=3 should return 2", 3, 2},
		{"n=5 should return 5", 5, 5},
		{"n=7 should return 13", 7, 13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fibonacciNumber(tt.input); got != tt.expected {
				t.Errorf("fibonacciNumber(%d) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

// TestFibonacciNumberLargeInput tests larger values for performance validation
func TestFibonacciNumberLargeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    uint
		expected uint
	}{
		{"n=10 should return 55", 10, 55},
		{"n=15 should return 610", 15, 610},
		{"n=20 should return 6765", 20, 6765},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fibonacciNumber(tt.input); got != tt.expected {
				t.Errorf("fibonacciNumber(%d) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

// TestFibonacciNumberOverflow tests potential overflow scenarios
// Note: Actual test values should be adjusted based on uint size
func TestFibonacciNumberOverflow(t *testing.T) {
	tests := []struct {
		name     string
		input    uint
		expected uint
	}{
		{"n=50 should return 12586269025", 50, 12586269025},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fibonacciNumber(tt.input); got != tt.expected {
				t.Errorf("fibonacciNumber(%d) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

// BenchmarkFibonacciNumber measures performance characteristics
func BenchmarkFibonacciNumber(b *testing.B) {
	benchmarks := []struct {
		name  string
		input uint
	}{
		{"n=10", 10},
		{"n=20", 20},
		{"n=30", 30},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				fibonacciNumber(bm.input)
			}
		})
	}
}
