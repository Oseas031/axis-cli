package tmp

// WordCount counts occurrences of each word in a string.
func WordCount(s string) map[string]int {
	counts := make(map[string]int)
	for _, word := range splitWords(s) {
		counts[word]++
	}
	return counts
}

func splitWords(s string) []string {
	var words []string
	current := ""
	for _, ch := range s {
		if ch == ' ' || ch == '\t' || ch == '\n' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}
