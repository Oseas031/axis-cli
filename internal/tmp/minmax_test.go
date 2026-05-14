package tmp

import "testing"

func TestMinMax(t *testing.T) {
	min, max := MinMax(3, 7)
	if min != 3 || max != 7 {
		t.Errorf("MinMax(3,7) = %d,%d want 3,7", min, max)
	}
	min, max = MinMax(9, 2)
	if min != 2 || max != 9 {
		t.Errorf("MinMax(9,2) = %d,%d want 2,9", min, max)
	}
}
