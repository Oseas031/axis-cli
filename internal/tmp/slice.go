package tmp

// Reverse returns a new slice with elements in reverse order.
func Reverse(s []int) []int {
	result := make([]int, len(s))
	for i := 0; i < len(s); i++ {
		result[i] = s[len(s)-1-i]
	}
	return result
}

// Sum returns the sum of all elements.
func Sum(s []int) int {
	total := 0
	for _, v := range s {
		total += v
	}
	return total
}