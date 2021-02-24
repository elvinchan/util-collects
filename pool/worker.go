package pool

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultIdleTimeout = 5 * time.Second
)

func defaultPanicHandler(panic interface{}) {
	fmt.Printf("pool: worker exits from a panic: %v\nStack trace: %s\n",
		panic, string(debug.Stack()))
}

type TaskFunc func() error

type ResizingStrategy interface {
	// Resize return true if should resize
	Resize(runningWorkers int) bool
}

type WorkOption func(wp *WorkerPool)

// WorkerWithMinWorkers set minimum number of workers for a worker pool
func WorkerWithMinWorkers(n int) WorkOption {
	return func(wp *WorkerPool) {
		wp.minWorkers = n
	}
}

// WorkerWithMaxWorkers set maximum number of workers for a worker pool
func WorkerWithMaxWorkers(n int) WorkOption {
	return func(wp *WorkerPool) {
		wp.maxWorkers = n
	}
}

// WorkerWithCapacity set capacity of task buffer for a worker pool.
// if n == 0, no buffer
// TODO: if n < 0, no limit (maybe persistent store task data)
// TODO: design a storage interface? to storage tasks in memory or disks
func WorkerWithCapacity(n int) WorkOption {
	return func(wp *WorkerPool) {
		wp.capacity = n
	}
}

// WorkWithIdle set idle timeout for a worker pool
func WorkerWithIdle(d time.Duration) WorkOption {
	return func(wp *WorkerPool) {
		wp.idleTimeout = d
	}
}

// WorkerWithStrategy set strategy for resizing the worker pool
func WorkerWithStrategy(s ResizingStrategy) WorkOption {
	return func(wp *WorkerPool) {
		wp.strategy = s
	}
}

// WorkerWithPanicHandler set panic handler for a worker pool
func WorkerWithPanicHandler(h func(interface{})) WorkOption {
	return func(wp *WorkerPool) {
		wp.panicHandler = h
	}
}

type WorkerPool struct {
	maxWorkers         int
	minWorkers         int
	capacity           int // capacity of task buffer
	idleTimeout        time.Duration
	strategy           ResizingStrategy
	panicHandler       func(interface{})
	runningWorkerCount int32 // if workerCount == idleWorkerCount => no tasks
	idleWorkerCount    int32
	// submitted = waiting + running + successful + failed
	submittedTaskCount  uint64
	waitingTaskCount    uint64
	successfulTaskCount uint64
	failedTaskCount     uint64
	tasks               chan TaskFunc
	shutdown            chan struct{}
	done                chan struct{}
	mu                  sync.Mutex
}

func New(options ...WorkOption) *WorkerPool {
	wp := &WorkerPool{
		idleTimeout:  defaultIdleTimeout,
		strategy:     Eager(),
		panicHandler: defaultPanicHandler,
		shutdown:     make(chan struct{}),
		done:         make(chan struct{}),
	}

	for _, opt := range options {
		opt(wp)
	}

	if wp.maxWorkers <= 0 {
		wp.maxWorkers = 1
	}
	if wp.minWorkers > wp.maxWorkers {
		wp.minWorkers = wp.maxWorkers
	} else if wp.minWorkers <= 0 {
		wp.minWorkers = 1
	}
	if wp.capacity < 0 {
		wp.capacity = 0
	}
	if wp.idleTimeout < 0 {
		wp.idleTimeout = defaultIdleTimeout
	}

	wp.tasks = make(chan TaskFunc, wp.capacity)

	go wp.purge()

	for i := 0; i < wp.minWorkers; i++ {
		// start worker directly, no need resizing
		go wp.work(nil)
	}
	return wp
}

// RunningWorkers returns the current number of running workers
func (wp *WorkerPool) RunningWorkers() int {
	return int(atomic.LoadInt32(&wp.runningWorkerCount))
}

// IdleWorkers returns the current number of idle workers
func (wp *WorkerPool) IdleWorkers() int {
	return int(atomic.LoadInt32(&wp.idleWorkerCount))
}

// MinWorkers returns the minimum number of worker goroutines
func (wp *WorkerPool) MinWorkers() int {
	return wp.minWorkers
}

// MaxWorkers returns the maximum number of worker goroutines
func (wp *WorkerPool) MaxWorkers() int {
	return wp.maxWorkers
}

// Capacity returns the current capacity of task buffer
func (wp *WorkerPool) Capacity() int {
	return wp.capacity
}

// Strategy returns the configured pool resizing strategy
func (wp *WorkerPool) Strategy() ResizingStrategy {
	return wp.strategy
}

// SubmittedTasks returns the total number of tasks submitted since the pool was
// created
func (wp *WorkerPool) SubmittedTasks() uint64 {
	return atomic.LoadUint64(&wp.submittedTaskCount)
}

// WaitingTasks returns the current number of submitted that are waiting to be
// executed
func (wp *WorkerPool) WaitingTasks() uint64 {
	return atomic.LoadUint64(&wp.waitingTaskCount)
}

// RunningTasks returns the current number of running tasks, maybe not accuracy
func (wp *WorkerPool) RunningTasks() uint64 {
	return wp.SubmittedTasks() - wp.WaitingTasks() - wp.CompletedTasks()
}

// SuccessfulTasks returns the total number of tasks that have successfully
// completed their exection since the pool was created
func (wp *WorkerPool) SuccessfulTasks() uint64 {
	return atomic.LoadUint64(&wp.successfulTaskCount)
}

