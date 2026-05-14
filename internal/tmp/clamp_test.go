package tmp

import "testing"

func TestClamp(t *testing.T) {
	tests := []struct{ value, min, max, want int }{
		{5, 0, 10, 5},
		{-3, 0, 10, 0},
		{15, 0, 10, 10},
		{0, 0, 10, 0},
		{10, 0, 10, 10},
	}
	for _, tt := range tests {
		got := Clamp(tt.value, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("Clamp(%d, %d, %d) = %d, want %d", tt.value, tt.min, tt.max, got, tt.want)
		}
	}
}
