package tmp

// Clamp restricts value to be within [min, max].
// If value < min, return min. If value > max, return max. Otherwise return value.
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}