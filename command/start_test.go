// +build !windows

package command

import (
	"bytes"
	"syscall"
	"testing"
	"time"

	"github.com/elvinchan/util-collects/testkit"
)

func TestStart(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		var buf bytes.Buffer
		p, err := Start("echo", StartWithArgs("hello"), StartWithStdout(&buf))
		testkit.Assert(t, err == nil)
		pgid, err := syscall.Getpgid(p.Process.Pid)
		testkit.Assert(t, err == nil)
		ppgid, err := syscall.Getpgid(syscall.Getpid())
		testkit.Assert(t, err == nil)
		testkit.Assert(t, pgid == ppgid)
		time.Sleep(time.Second)
		testkit.Assert(t, buf.String() == "hello\n")
	})

	t.Run("Detach", func(t *testing.T) {
		var buf bytes.Buffer
		p, err := Start("echo", StartWithArgs("hello"), StartWithStdout(&buf), StartWithDetach())
		testkit.Assert(t, err == nil)
		pgid, err := syscall.Getpgid(p.Process.Pid)
		testkit.Assert(t, err == nil)
		ppgid, err := syscall.Getpgid(syscall.Getpid())
		testkit.Assert(t, err == nil)
		testkit.Assert(t, pgid != ppgid)
		time.Sleep(time.Second)
		testkit.Assert(t, buf.String() == "hello\n")
	})
}
