// Package agent provides self-context management for agent autonomy.
package agent

import (
	"time"
)

// AutonomyLevel represents the level of autonomy an agent has earned.
// Higher levels indicate more independence in decision-making.
type AutonomyLevel int

const (
	AutonomyLevelNone   AutonomyLevel = 0 // No autonomy - requires full supervision
	AutonomyLevelLow    AutonomyLevel = 1 // Can execute tasks with approval
	AutonomyLevelMedium AutonomyLevel = 2 // Can decide approach with oversight
	AutonomyLevelHigh   AutonomyLevel = 3 // Can plan and execute autonomously
	AutonomyLevelFull   AutonomyLevel = 4 // Maximum autonomy - near complete independence
)

// String returns a human-readable name for the autonomy level.
func (a AutonomyLevel) String() string {
	switch a {
	case AutonomyLevelNone:
		return "none"
	case AutonomyLevelLow:
		return "low"
	case AutonomyLevelMedium:
		return "medium"
	case AutonomyLevelHigh:
		return "high"
	case AutonomyLevelFull:
		return "full"
	default:
		return "unknown"
	}
}

// IsValid checks if the autonomy level is within valid bounds (0-4).
func (a AutonomyLevel) IsValid() bool {
	return a >= AutonomyLevelNone && a <= AutonomyLevelFull
}

// CanTransitionTo checks if a transition to the target level is valid.
// Only adjacent level transitions (delta of 1) are allowed.
func (a AutonomyLevel) CanTransitionTo(target AutonomyLevel) bool {
	if !target.IsValid() {
		return false
	}
	diff := int(target) - int(a)
	if diff < 0 {
		diff = -diff
	}
	return diff <= 1
}

// IsUpgrade returns true if the target level is higher than the current level.
func (a AutonomyLevel) IsUpgrade(target AutonomyLevel) bool {
	return target > a
}

// IsDowngrade returns true if the target level is lower than the current level.
func (a AutonomyLevel) IsDowngrade(target AutonomyLevel) bool {
	return target < a
}

// CompetenceEvidence contains metrics used to evaluate autonomy transitions.
type CompetenceEvidence struct {
	TasksCompleted     int           // Total number of tasks completed
	SuccessRate        float64       // Ratio of successful tasks (0.0 to 1.0)
	ValidationPassRate float64       // Ratio of validations passed (0.0 to 1.0)
	AvgExecutionTime   time.Duration // Average time to complete tasks
}

// NewCompetenceEvidence creates a new CompetenceEvidence with zero values.
func NewCompetenceEvidence() CompetenceEvidence {
	return CompetenceEvidence{
		TasksCompleted:     0,
		SuccessRate:        0.0,
		ValidationPassRate: 0.0,
		AvgExecutionTime:   0,
	}
}

// IsComplete returns true if there is enough evidence to evaluate a transition.
// Evidence is considered complete when at least one task has been completed.
func (ce CompetenceEvidence) IsComplete() bool {
	return ce.TasksCompleted > 0
}

// Clone creates a deep copy of CompetenceEvidence.
func (ce CompetenceEvidence) Clone() CompetenceEvidence {
	return CompetenceEvidence{
		TasksCompleted:     ce.TasksCompleted,
		SuccessRate:        ce.SuccessRate,
		ValidationPassRate: ce.ValidationPassRate,
		AvgExecutionTime:   ce.AvgExecutionTime,
	}
}

// MeetsUpgradeThresholds returns true if evidence meets minimum thresholds for upgrade consideration.
func (ce CompetenceEvidence) MeetsUpgradeThresholds() bool {
	return ce.TasksCompleted >= 10 &&
		ce.SuccessRate > 0.8 &&
		ce.ValidationPassRate > 0.9
}

// IndicatesDowngrade returns true if evidence suggests a downgrade is warranted.
func (ce CompetenceEvidence) IndicatesDowngrade() bool {
	return ce.SuccessRate < 0.5 && ce.TasksCompleted >= 10
}

// AutonomyTransition represents a transition between autonomy levels.
type AutonomyTransition struct {
	From    AutonomyLevel      // The previous autonomy level
	To      AutonomyLevel      // The new autonomy level
	Reason  string             // Human-readable reason for the transition
	BasedOn CompetenceEvidence // Evidence used to evaluate the transition
}

// NewAutonomyTransition creates an AutonomyTransition with the given parameters.
func NewAutonomyTransition(from, to AutonomyLevel, reason string, evidence CompetenceEvidence) AutonomyTransition {
	return AutonomyTransition{
		From:    from,
		To:      to,
		Reason:  reason,
		BasedOn: evidence,
	}
}

// IsUpgrade returns true if this is an upgrade transition (to > from).
func (t AutonomyTransition) IsUpgrade() bool {
	return t.To > t.From
}

// IsDowngrade returns true if this is a downgrade transition (to < from).
func (t AutonomyTransition) IsDowngrade() bool {
	return t.To < t.From
}

// IsNoChange returns true if there is no change in autonomy level.
func (t AutonomyTransition) IsNoChange() bool {
	return t.From == t.To
}

// IsSignificant returns true if this is a meaningful transition (not no-change).
func (t AutonomyTransition) IsSignificant() bool {
	return !t.IsNoChange()
}

// GetDelta returns the autonomy level change as an integer.
func (t AutonomyTransition) GetDelta() int {
	return int(t.To) - int(t.From)
}

// WithEvidence creates a copy of the transition with updated evidence.
func (t AutonomyTransition) WithEvidence(evidence CompetenceEvidence) AutonomyTransition {
	t.BasedOn = evidence
	return t
}

// EvidenceMetrics returns key metrics from the transition's evidence.
func (t AutonomyTransition) EvidenceMetrics() map[string]interface{} {
	return map[string]interface{}{
		"tasks_completed":       t.BasedOn.TasksCompleted,
		"success_rate":          t.BasedOn.SuccessRate,
		"validation_pass_rate":  t.BasedOn.ValidationPassRate,
		"avg_execution_time_ms": t.BasedOn.AvgExecutionTime.Milliseconds(),
	}
}

// String returns a string representation of the transition.
func (t AutonomyTransition) String() string {
	return "AutonomyTransition{" + t.From.String() + " -> " + t.To.String() +
		", reason=" + "\"" + t.Reason + "\", delta=" + string(rune('0'+t.GetDelta())) + "}"
}

// AutonomyRecord tracks the autonomy level history with evidence.
type AutonomyRecord struct {
	Level     AutonomyLevel      // Current or final autonomy level
	Evidence  CompetenceEvidence // Evidence that led to this record
	UpdatedAt time.Time          // Timestamp of the last update
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
		Evidence:  ar.Evidence.Clone(),
		UpdatedAt: ar.UpdatedAt,
	}
}

// IsStale returns true if the record has not been updated within the given duration.
func (ar *AutonomyRecord) IsStale(threshold time.Duration) bool {
	if ar == nil {
		return true
	}
	return time.Since(ar.UpdatedAt) > threshold
}

// Update updates the record with new evidence and potentially new level.
func (ar *AutonomyRecord) Update(level AutonomyLevel, evidence CompetenceEvidence) {
	ar.Level = level
	ar.Evidence = evidence
	ar.UpdatedAt = time.Now()
}
