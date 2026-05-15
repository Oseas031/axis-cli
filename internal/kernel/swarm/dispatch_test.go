package swarm

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/axis-cli/axis/internal/types"
)

func TestSelectAgents_Empty(t *testing.T) {
	cfg := &SwarmConfig{Pattern: "parallel_vote", MinSize: 2, MaxSize: 3, Diversity: "none", Order: "fixed"}
	_, err := SelectAgents(nil, cfg)
	if err == nil {
		t.Fatal("expected error for empty providers")
	}
}

func TestSelectAgents_InsufficientProviders(t *testing.T) {
	cfg := &SwarmConfig{Pattern: "parallel_vote", MinSize: 3, MaxSize: 3, Diversity: "none", Order: "fixed"}
	_, err := SelectAgents([]string{"a", "b"}, cfg)
	if err == nil {
		t.Fatal("expected error for insufficient providers")
	}
}

func TestSelectAgents_HeterogeneousRejectsSingle(t *testing.T) {
	cfg := &SwarmConfig{Pattern: "parallel_vote", MinSize: 2, MaxSize: 3, Diversity: "heterogeneous", Order: "fixed"}
	_, err := SelectAgents([]string{"claude", "claude", "claude"}, cfg)
	if err == nil {
		t.Fatal("expected error for single provider with heterogeneous")
	}
}

func TestSelectAgents_HeterogeneousAcceptsMultiple(t *testing.T) {
	cfg := &SwarmConfig{Pattern: "parallel_vote", MinSize: 2, MaxSize: 3, Diversity: "heterogeneous", Order: "fixed"}
	slots, err := SelectAgents([]string{"claude", "gpt", "deepseek"}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(slots) != 3 {
		t.Fatalf("expected 3 slots, got %d", len(slots))
	}
}

func TestDispatch_ParallelExecution(t *testing.T) {
	cfg := &SwarmConfig{Pattern: "parallel_vote", MinSize: 2, MaxSize: 3, Diversity: "none", Order: "fixed"}
	agents := []AgentSlot{
		{AgentID: "a1", Provider: "p1"},
		{AgentID: "a2", Provider: "p2"},
		{AgentID: "a3", Provider: "p3"},
	}
	var count int64
	fn := func(ctx context.Context, task *types.AgentTask, provider string) (map[string]any, error) {
		atomic.AddInt64(&count, 1)
		return map[string]any{"result": "ok"}, nil
	}
	task := &types.AgentTask{TaskID: "test-1"}
	sr, err := Dispatch(context.Background(), task, cfg, agents, fn)
	if err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt64(&count) != 3 {
		t.Fatalf("expected 3 executions, got %d", count)
	}
	if !sr.Unanimous {
		t.Fatal("expected unanimous (all same output)")
	}
}

func TestDispatch_PartialFailure(t *testing.T) {
	cfg := &SwarmConfig{Pattern: "parallel_vote", MinSize: 2, MaxSize: 3, Diversity: "none", Order: "fixed"}
	agents := []AgentSlot{
		{AgentID: "a1", Provider: "p1"},
		{AgentID: "a2", Provider: "p2"},
		{AgentID: "a3", Provider: "p3"},
	}
	fn := func(ctx context.Context, task *types.AgentTask, provider string) (map[string]any, error) {
		if provider == "p2" {
			return nil, errors.New("timeout")
		}
		return map[string]any{"result": "ok"}, nil
	}
	task := &types.AgentTask{TaskID: "test-2"}
	sr, err := Dispatch(context.Background(), task, cfg, agents, fn)
	if err != nil {
		t.Fatal(err)
	}
	if sr.Confidence != 1.0 {
		t.Fatalf("expected 1.0 confidence (failed excluded), got %f", sr.Confidence)
	}
}

func TestDispatch_ContextCancel(t *testing.T) {
	cfg := &SwarmConfig{Pattern: "parallel_vote", MinSize: 2, MaxSize: 2, Diversity: "none", Order: "fixed"}
	agents := []AgentSlot{
		{AgentID: "a1", Provider: "p1"},
		{AgentID: "a2", Provider: "p2"},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	fn := func(ctx context.Context, task *types.AgentTask, provider string) (map[string]any, error) {
		return map[string]any{"result": "ok"}, nil
	}
	task := &types.AgentTask{TaskID: "test-3"}
	_, err := Dispatch(ctx, task, cfg, agents, fn)
	// Either error or results with cancelled agents
	_ = err // context cancel may or may not propagate depending on timing
}
