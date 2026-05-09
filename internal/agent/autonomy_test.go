// Package agent provides agent autonomy and transition rules.
package agent

import (
	"testing"
	"time"
)

// Tests for AutonomyLevel type and methods

func TestAutonomyLevel_String(t *testing.T) {
	tests := []struct {
		level    AutonomyLevel
		expected string
	}{
		{AutonomyLevelExecute, "execute"},
		{AutonomyLevelDecide, "decide"},
		{AutonomyLevelPlan, "plan"},
		{AutonomyLevelLearn, "learn"},
		{AutonomyLevelFull, "full"},
		{AutonomyLevel(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("AutonomyLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAutonomyLevel_IsValid(t *testing.T) {
	tests := []struct {
		level    AutonomyLevel
		expected bool
	}{
		{AutonomyLevelExecute, true},
		{AutonomyLevelDecide, true},
		{AutonomyLevelPlan, true},
		{AutonomyLevelLearn, true},
		{AutonomyLevelFull, true},
		{AutonomyLevel(-1), false},
		{AutonomyLevel(5), false},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			if got := tt.level.IsValid(); got != tt.expected {
				t.Errorf("AutonomyLevel.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAutonomyLevel_CanTransitionTo(t *testing.T) {
	tests := []struct {
		from     AutonomyLevel
		to       AutonomyLevel
		expected bool
	}{
		// Valid one-level transitions
		{AutonomyLevelExecute, AutonomyLevelDecide, true},
		{AutonomyLevelDecide, AutonomyLevelPlan, true},
		{AutonomyLevelPlan, AutonomyLevelLearn, true},
		{AutonomyLevelLearn, AutonomyLevelFull, true},
		// Same level is valid (no change)
		{AutonomyLevelDecide, AutonomyLevelDecide, true},
		// Invalid multi-level transitions
		{AutonomyLevelExecute, AutonomyLevelPlan, false},
		{AutonomyLevelDecide, AutonomyLevelFull, false},
		// Downgrade is one level
		{AutonomyLevelFull, AutonomyLevelLearn, true},
		{AutonomyLevelLearn, AutonomyLevelExecute, false},
		{AutonomyLevel(99), AutonomyLevelDecide, false},
		{AutonomyLevelDecide, AutonomyLevel(99), false},
	}

	for _, tt := range tests {
		t.Run(tt.from.String()+"_to_"+tt.to.String(), func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.expected {
				t.Errorf("AutonomyLevel.CanTransitionTo() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Tests for CompetenceEvidence

func TestNewCompetenceEvidence(t *testing.T) {
	ce := NewCompetenceEvidence()

	if ce.TasksCompleted != 0 {
		t.Errorf("TasksCompleted = %v, want 0", ce.TasksCompleted)
	}
	if ce.SuccessRate != 0.0 {
		t.Errorf("SuccessRate = %v, want 0.0", ce.SuccessRate)
	}
	if ce.ValidationPassRate != 0.0 {
		t.Errorf("ValidationPassRate = %v, want 0.0", ce.ValidationPassRate)
	}
	if ce.AvgExecutionTime != 0 {
		t.Errorf("AvgExecutionTime = %v, want 0", ce.AvgExecutionTime)
	}
}

func TestCompetenceEvidence_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		evidence CompetenceEvidence
		expected bool
	}{
		{"empty", CompetenceEvidence{}, false},
		{"zero_tasks", CompetenceEvidence{TasksCompleted: 0, SuccessRate: 0.9}, false},
		{"positive_tasks", CompetenceEvidence{TasksCompleted: 1, SuccessRate: 0.9}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.evidence.IsComplete(); got != tt.expected {
				t.Errorf("CompetenceEvidence.IsComplete() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Tests for AutonomyRecord

func TestAutonomyRecord_Clone(t *testing.T) {
	original := &AutonomyRecord{
		Level: AutonomyLevelPlan,
		Evidence: CompetenceEvidence{
			TasksCompleted:     15,
			SuccessRate:        0.85,
			ValidationPassRate: 0.92,
			AvgExecutionTime:   100 * time.Millisecond,
		},
		UpdatedAt: time.Now(),
	}

	clone := original.Clone()

	if clone == original {
		t.Error("Clone should return a new pointer")
	}
	if clone.Level != original.Level {
		t.Errorf("Level = %v, want %v", clone.Level, original.Level)
	}
	if clone.Evidence.TasksCompleted != original.Evidence.TasksCompleted {
		t.Errorf("Evidence.TasksCompleted = %v, want %v", clone.Evidence.TasksCompleted, original.Evidence.TasksCompleted)
	}
}

func TestAutonomyRecord_CloneNil(t *testing.T) {
	var record *AutonomyRecord
	clone := record.Clone()

	if clone != nil {
		t.Error("Clone of nil should return nil")
	}
}

func TestNewAutonomyRecord(t *testing.T) {
	evidence := CompetenceEvidence{
		TasksCompleted:     20,
		SuccessRate:        0.9,
		ValidationPassRate: 0.95,
	}
	level := AutonomyLevelLearn

	record := NewAutonomyRecord(level, evidence)

	if record.Level != level {
		t.Errorf("Level = %v, want %v", record.Level, level)
	}
	if record.Evidence.TasksCompleted != evidence.TasksCompleted {
		t.Errorf("Evidence.TasksCompleted = %v, want %v", record.Evidence.TasksCompleted, evidence.TasksCompleted)
	}
	if record.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}
