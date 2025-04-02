package syntax

import (
	"testing"

	"github.com/elvinchan/util-collects/as"
)

func TestIIf(t *testing.T) {
	a, b := 2, 3
	max := IIf(a > b, a, b)
	as.Equal(t, max, 3)
}
