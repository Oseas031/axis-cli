package tmp

func MinMax(a, b int) (int, int) {
	if a < b {
		return a, b
	}
	return b, a
}
