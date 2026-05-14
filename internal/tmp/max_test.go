package tmp

import "testing"

func TestMax(t *testing.T) {
	tests := []struct{ a, b, want int }{
		{3, 5, 5},
		{7, 2, 7},
		{4, 4, 4},
	}
	for _, tt := range tests {
		got := Max(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
