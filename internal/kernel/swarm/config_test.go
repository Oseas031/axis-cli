package swarm

import "testing"

func TestParseFromMetadata_Nil(t *testing.T) {
	if got := ParseFromMetadata(nil); got != nil {
		t.Fatal("expected nil for nil metadata")
	}
}

func TestParseFromMetadata_NoSwarmKeys(t *testing.T) {
	meta := map[string]string{"foo": "bar"}
	if got := ParseFromMetadata(meta); got != nil {
		t.Fatal("expected nil when no swarm.* keys")
	}
}

func TestParseFromMetadata_Defaults(t *testing.T) {
	meta := map[string]string{"swarm.pattern": "parallel_vote"}
	cfg := ParseFromMetadata(meta)
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.MinSize != 2 || cfg.MaxSize != 3 {
		t.Fatalf("bad defaults: min=%d max=%d", cfg.MinSize, cfg.MaxSize)
	}
	if cfg.Diversity != "heterogeneous" {
		t.Fatalf("bad diversity default: %s", cfg.Diversity)
	}
	if cfg.Order != "shuffled" {
		t.Fatalf("bad order default: %s", cfg.Order)
	}
}

func TestParseFromMetadata_CustomValues(t *testing.T) {
	meta := map[string]string{
		"swarm.pattern":   "parallel_vote",
		"swarm.min_size":  "3",
		"swarm.max_size":  "5",
		"swarm.diversity": "none",
		"swarm.order":     "fixed",
	}
	cfg := ParseFromMetadata(meta)
	if cfg.MinSize != 3 || cfg.MaxSize != 5 {
		t.Fatalf("custom sizes not parsed: min=%d max=%d", cfg.MinSize, cfg.MaxSize)
	}
	if cfg.Diversity != "none" || cfg.Order != "fixed" {
		t.Fatalf("custom values not parsed")
	}
}

func TestValidate_Valid(t *testing.T) {
	cfg := &SwarmConfig{Pattern: "parallel_vote", MinSize: 2, MaxSize: 3, Diversity: "heterogeneous", Order: "shuffled"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_BadPattern(t *testing.T) {
	cfg := &SwarmConfig{Pattern: "unknown", MinSize: 2, MaxSize: 3, Diversity: "none", Order: "fixed"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for bad pattern")
	}
}

func TestValidate_MinSizeTooSmall(t *testing.T) {
	cfg := &SwarmConfig{Pattern: "parallel_vote", MinSize: 1, MaxSize: 3, Diversity: "none", Order: "fixed"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for min_size < 2")
	}
}

func TestValidate_MaxLessThanMin(t *testing.T) {
	cfg := &SwarmConfig{Pattern: "parallel_vote", MinSize: 4, MaxSize: 3, Diversity: "none", Order: "fixed"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for max < min")
	}
}
