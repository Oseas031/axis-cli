package guarantee

import (
	"errors"
	"testing"
)

func TestEmptyRegistryVerify(t *testing.T) {
	r := NewRegistry()
	if v := r.Verify(); len(v) != 0 {
		t.Fatalf("expected 0 violations, got %d", len(v))
	}
}

func TestRegisterAndPass(t *testing.T) {
	r := NewRegistry()
	r.Register(Guarantee{ID: "g1", Level: LevelHard, Check: func() error { return nil }})
	if v := r.Verify(); len(v) != 0 {
		t.Fatalf("expected 0 violations, got %d", len(v))
	}
}

func TestRegisterAndFail(t *testing.T) {
	r := NewRegistry()
	r.Register(Guarantee{ID: "g1", Level: LevelHard, Check: func() error { return errors.New("broken") }})
	v := r.Verify()
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].GuaranteeID != "g1" || v[0].Error.Error() != "broken" {
		t.Fatalf("unexpected violation: %+v", v[0])
	}
}

func TestHardVsSoftLevel(t *testing.T) {
	r := NewRegistry()
	r.Register(Guarantee{ID: "hard", Level: LevelHard, Check: func() error { return errors.New("h") }})
	r.Register(Guarantee{ID: "soft", Level: LevelSoft, Check: func() error { return errors.New("s") }})
	v := r.Verify()
	if len(v) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(v))
	}
	if v[0].Level != LevelHard || v[1].Level != LevelSoft {
		t.Fatalf("unexpected levels: %d, %d", v[0].Level, v[1].Level)
	}
}
