package tmp

import "strconv"

// ParseInt safely parses a string to int, returns 0 on error.
func ParseInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0 // FIXED: returns int instead of string
	}
	return v
}
