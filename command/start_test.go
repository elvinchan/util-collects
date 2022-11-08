//go:build !windows
// +build !windows

package command

import (
	"bytes"
	"syscall"
	"testing"
	"time"

	"github.com/elvinchan/util-collects/as"
)

func TestStart(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		var buf bytes.Buffer
		p, err := Start("echo", StartWithArgs("hello"), StartWithStdout(&buf))
		as.NoError(t, err)
		pgid, err := syscall.Getpgid(p.Process.Pid)
		as.NoError(t, err)
		ppgid, err := syscall.Getpgid(syscall.Getpid())
		as.NoError(t, err)
		as.Equal(t, pgid, ppgid)
		time.Sleep(time.Millisecond * 500)
		as.Equal(t, buf.String(), "hello\n")
	})

	t.Run("Detach", func(t *testing.T) {
		var buf bytes.Buffer
		p, err := Start("echo", StartWithArgs("hello"), StartWithStdout(&buf), StartWithDetach())
		as.NoError(t, err)
		pgid, err := syscall.Getpgid(p.Process.Pid)
		as.NoError(t, err)
		ppgid, err := syscall.Getpgid(syscall.Getpid())
		as.NoError(t, err)
		as.NotEqual(t, pgid, ppgid)
		time.Sleep(time.Millisecond * 500)
		as.Equal(t, buf.String(), "hello\n")
	})
}
