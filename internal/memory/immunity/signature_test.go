package immunity

import (
	"strings"
	"testing"
)

func TestSignature_Hash_StableAcrossKeyOrder(t *testing.T) {
	a := BuildSignature("build.binary",
		map[string]string{"alpha": "1", "beta": "2", "gamma": "3"},
		[]string{"go", "git"},
		"failure.provider.timeout",
	)
	b := BuildSignature("build.binary",
		map[string]string{"gamma": "3", "alpha": "1", "beta": "2"},
		[]string{"git", "go"},
		"failure.provider.timeout",
	)
	if a.Hash() != b.Hash() {
		t.Errorf("hash should be stable across key/slice order:\n  a=%s\n  b=%s", a.Hash(), b.Hash())
	}
}

func TestSignature_Hash_Length(t *testing.T) {
	s := BuildSignature("x", nil, nil, "failure.runtime.panic")
	h := s.Hash()
	if len(h) != 32 {
		t.Errorf("hash length = %d, want 32", len(h))
	}
	// hex chars only
	for _, r := range h {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			t.Errorf("hash contains non-hex char %q in %q", r, h)
			break
		}
	}
}

func TestSignature_Hash_DifferentInputsDiffer(t *testing.T) {
	tests := []struct {
		name string
		a, b Signature
	}{
		{
			"different intent",
			BuildSignature("a", nil, nil, "failure.runtime.panic"),
			BuildSignature("b", nil, nil, "failure.runtime.panic"),
		},
		{
			"different error class",
			BuildSignature("a", nil, nil, "failure.runtime.panic"),
			BuildSignature("a", nil, nil, "failure.provider.timeout"),
		},
		{
			"different args",
			BuildSignature("a", map[string]string{"k": "1"}, nil, "failure.runtime.panic"),
			BuildSignature("a", map[string]string{"k": "2"}, nil, "failure.runtime.panic"),
		},
		{
			"different tools",
			BuildSignature("a", nil, []string{"go"}, "failure.runtime.panic"),
			BuildSignature("a", nil, []string{"git"}, "failure.runtime.panic"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.a.Hash() == tt.b.Hash() {
				t.Errorf("expected distinct hashes, both = %s", tt.a.Hash())
			}
		})
	}
}

func TestSignature_Hash_EmptyValuesDropped(t *testing.T) {
	with := BuildSignature("x",
		map[string]string{"k": "v", "empty": ""},
		[]string{"go", ""},
		"failure.runtime.panic",
	)
	without := BuildSignature("x",
		map[string]string{"k": "v"},
		[]string{"go"},
		"failure.runtime.panic",
	)
	if with.Hash() != without.Hash() {
		t.Errorf("empty values should be dropped from hash:\n  with=%s\n  without=%s", with.Hash(), without.Hash())
	}
}

func TestSignature_Hash_ToolDeduplication(t *testing.T) {
	dup := BuildSignature("x", nil, []string{"go", "git", "go", "git"}, "failure.runtime.panic")
	uniq := BuildSignature("x", nil, []string{"go", "git"}, "failure.runtime.panic")
	if dup.Hash() != uniq.Hash() {
		t.Errorf("tool dedup expected:\n  dup=%s\n  uniq=%s", dup.Hash(), uniq.Hash())
	}
}

func TestNormalizeArgs_DropsSensitiveKeys(t *testing.T) {
	raw := map[string]any{
		"intent":         "build",
		"api_key":        "sk-secret-123",
		"openai_api_key": "sk-secret-456",
		"Authorization":  "Bearer xyz",
		"my_password":    "hunter2",
		"user_id":        "alice",
	}
	out := NormalizeArgs(raw)
	for _, banned := range []string{"api_key", "openai_api_key", "Authorization", "my_password"} {
		if _, found := out[banned]; found {
			t.Errorf("sensitive key %q should have been dropped, got: %v", banned, out)
		}
	}
	if out["intent"] != "build" || out["user_id"] != "alice" {
		t.Errorf("non-sensitive keys lost: %v", out)
	}
}

func TestNormalizeArgs_HashUnaffectedBySensitiveKey(t *testing.T) {
	withSecret := NormalizeArgs(map[string]any{
		"intent":  "build",
		"api_key": "sk-leak",
	})
	without := NormalizeArgs(map[string]any{
		"intent": "build",
	})
	a := BuildSignature("build", withSecret, nil, "failure.provider.timeout")
	b := BuildSignature("build", without, nil, "failure.provider.timeout")
	if a.Hash() != b.Hash() {
		t.Errorf("sensitive key should not affect hash:\n  with=%s\n  without=%s", a.Hash(), b.Hash())
	}
}

func TestNormalizeArgs_ValueTypes(t *testing.T) {
	out := NormalizeArgs(map[string]any{
		"s":     "hello",
		"i":     42,
		"f":     3.14,
		"b":     true,
		"nil":   nil,
		"empty": "",
		"slice": []string{"a", "b"},
	})
	if out["s"] != "hello" {
		t.Errorf("string mishandled: %q", out["s"])
	}
	if out["i"] != "42" {
		t.Errorf("int mishandled: %q", out["i"])
	}
	if out["b"] != "true" {
		t.Errorf("bool mishandled: %q", out["b"])
	}
	if _, ok := out["nil"]; ok {
		t.Errorf("nil should be dropped, got: %v", out)
	}
	if _, ok := out["empty"]; ok {
		t.Errorf("empty string should be dropped, got: %v", out)
	}
	if !strings.Contains(out["slice"], "a") || !strings.Contains(out["slice"], "b") {
		t.Errorf("slice should serialise to JSON-ish, got: %q", out["slice"])
	}
}

func TestNormalizeArgs_EmptyInput(t *testing.T) {
	got := NormalizeArgs(nil)
	if got == nil {
		t.Error("expected non-nil empty map")
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}
