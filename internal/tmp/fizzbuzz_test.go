package tmp

import "testing"

func TestFizzBuzz(t *testing.T) {
	cases := []struct{ n int; want string }{
		{1, "1"}, {3, "Fizz"}, {5, "Buzz"}, {15, "FizzBuzz"},
		{7, "7"}, {9, "Fizz"}, {10, "Buzz"}, {30, "FizzBuzz"},
	}
	for _, c := range cases {
		got := FizzBuzz(c.n)
		if got != c.want {
			t.Errorf("FizzBuzz(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}
