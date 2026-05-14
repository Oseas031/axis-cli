package tmp

// Contains checks if slice contains the target value.
func Contains(slice []int, target int) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}