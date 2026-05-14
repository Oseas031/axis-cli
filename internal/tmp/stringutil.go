package tmp

// IsPalindrome checks if a string reads the same forwards and backwards.
func IsPalindrome(s string) bool {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		if runes[i] != runes[j] {
			return false
		}
	}
	return true
}

// CountVowels returns the number of vowels (aeiouAEIOU) in a string.
func CountVowels(s string) int {
	vowels := map[rune]bool{'a': true, 'e': true, 'i': true, 'o': true, 'u': true, 'A': true, 'E': true, 'I': true, 'O': true, 'U': true}
	count := 0
	for _, r := range s {
		if vowels[r] {
			count++
		}
	}
	return count
}
