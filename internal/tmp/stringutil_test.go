package tmp

import "testing"

func TestIsPalindrome(t *testing.T) {
	if !IsPalindrome("racecar") { t.Error("racecar should be palindrome") }
	if !IsPalindrome("aba") { t.Error("aba should be palindrome") }
	if IsPalindrome("hello") { t.Error("hello should not be palindrome") }
	if !IsPalindrome("") { t.Error("empty string should be palindrome") }
}

func TestCountVowels(t *testing.T) {
	if got := CountVowels("hello"); got != 2 { t.Errorf("got %d want 2", got) }
	if got := CountVowels("AEIOU"); got != 5 { t.Errorf("got %d want 5", got) }
	if got := CountVowels("xyz"); got != 0 { t.Errorf("got %d want 0", got) }
}
