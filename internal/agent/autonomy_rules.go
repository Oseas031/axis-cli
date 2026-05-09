// Package agent provides self-context management for agent autonomy.
package agent

import (
	"fmt"
	"log"
)

// Upgrade thresholds
const (
	UpgradeMinTasksCompleted = 10
	UpgradeMinSuccessRate    = 0.8
	UpgradeMinValidationRate = 0.9
)

// Downgrade thresholds
const (
	DowngradeMaxSuccessRate = 0.5
)

// RuleEngine evaluates competence evidence to determine autonomy transitions.
type RuleEngine struct {
	logger func(string, ...interface{})
}

// NewRuleEngine creates a new RuleEngine with optional logger.
func NewRuleEngine(logger func(string, ...interface{})) *RuleEngine {
	if logger == nil {
		logger = func(format string, args ...interface{}) {
			log.Printf("[AutonomyRule] "+format, args...)
		}
	}
	return &RuleEngine{
		logger: logger,
	}
}

// levelName returns a human-readable name for an autonomy level for audit logging.
func levelName(level AutonomyLevel) string {
	switch level {
	case AutonomyLevelExecute:
		return "execute"
	case AutonomyLevelDecide:
		return "decide"
	case AutonomyLevelPlan:
		return "plan"
	case AutonomyLevelLearn:
		return "learn"
	case AutonomyLevelFull:
		return "full"
	default:
		return "unknown"
	}
}

// EvaluateTransition evaluates the evidence and returns the appropriate transition.
func (re *RuleEngine) EvaluateTransition(currentLevel AutonomyLevel, evidence CompetenceEvidence) AutonomyTransition {
	// Check if already at max level - this is "insufficient evidence" since we can't go higher
	if currentLevel == AutonomyLevelFull {
		return re.logNoTransition(currentLevel, evidence)
	}

	// Check for downgrade first (critical)
	if evidence.SuccessRate < DowngradeMaxSuccessRate && evidence.TasksCompleted >= UpgradeMinTasksCompleted {
		return re.evaluateDowngrade(currentLevel, evidence)
	}

	// Check for upgrade
	if re.shouldUpgrade(evidence) {
		return re.evaluateUpgrade(currentLevel, evidence)
	}

	// No transition
	return re.logNoTransition(currentLevel, evidence)
}

func (re *RuleEngine) shouldUpgrade(evidence CompetenceEvidence) bool {
	return evidence.TasksCompleted >= UpgradeMinTasksCompleted &&
		evidence.SuccessRate >= UpgradeMinSuccessRate &&
		evidence.ValidationPassRate >= UpgradeMinValidationRate
}

func (re *RuleEngine) evaluateUpgrade(currentLevel AutonomyLevel, evidence CompetenceEvidence) AutonomyTransition {
	var reason string
	var newLevel AutonomyLevel

	switch currentLevel {
	case AutonomyLevelExecute:
		newLevel = AutonomyLevelDecide
		reason = "sufficient competence evidence"
	case AutonomyLevelDecide:
		newLevel = AutonomyLevelPlan
		reason = fmt.Sprintf("high success rate (%.1f%%) and validation pass rate (%.1f%%)",
			evidence.SuccessRate*100, evidence.ValidationPassRate*100)
	case AutonomyLevelPlan:
		newLevel = AutonomyLevelLearn
		reason = fmt.Sprintf("excellent success rate (%.1f%%) and validation pass rate (%.1f%%)",
			evidence.SuccessRate*100, evidence.ValidationPassRate*100)
	case AutonomyLevelLearn:
		newLevel = AutonomyLevelFull
		reason = "demonstrated full competence"
	default:
		// Already at max or unknown level
		return AutonomyTransition{
			From:    currentLevel,
			To:      currentLevel,
			Reason:  "already at maximum autonomy level",
			BasedOn: evidence,
		}
	}

	re.logger("AUDIT: Upgrade transition %s -> %s for %d tasks, success=%.2f, validation=%.2f",
		levelName(currentLevel), levelName(newLevel), evidence.TasksCompleted,
		evidence.SuccessRate, evidence.ValidationPassRate)

	return AutonomyTransition{
		From:    currentLevel,
		To:      newLevel,
		Reason:  reason,
		BasedOn: evidence,
	}
}

func (re *RuleEngine) evaluateDowngrade(currentLevel AutonomyLevel, evidence CompetenceEvidence) AutonomyTransition {
	var reason string
	var newLevel AutonomyLevel

	switch currentLevel {
	case AutonomyLevelFull:
		newLevel = AutonomyLevelLearn
		reason = fmt.Sprintf("success rate (%.1f%%) below threshold (%.1f%%)",
			evidence.SuccessRate*100, DowngradeMaxSuccessRate*100)
	case AutonomyLevelLearn:
		newLevel = AutonomyLevelPlan
		reason = fmt.Sprintf("success rate (%.1f%%) below threshold (%.1f%%)",
			evidence.SuccessRate*100, DowngradeMaxSuccessRate*100)
	case AutonomyLevelPlan:
		newLevel = AutonomyLevelDecide
		reason = fmt.Sprintf("success rate (%.1f%%) below threshold (%.1f%%)",
			evidence.SuccessRate*100, DowngradeMaxSuccessRate*100)
	case AutonomyLevelDecide:
		newLevel = AutonomyLevelExecute
		reason = fmt.Sprintf("success rate (%.1f%%) critically low",
			evidence.SuccessRate*100)
	default:
		// Already at min
		return AutonomyTransition{
			From:    currentLevel,
			To:      currentLevel,
			Reason:  "already at minimum autonomy level",
			BasedOn: evidence,
		}
	}

	re.logger("AUDIT: Downgrade transition %s -> %s, success rate=%.2f",
		levelName(currentLevel), levelName(newLevel), evidence.SuccessRate)

	return AutonomyTransition{
		From:    currentLevel,
		To:      newLevel,
		Reason:  reason,
		BasedOn: evidence,
	}
}

