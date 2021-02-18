package convert

import (
	"bytes"
	"strings"
	"testing"

	"github.com/elvinchan/util-collects/testkit"
)

func TestStrToBytes(t *testing.T) {
	x := "hello"
	y := []byte("hello")
	c := bytes.Compare(StrToBytes(x), y)
	testkit.Assert(t, c == 0)
}

func TestBytesToStr(t *testing.T) {
	x := "hello"
	y := []byte("hello")
	c := strings.Compare(x, BytesToStr(y))
	testkit.Assert(t, c == 0)
}
