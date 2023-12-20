package slicer

func UniqGroupBy[S ~[]E, E any](items S, iteratee func(current, include int) bool) S {
	idx := 0
	for i := range items {
		exist := false
		for j := 0; j < idx; j++ {
			if iteratee(i, j) {
				exist = true
				break
			}
		}
		if !exist {
			if idx != i {
				items[idx] = items[i]
			}
			idx++
		}
	}
	return items[:idx]
}
