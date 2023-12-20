package slicer

import (
	"testing"

	"github.com/elvinchan/util-collects/as"
)

func TestUniqGroupBy(t *testing.T) {
	type T struct {
		Key string
		Cnt int
	}

	var cases = []struct {
		Name   string
		Input  []T
		Output []T
	}{
		{
			"Normal",
			[]T{{"x", 1}, {"y", 2}, {"x", 3}},
			[]T{{"x", 4}, {"y", 2}},
		},
		{
			"NoRepeat",
			[]T{{"x", 1}, {"y", 2}, {"z", 3}},
			[]T{{"x", 1}, {"y", 2}, {"z", 3}},
		},
		{
			"More",
			[]T{{"x", 1}, {"y", 2}, {"y", 2}, {"z", 3}, {"x", 4}, {"z", 1}},
			[]T{{"x", 5}, {"y", 4}, {"z", 4}},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			output := UniqGroupBy(c.Input, func(current, include int) bool {
				if c.Input[current].Key == c.Input[include].Key {
					c.Input[include].Cnt += c.Input[current].Cnt
					return true
				}
				return false
			})
			as.Equal(t, output, c.Output)
		})
	}
}
