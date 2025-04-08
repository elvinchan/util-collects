# Retry

Robust Go retry library with context support, flexible backoff strategies, and comprehensive error handling.

## Inspired by
- [github.com/avast/retry-go](https://github.com/avast/retry-go)  
but with more backoff strategies, limited size of wrapped errors, etc.

- [github.com/Rican7/retry](https://github.com/Rican7/retry)  
but with more options like conditional retry, on retry hooks, etc.

## Features

- **Multiple Backoff Strategies**  
  Constant/Linear/Exponential/Fibonacci with jitter support
- **Smart Error Handling**  
  Circular error buffer & unrecoverable error detection
- **Context Integration**  
  Full context cancellation support
- **Customizable Policies**  
  Conditional retry, max attempts, per-error limits
- **Metrics & Observability**  
  Built-in retry attempt hooks
- **Concurrency Safe**  
  Async execution support via `Go()` method

# Quick Start
```go
package main

import (
	"context"
	"fmt"
	"time"
	
	"github.com/elvinchan/util-collects/retry"
	"github.com/elvinchan/util-collects/retry/backoff"
)

func main() {
	ctx := context.Background()
	
	// basic retry with exponential backoff, up to 3 attempts.
	err := retry.Do(ctx, func(ctx context.Context, attempt uint) error {
		fmt.Printf("Attempt %d\n", attempt)
		return fmt.Errorf("transient error")
	}, retry.MaxAttempts(3), 
		retry.Backoff(backoff.BinaryExponential(100*time.Millisecond)))
	
	fmt.Println("Final error:", err)
}
```

# Core Concepts
## 1. Error Handling
```go
// wrap last 3 errors
retry.Do(ctx, action, retry.WrapErrorsSize(3))

// mark unrecoverable error
retry.Do(ctx, func(ctx context.Context, _ uint) error {
	return retry.Unrecoverable(errors.New("fatal"))
})
```

## 2. Backoff Strategies
| Strategy | Description |
| --- | --- |
| `Constant` | Fixed delay between retries |
| `Linear` | Linear time increase: `factor * retries` |
| `Exponential` | Exponential growth: `factor * base^retries` |
| `Fibonacci` | Fibonacci sequence-based delays |

### With Jitter:
```go
retry.BackoffJitter(
	backoff.Linear(1*time.Second),
	backoff.Deviation(nil, 0.3), // ±30% jitter
)
```

## 3. Advanced Control
```go
retry.Do(ctx, action,
	// policy configuration
	retry.MaxAttempts(5),          // total attempts
	retry.Delay(500*time.Millisecond), // initial delay (only for first attempt)
	
	// error-specific limits
	retry.AttemptsForError(errTimeout, 3),
	
	// conditional retry
	retry.RetryIf(func(err error) bool {
		return !isFatalError(err)
	}),
	
	// monitoring hooks
	retry.OnRetry(func(n uint, err error) {
		metrics.RecordRetry(n, err)
	}),
)
```

# Best Practices
## 1. Context Integration
```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

retry.Do(ctx, func(ctx context.Context, _ uint) error {
	return database.Ping(ctx) // propagate context
})
```

## 2. Async Operations
```go
resultChan := retry.Go(ctx, asyncAction, retry.MaxAttempts(3))

select {
case err := <-resultChan:
	// handle final result
case <-ctx.Done():
	// handle cancellation
}
```

## 3. Custom Backoff
```go
// hybrid strategy: Linear + Fibonacci
customBackoff := backoff.Sum(
	backoff.Linear(1*time.Second),
	backoff.Fibonacci(500*time.Millisecond),
)

retry.Do(ctx, action, retry.Backoff(customBackoff))
```
