//go:build go1.18
// +build go1.18

package slicer

import (
	"testing"

	"github.com/elvinchan/util-collects/as"
)

func TestPaginate(t *testing.T) {
	t.Run("NilSlice", func(t *testing.T) {
		v := Paginate([]string{}, 1, 2)
		as.Equal(t, v, []string{})
	})

	t.Run("Normal", func(t *testing.T) {
		v := Paginate([]int{1, 2, 3}, 1, 2)
		as.Equal(t, v, []int{1, 2})
	})

	t.Run("NormalStruct", func(t *testing.T) {
		type T struct {
			Id   int
			Name string
		}
		v := Paginate([]T{
			{Id: 1, Name: "x"},
			{Id: 2, Name: "y"},
			{Id: 3, Name: "z"},
		}, 2, 2)
		as.Equal(t, v, []T{
			{Id: 3, Name: "z"},
		})
	})

	t.Run("OutOfRange", func(t *testing.T) {
		v := Paginate([]int{1, 2, 3}, 5, 1)
		as.Equal(t, v, []int{})
	})
}
