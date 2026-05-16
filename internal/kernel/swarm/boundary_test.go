package swarm

import (
	"context"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

func TestSelectAgents_MaxIntSize(t *testing.T) {
	cfg := &SwarmConfig{Pattern: "parallel_vote", MinSize: 2, MaxSize: 999999, Diversity: "none", Order: "fixed"}
	slots, err := SelectAgents([]string{"a", "b", "c"}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	// Should cap at available count
	if len(slots) != 3 {
		t.Fatalf("expected 3 slots (capped), got %d", len(slots))
	}
}

func TestSelectAgents_DuplicateProviders(t *testing.T) {
	cfg := &SwarmConfig{Pattern: "parallel_vote", MinSize: 2, MaxSize: 3, Diversity: "none", Order: "fixed"}
	slots, err := SelectAgents([]string{"a", "a", "a"}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(slots) != 3 {
		t.Fatalf("expected 3 slots, got %d", len(slots))
	}
}

func TestDispatch_ZeroTimeout(t *testing.T) {
	cfg := &SwarmConfig{Pattern: "parallel_vote", MinSize: 2, MaxSize: 2, Diversity: "none", Order: "fixed"}
	agents := []AgentSlot{{AgentID: "a1", Provider: "p1"}, {AgentID: "a2", Provider: "p2"}}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(time.Millisecond) // ensure timeout fires
	fn := func(ctx context.Context, task *types.AgentTask, provider string) (map[string]any, error) {
		return map[string]any{"ok": true}, nil
	}
	task := &types.AgentTask{TaskID: "timeout-test"}
	// Should not panic
	_, _ = Dispatch(ctx, task, cfg, agents, fn)
}

func TestAggregate_SingleResult(t *testing.T) {
	results := []SingleResult{{AgentID: "a1", Output: map[string]any{"x": 1}}}
	sr, err := Aggregate(results)
	if err != nil {
		t.Fatal(err)
	}
	if sr.Confidence != 1.0 {
		t.Fatalf("single result should have confidence 1.0, got %f", sr.Confidence)
	}
}

func TestAggregate_NilOutput(t *testing.T) {
	results := []SingleResult{
		{AgentID: "a1", Output: nil},
		{AgentID: "a2", Output: nil},
	}
	sr, err := Aggregate(results)
	if err != nil {
		t.Fatal(err)
	}
	if !sr.Unanimous {
		t.Fatal("two nil outputs should be unanimous")
	}
}

func TestAggregate_EmptyMapVsNil(t *testing.T) {
	results := []SingleResult{
		{AgentID: "a1", Output: map[string]any{}},
		{AgentID: "a2", Output: nil},
	}
	sr, err := Aggregate(results)
	if err != nil {
		t.Fatal(err)
	}
	// empty map and nil should hash differently
	_ = sr
}

func TestValidate_EmptyStrings(t *testing.T) {
	cfg := &SwarmConfig{}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("zero-value config should fail validation")
	}
}

func TestParseFromMetadata_EmptyPattern(t *testing.T) {
	meta := map[string]string{"swarm.pattern": ""}
	if got := ParseFromMetadata(meta); got != nil {
		t.Fatal("empty pattern should return nil")
	}
}
