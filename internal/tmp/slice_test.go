package tmp

import (
	"testing"
	"reflect"
)

func TestReverse(t *testing.T) {
	got := Reverse([]int{1, 2, 3, 4})
	want := []int{4, 3, 2, 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Reverse([1,2,3,4]) = %v, want %v", got, want)
	}
}

func TestSum(t *testing.T) {
	got := Sum([]int{1, 2, 3})
	if got != 6 {
		t.Errorf("Sum([1,2,3]) = %d, want 6", got)
	}
}
