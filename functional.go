package flexwriter

import "cmp"

// transpose transforms a list of columns (each possibly containing multiple
// rows) into a list of rows.
// The input may be sparse (i.e. some columns may have less rows than others),
// transpose will pad the shorter columns with zero values so that the output
// is rectangular.
func transpose[T any](data [][]T) [][]T {
	var maxLines int
	for _, col := range data {
		if len(col) > maxLines {
			maxLines = len(col)
		}
	}

	out := make([][]T, maxLines)
	for i := range out {
		out[i] = make([]T, len(data))
	}

	for i, col := range data {
		for j, line := range col {
			out[j][i] = line
		}
	}
	return out
}

// transform is a map() function, but it can't be called "map".
func transform[F any, T any](s []F, f func(F) T) []T {
	out := make([]T, len(s))
	for i, v := range s {
		out[i] = f(v)
	}
	return out
}

func max[T cmp.Ordered](s []T) T {
	if len(s) == 0 {
		var zero T
		return zero
	}

	max := s[0]
	for _, v := range s[1:] {
		if v > max {
			max = v
		}
	}
	return max
}
