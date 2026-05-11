package evolution

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEvolutionStatus_Values(t *testing.T) {
	statuses := []EvolutionStatus{
		StatusPending, StatusRunning, StatusCompleted,
		StatusFailed, StatusDiscarded, StatusPromoted, StatusPaused,
	}
	for _, s := range statuses {
		if s == "" {
			t.Error("expected non-empty status value")
		}
	}
}

func TestRiskLevel_Values(t *testing.T) {
	levels := []RiskLevel{RiskLow, RiskMedium, RiskHigh}
	for _, l := range levels {
		if l == "" {
			t.Error("expected non-empty risk level value")
		}
	}
}

func TestEvolutionIntent_JSONRoundTrip(t *testing.T) {
	original := EvolutionIntent{
		ID:           "intent-1",
		CreatedAt:    time.Date(2026, 5, 11, 9, 0, 0, 0, time.UTC),
		Actor:        "agent",
		Summary:      "refactor scheduler",
		TargetDomain: "kernel/scheduler",
		RiskLevel:    RiskMedium,
		Status:       StatusPending,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded EvolutionIntent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: %s != %s", decoded.ID, original.ID)
	}
	if decoded.Actor != original.Actor {
		t.Errorf("Actor mismatch")
	}
	if decoded.Status != original.Status {
		t.Errorf("Status mismatch")
	}
	if decoded.RiskLevel != original.RiskLevel {
		t.Errorf("RiskLevel mismatch")
	}
}

func TestEvolutionRun_ZeroValue(t *testing.T) {
	var run EvolutionRun
	if run.RunID != "" || run.Status != "" || run.CreatedAt != (time.Time{}) {
		t.Error("expected zero values")
	}
}

func TestEvolutionStep_JSONRoundTrip(t *testing.T) {
	now := time.Now()
	original := EvolutionStep{
		StepID:      "step-1",
		RunID:       "run-1",
		Sequence:    1,
		TargetPath:  "internal/scheduler/scheduler.go",
		Action:      StepActionPatch,
		PatchRef:    "patches/0001.patch",
		Status:      StatusCompleted,
		StartedAt:   &now,
		CompletedAt: &now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded EvolutionStep
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.StepID != original.StepID {
		t.Errorf("StepID mismatch")
	}
	if decoded.Sequence != original.Sequence {
		t.Errorf("Sequence mismatch")
	}
	if decoded.Action != original.Action {
		t.Errorf("Action mismatch")
	}
	if decoded.Status != original.Status {
		t.Errorf("Status mismatch")
	}
}

func TestVerificationRecord_JSONRoundTrip(t *testing.T) {
	now := time.Now()
	original := VerificationRecord{
		RunID:     "run-1",
		Command:   "go test ./internal/kernel/scheduler",
		StartedAt: now,
		ExitCode:  0,
		StdoutRef: "stdout.txt",
		StderrRef: "stderr.txt",
		Status:    VerificationPassed,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded VerificationRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.RunID != original.RunID {
		t.Errorf("RunID mismatch")
	}
	if decoded.ExitCode != original.ExitCode {
		t.Errorf("ExitCode mismatch")
	}
	if decoded.Status != original.Status {
		t.Errorf("Status mismatch")
	}
}

func TestEvolutionDecision_JSONRoundTrip(t *testing.T) {
	original := EvolutionDecision{
		RunID:     "run-1",
		Decision:  DecisionPromoted,
		Actor:     "human",
		Reason:    "tests passed and code review approved",
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded EvolutionDecision
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Decision != original.Decision {
		t.Errorf("Decision mismatch")
	}
	if decoded.Reason != original.Reason {
		t.Errorf("Reason mismatch")
	}
}

func TestStepAction_Values(t *testing.T) {
	actions := []StepAction{
		StepActionPatch, StepActionCreate, StepActionDelete,
		StepActionVerify, StepActionPromote, StepActionDiscard,
	}
	for _, a := range actions {
		if a == "" {
			t.Error("expected non-empty action value")
		}
	}
}

func TestDecisionType_Values(t *testing.T) {
	decisions := []DecisionType{DecisionPromoted, DecisionDiscarded, DecisionPaused}
	for _, d := range decisions {
		if d == "" {
			t.Error("expected non-empty decision type value")
		}
	}
}

func TestVerificationStatus_Values(t *testing.T) {
	statuses := []VerificationStatus{
		VerificationPending, VerificationPassed,
		VerificationFailed, VerificationCancelled,
	}
	for _, s := range statuses {
		if s == "" {
			t.Error("expected non-empty verification status value")
		}
	}
}
