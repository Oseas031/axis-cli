package tmp

// Unique returns a new slice with duplicates removed, preserving order.
func Unique(s []int) []int {
	if s == nil {
		return nil
	}
	seen := make(map[int]bool)
	result := make([]int, 0, len(s))
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}