func (re *RuleEngine) logNoTransition(currentLevel AutonomyLevel, evidence CompetenceEvidence) AutonomyTransition {
	reason := "insufficient evidence for transition"
	if evidence.TasksCompleted > 0 {
		re.logger("AUDIT: No transition from %s, tasks=%d, success=%.2f, validation=%.2f",
			levelName(currentLevel), evidence.TasksCompleted,
			evidence.SuccessRate, evidence.ValidationPassRate)
	} else {
		reason = "no tasks completed yet"
	}

	return AutonomyTransition{
		From:    currentLevel,
		To:      currentLevel,
		Reason:  reason,
		BasedOn: evidence,
	}
}

// ValidateTransition validates that a transition is allowed and returns an error if not.
func (re *RuleEngine) ValidateTransition(from, to AutonomyLevel) error {
	if !from.IsValid() {
		return fmt.Errorf("invalid source autonomy level: %d", from)
	}
	if !to.IsValid() {
		return fmt.Errorf("invalid target autonomy level: %d", to)
	}
	if !from.CanTransitionTo(to) {
		return fmt.Errorf("transition from %s to %s is not allowed (max one level change)",
			from.String(), to.String())
	}
	return nil
}

// GetRequiredEvidence returns the evidence thresholds needed for an upgrade from the given level.
func (re *RuleEngine) GetRequiredEvidence(fromLevel AutonomyLevel) map[string]interface{} {
	return map[string]interface{}{
		"min_tasks_completed":      UpgradeMinTasksCompleted,
		"min_success_rate":         UpgradeMinSuccessRate,
		"min_validation_pass_rate": UpgradeMinValidationRate,
		"current_level":            fromLevel.String(),
	}
}

// RecordTransition creates an AutonomyRecord from a transition.
func (re *RuleEngine) RecordTransition(transition AutonomyTransition) *AutonomyRecord {
	return NewAutonomyRecord(transition.To, transition.BasedOn)
}

// CanUpgrade checks if the evidence is sufficient for an upgrade from the given level.
func (re *RuleEngine) CanUpgrade(fromLevel AutonomyLevel, evidence CompetenceEvidence) bool {
	if fromLevel == AutonomyLevelFull {
		return false
	}
	return re.shouldUpgrade(evidence)
}

// CanDowngrade checks if the evidence requires a downgrade from the given level.
func (re *RuleEngine) CanDowngrade(fromLevel AutonomyLevel, evidence CompetenceEvidence) bool {
	if fromLevel == AutonomyLevelExecute {
		return false
	}
	return evidence.SuccessRate < DowngradeMaxSuccessRate && evidence.TasksCompleted >= UpgradeMinTasksCompleted
}

// CalculateAutonomyDelta returns the delta between two autonomy levels.
func CalculateAutonomyDelta(from, to AutonomyLevel) int {
	return int(to) - int(from)
}

// NewAutonomyTransition creates an AutonomyTransition with validation.
func NewAutonomyTransition(from, to AutonomyLevel, reason string, evidence CompetenceEvidence) AutonomyTransition {
	re := NewRuleEngine(nil)
	if err := re.ValidateTransition(from, to); err != nil {
		return AutonomyTransition{
			From:    from,
			To:      from,
			Reason:  fmt.Sprintf("invalid transition: %s", err),
			BasedOn: evidence,
		}
	}

	return AutonomyTransition{
		From:    from,
		To:      to,
		Reason:  reason,
		BasedOn: evidence,
	}
}

// WithEvidence creates a copy of the transition with updated evidence.
func (t AutonomyTransition) WithEvidence(evidence CompetenceEvidence) AutonomyTransition {
	t.BasedOn = evidence
	return t
}

// GetDelta returns the autonomy level change as an integer.
func (t AutonomyTransition) GetDelta() int {
	return CalculateAutonomyDelta(t.From, t.To)
}

// IsUpgrade returns true if this is an upgrade transition.
func (t AutonomyTransition) IsUpgrade() bool {
	return t.To > t.From
}

// IsDowngrade returns true if this is a downgrade transition.
func (t AutonomyTransition) IsDowngrade() bool {
	return t.To < t.From
}

// IsNoChange returns true if there is no change in autonomy level.
func (t AutonomyTransition) IsNoChange() bool {
	return t.From == t.To
}

// String returns a string representation of the transition.
func (t AutonomyTransition) String() string {
	return fmt.Sprintf("AutonomyTransition{%s -> %s, reason=%q, delta=%d}",
		t.From.String(), t.To.String(), t.Reason, t.GetDelta())
}

// IsSignificant returns true if this is a meaningful transition (not no-change).
func (t AutonomyTransition) IsSignificant() bool {
	return !t.IsNoChange()
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
