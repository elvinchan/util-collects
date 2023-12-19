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

	t.Run("NoMaxResultsLimit", func(t *testing.T) {
		v := Paginate([]int{1, 2, 3}, 1, -1)
		as.Equal(t, v, []int{1, 2, 3})
	})

	t.Run("OutOfRange", func(t *testing.T) {
		v := Paginate([]int{1, 2, 3}, 5, 1)
		as.Equal(t, v, []int{})
	})

	t.Run("InvalidPage", func(t *testing.T) {
		v := Paginate([]int{1, 2, 3}, -1, 1)
		as.Equal(t, v, []int{1})
	})

	t.Run("InvalidMaxResults", func(t *testing.T) {
		v := Paginate([]int{1, 2, 3}, -1, 1)
		as.Equal(t, v, []int{1})
	})

	t.Run("InvalidAll", func(t *testing.T) {
		v := Paginate([]int{1, 2, 3}, 0, 0)
		as.Equal(t, v, []int{})
	})
}

func TestSubset(t *testing.T) {
	t.Run("NilSlice", func(t *testing.T) {
		v := Subset([]string{}, 1, 0)
		as.Equal(t, v, []string{})
	})

	t.Run("Normal", func(t *testing.T) {
		v := Subset([]int{1, 2, 3}, 2, 1)
		as.Equal(t, v, []int{2, 3})
	})

	t.Run("NormalStruct", func(t *testing.T) {
		type T struct {
			Id   int
			Name string
		}
		v := Subset([]T{
			{Id: 1, Name: "x"},
			{Id: 2, Name: "y"},
			{Id: 3, Name: "z"},
		}, 2, 2)
		as.Equal(t, v, []T{
			{Id: 3, Name: "z"},
		})
	})

	t.Run("OutOfRange", func(t *testing.T) {
		v := Subset([]int{1, 2, 3}, 1, 5)
		as.Equal(t, v, []int{})
	})

	t.Run("NoLimit", func(t *testing.T) {
		v := Subset([]int{1, 2, 3}, -1, 0)
		as.Equal(t, v, []int{1, 2, 3})
	})

	t.Run("InvalidOffset", func(t *testing.T) {
		v := Subset([]int{1, 2, 3}, 1, -1)
		as.Equal(t, v, []int{1})
	})

	t.Run("InvalidLimit", func(t *testing.T) {
		v := Subset([]int{1, 2, 3}, 0, 1)
		as.Equal(t, v, []int{})
	})

	t.Run("InvalidAll", func(t *testing.T) {
		v := Paginate([]int{1, 2, 3}, 0, 0)
		as.Equal(t, v, []int{})
	})
}
