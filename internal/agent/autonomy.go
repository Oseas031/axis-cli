// Package agent provides self-context management for agent autonomy.
package agent

import (
	"time"
)

// AutonomyTransition represents a transition between autonomy levels.
type AutonomyTransition struct {
	From    AutonomyLevel
	To      AutonomyLevel
	Reason  string
	BasedOn CompetenceEvidence
}

// CompetenceEvidence contains metrics used to evaluate autonomy transitions.
type CompetenceEvidence struct {
	TasksCompleted     int
	SuccessRate        float64
	ValidationPassRate float64
	AvgExecutionTime   time.Duration
}

// NewCompetenceEvidence creates a new CompetenceEvidence with default values.
func NewCompetenceEvidence() CompetenceEvidence {
	return CompetenceEvidence{
		TasksCompleted:     0,
		SuccessRate:        0.0,
		ValidationPassRate: 0.0,
		AvgExecutionTime:   0,
	}
}

// IsComplete returns true if there is enough evidence to evaluate a transition.
func (ce CompetenceEvidence) IsComplete() bool {
	return ce.TasksCompleted > 0
}

// AutonomyRecord tracks the autonomy level history.
type AutonomyRecord struct {
	Level     AutonomyLevel
	Evidence  CompetenceEvidence
	UpdatedAt time.Time
}

// NewAutonomyRecord creates a new AutonomyRecord with the given level and evidence.
func NewAutonomyRecord(level AutonomyLevel, evidence CompetenceEvidence) *AutonomyRecord {
	return &AutonomyRecord{
		Level:     level,
		Evidence:  evidence,
		UpdatedAt: time.Now(),
	}
}

// Clone creates a deep copy of AutonomyRecord.
func (ar *AutonomyRecord) Clone() *AutonomyRecord {
	if ar == nil {
		return nil
	}
	return &AutonomyRecord{
		Level:     ar.Level,
		Evidence:  ar.Evidence,
		UpdatedAt: ar.UpdatedAt,
	}
}
