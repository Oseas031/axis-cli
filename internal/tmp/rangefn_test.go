package tmp

import (
	"testing"
	"reflect"
)

func TestRange(t *testing.T) {
	got := Range(1, 5)
	want := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Range(1,5) = %v, want %v", got, want)
	}
}
