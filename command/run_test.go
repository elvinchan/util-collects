package command

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/elvinchan/util-collects/human"
	"github.com/elvinchan/util-collects/testkit"
)

func TestRunBytes(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		b, err := RunBytes("echo", RunWithArgs("hello"), RunWithTimeout(time.Millisecond*800))
		testkit.Assert(t, err == nil)
		testkit.Assert(t, string(b) == "hello\n")
	})

	t.Run("Timeout", func(t *testing.T) {
		b, err := RunBytes("sleep", RunWithArgs("2"), RunWithTimeout(time.Millisecond*800))
		testkit.Assert(t, err.Error() == context.DeadlineExceeded.Error())
		testkit.Assert(t, string(b) == "")
	})

	t.Run("BeyondLimit", func(t *testing.T) {
		longTxt := "abcdefghijklmnopqrstuvwxyz"
		b, err := RunBytes("echo", RunWithArgs(longTxt), RunWithTimeout(time.Millisecond*800), RunWithSize(16))
		testkit.Assert(t, err.Error() == fmt.Sprintf("data beyond limit: %v", human.IBytes(16)))
		testkit.Assert(t, string(b) == longTxt[:16])
	})

	t.Run("ErrToOutput", func(t *testing.T) {
		b, err := RunBytes("sh", RunWithArgs("-c", "echo stdout; echo 1>&2 stderr"))
		testkit.Assert(t, err == nil)
		testkit.Assert(t, string(b) == "stdout\n")

		b, err = RunBytes("sh", RunWithArgs("-c", "echo stdout; echo 1>&2 stderr"), RunWithErrToOutput())
		testkit.Assert(t, err == nil)
		testkit.Assert(t, string(b) == "stdout\nstderr\n")
	})
}
