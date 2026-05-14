package tmp

import "testing"

func TestChainFormat(t *testing.T) {
	result := ChainFormat("  hello world  ", TrimFormatter{}, UpperFormatter{})
	if result != "HELLO WORLD" {
		t.Errorf("got %q, want %q", result, "HELLO WORLD")
	}
}

func TestTrimFormatter(t *testing.T) {
	f := TrimFormatter{}
	if got := f.Format("  hi  "); got != "hi" {
		t.Errorf("got %q, want %q", got, "hi")
	}
}
