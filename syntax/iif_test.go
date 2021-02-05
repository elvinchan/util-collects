package syntax

import (
	"testing"

	"github.com/elvinchan/util-collects/testkit"
)

func TestIIf(t *testing.T) {
	a, b := 2, 3
	max := IIf(a > b, a, b).(int)
	testkit.Assert(t, max == 3)
}