// FailedTasks returns the total number of tasks that completed with error or
// panic since the pool was created
func (wp *WorkerPool) FailedTasks() uint64 {
	return atomic.LoadUint64(&wp.failedTaskCount)
}

// CompletedTasks returns the total number of tasks that have completed their
// exection either successfully or failed since the pool was created
func (wp *WorkerPool) CompletedTasks() uint64 {
	return wp.SuccessfulTasks() + wp.FailedTasks()
}

// Submit submit a task to the worker pool. It blocks until the task is
// dispatched to a worker.
func (wp *WorkerPool) Submit(ctx context.Context, task TaskFunc) {
	wp.submit(ctx, task, true)
}

// TrySubmit attempts to submit a task to the worker pool.
// It would not block if there's no idle worker.
// It returns true if it the task has been dispatched to a worker.
func (wp *WorkerPool) TrySubmit(ctx context.Context, task TaskFunc) bool {
	return wp.submit(ctx, task, false)
}

// SubmitAndWait submit a task to the worker pool and waits for complete.
func (wp *WorkerPool) SubmitAndWait(ctx context.Context, task TaskFunc) {
	done := make(chan struct{})
	submitted := wp.submit(ctx, func() error {
		defer close(done)
		return task()
	}, true)
	if submitted {
		<-done
	}
}

// TrySubmitAndWait submit a task to the worker pool and waits for complete.
// It would not block if there's no idle worker.
// It returns true if it the task has been dispatched to a worker.
func (wp *WorkerPool) TrySubmitAndWait(ctx context.Context, task TaskFunc) bool {
	done := make(chan struct{})
	submitted := wp.submit(ctx, func() error {
		defer close(done)
		return task()
	}, true)
	if submitted {
		<-done
	}
	return submitted
}

func (p *WorkerPool) submit(ctx context.Context, task TaskFunc, wait bool,
) (submitted bool) {
	if task == nil {
		return false
	}

	select {
	case <-p.shutdown:
		panic("pool: use of a closed worker pool")
	default:
	}

	defer func() {
		if submitted {
			atomic.AddUint64(&p.submittedTaskCount, 1)
			atomic.AddUint64(&p.waitingTaskCount, 1)
		}
	}()

	if p.IdleWorkers() == 0 && p.allowAddWorker() {
		go p.work(task)
		submitted = true
		return
	}

	// exist idle workers or cannot add new worker
	if wait {
		select {
		case <-p.shutdown:
		case p.tasks <- task:
			submitted = true
		}
		return
	}

	select {
	case <-ctx.Done():
		// abort by context
	case <-p.shutdown:
	case p.tasks <- task:
		submitted = true
	default:
	}
	return
}

func (p *WorkerPool) allowAddWorker() bool {
	runningWorkers := p.RunningWorkers()
	if !p.strategy.Resize(runningWorkers) || runningWorkers >= p.maxWorkers {
		return false
	}
	return true
}

// Close shutdown the pool and remove all waiting tasks.
// Panics if called for a closed worker pool.
func (wp *WorkerPool) Close() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	select {
	case <-wp.shutdown:
		panic("pool: close of a closed worker pool")
	default:
	}
	close(wp.shutdown)
}

// CloseAndWait causes this pool to stop accepting tasks, waiting for all the
// submitted tasks to complete.
func (wp *WorkerPool) CloseAndWait(ctx context.Context) error {
	wp.Close()
	// wait for no running tasks
	if atomic.LoadInt32(&wp.runningWorkerCount) != 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-wp.done:
			return nil
		}
	}
	return nil
}

// purge kill idle work as idleTimeout
func (wp *WorkerPool) purge() {
	idleTicker := time.NewTicker(wp.idleTimeout)
	defer func() {
		idleTicker.Stop()
		close(wp.tasks)
	}()

	for {
		select {
		case <-idleTicker.C:
			// runningWorkers - wp.minWorkers = canCutWorkers
			// idleWorkers - canCutWorkers = shouldCutWorkers
			cut := wp.IdleWorkers() - (wp.RunningWorkers() - wp.minWorkers)
			if cut > 0 {
				for i := 0; i < cut; i++ {
					wp.tasks <- nil
				}
			}
		case <-wp.shutdown:
			return
		}
	}
}

// work start a worker with firstTask
func (wp *WorkerPool) work(firstTask TaskFunc) {
	atomic.AddInt32(&wp.runningWorkerCount, 1)
	if firstTask != nil {
		wp.exec(firstTask)
	}
	atomic.AddInt32(&wp.idleWorkerCount, 1)

	for task := range wp.tasks {
		if task == nil {
			// received quit singal by purge
			break
		}
		atomic.AddInt32(&wp.idleWorkerCount, -1)
		wp.exec(task)
		atomic.AddInt32(&wp.idleWorkerCount, 1)
	}
	atomic.AddInt32(&wp.idleWorkerCount, -1)
	cnt := atomic.AddInt32(&wp.runningWorkerCount, -1)
	if cnt == 0 {
		// send done singal after shutdown if there's no running workers
		select {
		case <-wp.shutdown:
			close(wp.done)
		default:
		}
	}
}

func (wp *WorkerPool) exec(task TaskFunc) {
	defer func() {
		if p := recover(); p != nil {
			atomic.AddUint64(&wp.failedTaskCount, 1)
			wp.panicHandler(p)
		}
	}()
	atomic.AddUint64(&wp.waitingTaskCount, ^uint64(0))

	if err := task(); err != nil {
		atomic.AddUint64(&wp.failedTaskCount, 1)
	} else {
		atomic.AddUint64(&wp.successfulTaskCount, 1)
	}
}
