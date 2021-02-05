package command

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"time"

	"github.com/elvinchan/util-collects/human"
)

type Commander struct {
	args    []string
	envs    []string
	timeout time.Duration
	size    uint64
}

type CommandOption func(*Commander)

func WithArgs(args ...string) CommandOption {
	return func(c *Commander) {
		c.args = append(c.args, args...)
	}
}

func WithEnv(envs ...string) CommandOption {
	return func(c *Commander) {
		c.envs = append(c.envs, envs...)
	}
}

func WithTimeout(timeout time.Duration) CommandOption {
	return func(c *Commander) {
		c.timeout = timeout
	}
}

func WithSize(size uint64) CommandOption {
	return func(c *Commander) {
		c.size = size
	}
}

// refer: https://medium.com/@vCabbage/go-timeout-commands-with-os-exec-commandcontext-ba0c861ed738
func Run(name string, opts ...CommandOption) ([]byte, error) {
	var c *Commander
	for _, opt := range opts {
		opt(c)
	}
	var cmd *exec.Cmd
	ctx := context.Background()
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}
	cmd = exec.CommandContext(ctx, name, c.args...)
	cmd.Env = c.envs
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	data, err := readLimit(stdout, c.size)
	if err != nil {
		return data, err
	}

	errc := make(chan error, 1)
	go func() {
		errc <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		return data, ctx.Err()
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
