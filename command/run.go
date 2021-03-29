package command

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/elvinchan/util-collects/human"
)

type Runner struct {
	ctx         context.Context
	args, envs  []string
	timeout     time.Duration
	size        uint64
	errToOutput bool
}

type RunOption func(*Runner)

func RunWithContext(ctx context.Context) RunOption {
	return func(r *Runner) {
		r.ctx = ctx
	}
}

func RunWithArgs(args ...string) RunOption {
	return func(r *Runner) {
		r.args = append(r.args, args...)
	}
}

func RunWithEnv(envs ...string) RunOption {
	return func(r *Runner) {
		r.envs = append(r.envs, envs...)
	}
}

func RunWithTimeout(timeout time.Duration) RunOption {
	return func(r *Runner) {
		r.timeout = timeout
	}
}

func RunWithSize(size uint64) RunOption {
	return func(r *Runner) {
		r.size = size
	}
}

func RunWithErrToOutput() RunOption {
	return func(r *Runner) {
		r.errToOutput = true
	}
}

// RunBytes runs command and receives byte slice from stdout until
// process exit or any error when reading from stdout.
//
// refer: https://medium.com/@vCabbage/go-timeout-commands-with-os-exec-commandcontext-ba0c861ed738
func RunBytes(name string, opts ...RunOption) ([]byte, error) {
	r := &Runner{
		ctx: context.Background(),
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.timeout > 0 {
		var cancel context.CancelFunc
		r.ctx, cancel = context.WithTimeout(r.ctx, r.timeout)
		defer cancel()
	}
	cmd := exec.CommandContext(r.ctx, name, r.args...)
	cmd.Env = r.envs
	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var stdout io.Reader
	if r.errToOutput {
		errReader, err := cmd.StderrPipe()
		if err != nil {
			return nil, err
		}
		stdout = io.MultiReader(outReader, errReader)
	} else {
		stdout = outReader
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var hadWait uint32
	defer func() {
		// prevent two approach:
		// panic when reading from pipe
		// read from pipe with error
		if atomic.CompareAndSwapUint32(&hadWait, 0, 1) {
			_ = cmd.Wait()
		}
	}()

	// Not waiting for context done in parallel because we need this approach:
	// context done -> process been killed -> read all from pipe
	// -> return both the data has been read and ctx.Err()
	data, err := readLimit(stdout, r.size)
	if err != nil {
		return data, err
	}

	errc := make(chan error, 1)
	go func() {
		if atomic.CompareAndSwapUint32(&hadWait, 0, 1) {
			errc <- cmd.Wait()
		}
	}()

	select {
	case <-r.ctx.Done():
		return data, r.ctx.Err()
	case err := <-errc:
		return data, err
	}
}

func readLimit(r io.Reader, size uint64) ([]byte, error) {
	if size == 0 {
		// read all
		return ioutil.ReadAll(r)
	}
	var buf bytes.Buffer
	buf.Grow(1024) // default 1KB
	n, err := buf.ReadFrom(&io.LimitedReader{R: r, N: int64(size + 1)})
	if n > int64(size) {
		return buf.Bytes()[:size], fmt.Errorf("data beyond limit: %v", human.IBytes(size))
	}
	return buf.Bytes(), err
}
