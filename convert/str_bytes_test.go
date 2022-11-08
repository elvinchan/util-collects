package convert

import (
	"bytes"
	"strings"
	"testing"

	"github.com/elvinchan/util-collects/as"
)

func TestStrToBytes(t *testing.T) {
	x := "hello"
	y := []byte("hello")
	c := bytes.Compare(StrToBytes(x), y)
	as.Equal(t, c, 0)
}

func TestBytesToStr(t *testing.T) {
	x := "hello"
	y := []byte("hello")
	c := strings.Compare(x, BytesToStr(y))
	as.Equal(t, c, 0)
}
