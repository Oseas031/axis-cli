package tmp

// Range returns integers from start to end (inclusive).
func Range(start, end int) []int {
	result := make([]int, 0)
	for i := start; i <= end; i++ {
		result = append(result, i)
	}
	return result
}
