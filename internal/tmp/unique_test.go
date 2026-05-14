package tmp

import (
	"testing"
	"reflect"
)

func TestUnique(t *testing.T) {
	got := Unique([]int{1, 2, 2, 3, 1, 4, 3})
	want := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Unique([1,2,2,3,1,4,3]) = %v, want %v", got, want)
	}
	if got := Unique(nil); got != nil {
		t.Errorf("Unique(nil) = %v, want nil", got)
	}
}
