package tmp

import "testing"

func TestParseInt(t *testing.T) {
	if v := ParseInt("42"); v != 42 { t.Errorf("got %d want 42", v) }
	if v := ParseInt("abc"); v != 0 { t.Errorf("got %d want 0", v) }
	if v := ParseInt("-7"); v != -7 { t.Errorf("got %d want -7", v) }
}
