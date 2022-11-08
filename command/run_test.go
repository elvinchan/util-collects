package command

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/elvinchan/util-collects/as"
	"github.com/elvinchan/util-collects/human"
)

func TestRunBytes(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		b, err := RunBytes("echo", RunWithArgs("hello"), RunWithTimeout(time.Millisecond*800))
		as.NoError(t, err)
		as.Equal(t, string(b), "hello\n")
	})

	t.Run("Timeout", func(t *testing.T) {
		b, err := RunBytes("sleep", RunWithArgs("2"), RunWithTimeout(time.Millisecond*800))
		as.Equal(t, err.Error(), context.DeadlineExceeded.Error())
		as.Equal(t, string(b), "")
	})

	t.Run("BeyondLimit", func(t *testing.T) {
		longTxt := "abcdefghijklmnopqrstuvwxyz"
		b, err := RunBytes("echo", RunWithArgs(longTxt), RunWithTimeout(time.Millisecond*800), RunWithSize(16))
		as.Equal(t, err.Error(), fmt.Sprintf("data beyond limit: %v", human.IBytes(16)))
		as.Equal(t, string(b), longTxt[:16])
	})

	t.Run("ErrToOutput", func(t *testing.T) {
		b, err := RunBytes("sh", RunWithArgs("-c", "echo stdout; echo 1>&2 stderr"))
		as.NoError(t, err)
		as.Equal(t, string(b), "stdout\n")

		b, err = RunBytes("sh", RunWithArgs("-c", "echo stdout; echo 1>&2 stderr"), RunWithErrToOutput())
		as.NoError(t, err)
		as.Equal(t, string(b), "stdout\nstderr\n")
	})
}
