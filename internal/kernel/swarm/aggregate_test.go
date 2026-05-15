package swarm

import (
	"math"
	"testing"
)

func TestAggregate_Unanimous(t *testing.T) {
	results := []SingleResult{
		{AgentID: "a1", Output: map[string]any{"answer": "yes"}},
		{AgentID: "a2", Output: map[string]any{"answer": "yes"}},
		{AgentID: "a3", Output: map[string]any{"answer": "yes"}},
	}
	sr, err := Aggregate(results)
	if err != nil {
		t.Fatal(err)
	}
	if !sr.Unanimous {
		t.Fatal("expected unanimous")
	}
	if sr.Confidence != 1.0 {
		t.Fatalf("expected confidence 1.0, got %f", sr.Confidence)
	}
}

func TestAggregate_MajorityVote(t *testing.T) {
	results := []SingleResult{
		{AgentID: "a1", Output: map[string]any{"answer": "yes"}},
		{AgentID: "a2", Output: map[string]any{"answer": "yes"}},
		{AgentID: "a3", Output: map[string]any{"answer": "no"}},
	}
	sr, err := Aggregate(results)
	if err != nil {
		t.Fatal(err)
	}
	if sr.Unanimous {
		t.Fatal("should not be unanimous")
	}
	expected := 2.0 / 3.0
	if math.Abs(sr.Confidence-expected) > 0.01 {
		t.Fatalf("expected confidence ~0.67, got %f", sr.Confidence)
	}
	if sr.Winner.Output["answer"] != "yes" {
		t.Fatal("wrong winner")
	}
}

func TestAggregate_AllDifferent(t *testing.T) {
	results := []SingleResult{
		{AgentID: "a1", Output: map[string]any{"answer": "a"}},
		{AgentID: "a2", Output: map[string]any{"answer": "b"}},
		{AgentID: "a3", Output: map[string]any{"answer": "c"}},
	}
	sr, err := Aggregate(results)
	if err != nil {
		t.Fatal(err)
	}
	// All different: confidence = 1/3
	expected := 1.0 / 3.0
	if math.Abs(sr.Confidence-expected) > 0.01 {
		t.Fatalf("expected confidence ~0.33, got %f", sr.Confidence)
	}
}

func TestAggregate_AllFailed(t *testing.T) {
	results := []SingleResult{
		{AgentID: "a1", Error: "timeout"},
		{AgentID: "a2", Error: "crash"},
	}
	_, err := Aggregate(results)
	if err == nil {
		t.Fatal("expected error when all agents failed")
	}
}

func TestAggregate_Empty(t *testing.T) {
	_, err := Aggregate(nil)
	if err == nil {
		t.Fatal("expected error for empty results")
	}
}

func TestAggregate_PartialFailure(t *testing.T) {
	results := []SingleResult{
		{AgentID: "a1", Output: map[string]any{"answer": "yes"}},
		{AgentID: "a2", Error: "timeout"},
		{AgentID: "a3", Output: map[string]any{"answer": "yes"}},
	}
	sr, err := Aggregate(results)
	if err != nil {
		t.Fatal(err)
	}
	if sr.Confidence != 1.0 {
		t.Fatalf("expected 1.0 confidence (failed excluded), got %f", sr.Confidence)
	}
}
