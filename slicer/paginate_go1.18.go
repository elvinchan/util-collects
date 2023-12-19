//go:build go1.18
// +build go1.18

package slicer

func Paginate[S ~[]E, E any](items S, currentPage, maxResults int) S {
	offset := (currentPage - 1) * maxResults
	if offset > len(items) {
		offset = len(items)
	}
	end := offset + maxResults
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}
