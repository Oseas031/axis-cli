package types

import (
	"encoding/json"
	"testing"
	"time"
)

// Regression (T1): legacy JSON written before the human-agent-scheduling
// extension must unmarshal without losing the old fields.
func TestHumanCallRequest_BackwardCompat(t *testing.T) {
	legacy := `{"call_id":"c1","task_id":"t1","input":{"q":"where"},"metadata":{"executor":"human"}}`
	var req HumanCallRequest
	if err := json.Unmarshal([]byte(legacy), &req); err != nil {
		t.Fatal(err)
	}
	if req.CallID != "c1" || req.TaskID != "t1" {
		t.Fatalf("legacy ids lost: %+v", req)
	}
	if req.Input["q"] != "where" {
		t.Fatalf("legacy input lost: %+v", req.Input)
	}
	if req.Metadata["executor"] != "human" {
		t.Fatalf("legacy metadata lost: %+v", req.Metadata)
	}
	if req.Kind != "" || len(req.Options) != 0 || !req.Deadline.IsZero() {
		t.Fatalf("new fields must be zero for legacy input: %+v", req)
	}
}

// T1: new fields round-trip and Kind values stay stable.
func TestHumanCallRequest_ExtendedRoundTrip(t *testing.T) {
	deadline := time.Date(2026, 8, 27, 9, 0, 0, 0, time.UTC)
	req := HumanCallRequest{
		CallID:   "call-1",
		TaskID:   "task-1",
		Kind:     HumanKindDecision,
		Prompt:   "plan A or B?",
		Options:  []string{"A", "B"},
		Context:  map[string]any{"files": []string{"x.go"}},
		Deadline: deadline,
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var back HumanCallRequest
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.Kind != HumanKindDecision || back.Prompt != "plan A or B?" {
		t.Fatalf("extended fields lost: %+v", back)
	}
	if len(back.Options) != 2 || back.Options[1] != "B" {
		t.Fatalf("options lost: %+v", back.Options)
	}
	if !back.Deadline.Equal(deadline) {
		t.Fatalf("deadline lost: %v", back.Deadline)
	}
}

// T1: waiting_human task status and degraded call status are distinct values.
func TestNewStatusConstants(t *testing.T) {
	if TaskStatusWaitingHuman == TaskStatusRunning ||
		TaskStatusWaitingHuman == TaskStatusPending ||
		TaskStatusWaitingHuman == TaskStatusFailed {
		t.Fatal("waiting_human must be a distinct task status")
	}
	if CallStatusDegraded == CallStatusPending ||
		CallStatusDegraded == CallStatusCompleted ||
		CallStatusDegraded == CallStatusFailed {
		t.Fatal("degraded must be a distinct call status")
	}
}

// T1: preference routing keys/values keep their namespaced spellings.
func TestExecutorPreferenceKeys(t *testing.T) {
	if ExecutorKeyPreference != "axis.executor.preference" {
		t.Fatalf("preference key drift: %s", ExecutorKeyPreference)
	}
	if ExecutorPrefRequireHuman != "require-human" || ExecutorPrefRequireAgent != "require-agent" {
		t.Fatal("preference value drift")
	}
}
