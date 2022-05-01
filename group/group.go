package group

import (
	"context"
	"sync"

	"github.com/elvinchan/util-collects/group/semaphore"
)

type group struct {
	ctx    context.Context
	cancel func()
	sem    *semaphore.Dynamic

	wg sync.WaitGroup

	errOnce sync.Once
	err     error
}

// New create a group with limited concurrent number of workers.
func New(ctx context.Context, limit int64) *group {
	ctx, cancel := context.WithCancel(ctx)
	return &group{
		ctx:    ctx,
		cancel: cancel,
		sem:    semaphore.NewDynamic(limit),
	}
}

// Context returns context passed by New().
func (g *group) Context() context.Context {
	return g.ctx
}

// SetLimit changes limit of concurrent num for the group. if n <= 0, no limit.
func (g *group) SetLimit(n int64) {
	g.sem.SetSize(n)
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (g *group) Go(f func() error) {
	g.wg.Add(1)
	if err := g.sem.Acquire(g.ctx); err != nil {
		g.wg.Done()
		return
	}

	finish := func() {
		g.sem.Release()
		g.wg.Done()
	}
	select {
	case <-g.ctx.Done():
		finish() // prevent go into the worker
		return
	default:
	}

	go func() {
		defer finish()
		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}
