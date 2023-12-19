//go:build !go1.18
// +build !go1.18

package slicer

import "reflect"

func Paginate(items interface{}, currentPage, maxResults int) interface{} {
	data := reflect.ValueOf(items)
	if data.Kind() != reflect.Slice {
		// no-op
		return items
	}

	start := (currentPage - 1) * maxResults
	if start > data.Len() {
		start = data.Len()
	}
	end := start + maxResults
	if end > data.Len() {
		end = data.Len()
	}
	res := reflect.MakeSlice(data.Type(), end-start, end-start)
	reflect.Copy(res, data.Slice(start, end))
	return res.Interface()
}
