package tmp

import "testing"

func TestContains(t *testing.T) {
	if !Contains([]int{1, 2, 3}, 2) {
		t.Error("expected true for [1,2,3] contains 2")
	}
	if Contains([]int{1, 2, 3}, 5) {
		t.Error("expected false for [1,2,3] contains 5")
	}
	if Contains(nil, 1) {
		t.Error("expected false for nil slice")
	}
}
