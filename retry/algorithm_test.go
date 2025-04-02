package retry

import "testing"

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
