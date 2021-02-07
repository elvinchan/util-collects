package command

import (
	"context"
	"io"
	"os/exec"
	"syscall"
)

type Starter struct {
	ctx        context.Context
	args, envs []string
	out, err   io.Writer
	detach     bool
}

type StartOption func(*Starter)

func StartWithContext(ctx context.Context) StartOption {
	return func(s *Starter) {
		s.ctx = ctx
	}
}

func StartWithArgs(args ...string) StartOption {
	return func(s *Starter) {
		s.args = append(s.args, args...)
	}
}

func StartWithEnv(envs ...string) StartOption {
	return func(s *Starter) {
		s.envs = append(s.envs, envs...)
	}
}

func StartWithStdout(out io.Writer) StartOption {
	return func(s *Starter) {
		s.out = out
	}
}

func StartWithStderr(err io.Writer) StartOption {
	return func(s *Starter) {
		s.err = err
	}
}

func StartWithDetach() StartOption {
	return func(s *Starter) {
		s.detach = true
	}
}

func Start(name string, opts ...StartOption) (*exec.Cmd, error) {
	s := &Starter{
		ctx: context.Background(),
	}
	for _, opt := range opts {
		opt(s)
	}
	cmd := exec.CommandContext(s.ctx, name, s.args...)
	cmd.Env = s.envs
	cmd.Stdout = s.out
	cmd.Stderr = s.err
	if s.detach {
		if cmd.SysProcAttr == nil {
			cmd.SysProcAttr = &syscall.SysProcAttr{}
		}
		detachAttr(cmd.SysProcAttr)
	}
	return cmd, cmd.Start()
}
