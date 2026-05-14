package tmp

import "testing"

func TestWordCount(t *testing.T) {
	got := WordCount("hello world hello")
	if got["hello"] != 2 { t.Errorf("hello count = %d, want 2", got["hello"]) }
	if got["world"] != 1 { t.Errorf("world count = %d, want 1", got["world"]) }
}
