package slicer

// Paginate returns a page of items. If currentPage <= 0, reguard it as 1, if
// currentPage > last page, returns empty set. If maxResults <= -1, reguard it
// as no max results limit.
func Paginate[S ~[]E, E any](items S, currentPage, maxResults int) S {
	offset := (currentPage - 1) * maxResults
	return Subset(items, maxResults, offset)
}

// Subset returns a subset of items like SQL LIMIT/OFFSET. If offset < 0,
// reguard it as 0, if offset > last index, returns empty set. If limit <= -1,
// reguard it as no limit.
func Subset[S ~[]E, E any](items S, limit, offset int) S {
	if limit <= -1 {
		return items
	}
	if offset > len(items) {
		offset = len(items)
	} else if offset < 0 {
		offset = 0
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}
