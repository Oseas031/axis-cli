// Package agent provides self-context management for agent autonomy.
package agent

import (
	"fmt"
	"log"
)

// Upgrade thresholds - minimum evidence required to consider an upgrade.
const (
	UpgradeMinTasksCompleted = 10  // Minimum tasks completed before considering upgrade
	UpgradeMinSuccessRate    = 0.8 // Success rate must exceed 80%
	UpgradeMinValidationRate = 0.9 // Validation pass rate must exceed 90%
)

// Downgrade thresholds - evidence that triggers automatic downgrade.
const (
	DowngradeMaxSuccessRate = 0.5 // Success rate below 50% triggers downgrade consideration
)

// RuleEngine evaluates competence evidence to determine autonomy transitions.
type RuleEngine struct {
	logger func(string, ...interface{})
}

// NewRuleEngine creates a new RuleEngine with optional logger.
// If logger is nil, a default logger that uses log.Printf is used.
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

// levelName returns a human-readable name for an autonomy level.
func levelName(level AutonomyLevel) string {
	return level.String()
}

// EvaluateTransition evaluates the evidence and determines the appropriate transition.
// It checks for downgrade first (critical), then upgrade, otherwise no change.
func (re *RuleEngine) EvaluateTransition(currentLevel AutonomyLevel, evidence CompetenceEvidence) AutonomyTransition {
	// Check for downgrade first (critical safety mechanism)
	if re.shouldDowngrade(currentLevel, evidence) {
		return re.evaluateDowngrade(currentLevel, evidence)
	}

	// Check for upgrade (cannot upgrade from Full)
	if re.shouldUpgrade(evidence) && currentLevel != AutonomyLevelFull {
		return re.evaluateUpgrade(currentLevel, evidence)
	}

	// No transition
	return re.logNoTransition(currentLevel, evidence)
}

// shouldUpgrade returns true if evidence meets upgrade thresholds.
// Upgrade requires: TasksCompleted >= 10 AND SuccessRate > 0.8 AND ValidationPassRate > 0.9
func (re *RuleEngine) shouldUpgrade(evidence CompetenceEvidence) bool {
	return evidence.TasksCompleted >= UpgradeMinTasksCompleted &&
		evidence.SuccessRate > UpgradeMinSuccessRate &&
		evidence.ValidationPassRate > UpgradeMinValidationRate
}

// shouldDowngrade returns true if evidence requires a downgrade.
// Downgrade triggers when: SuccessRate < 0.5 AND TasksCompleted >= 10
func (re *RuleEngine) shouldDowngrade(currentLevel AutonomyLevel, evidence CompetenceEvidence) bool {
	if currentLevel == AutonomyLevelNone {
		return false // Cannot downgrade below minimum
	}
	return evidence.SuccessRate < DowngradeMaxSuccessRate && evidence.TasksCompleted >= UpgradeMinTasksCompleted
}

// evaluateUpgrade determines the appropriate upgrade level.
func (re *RuleEngine) evaluateUpgrade(currentLevel AutonomyLevel, evidence CompetenceEvidence) AutonomyTransition {
	var reason string
	var newLevel AutonomyLevel

	switch currentLevel {
	case AutonomyLevelNone:
		newLevel = AutonomyLevelLow
		reason = "sufficient evidence for supervised execution"
	case AutonomyLevelLow:
		newLevel = AutonomyLevelMedium
		reason = fmt.Sprintf("high success rate (%.1f%%) and validation pass rate (%.1f%%)",
			evidence.SuccessRate*100, evidence.ValidationPassRate*100)
	case AutonomyLevelMedium:
		newLevel = AutonomyLevelHigh
		reason = fmt.Sprintf("excellent success rate (%.1f%%) and validation pass rate (%.1f%%)",
			evidence.SuccessRate*100, evidence.ValidationPassRate*100)
	case AutonomyLevelHigh:
		newLevel = AutonomyLevelFull
		reason = "demonstrated full competence across all dimensions"
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

// evaluateDowngrade determines the appropriate downgrade level.
func (re *RuleEngine) evaluateDowngrade(currentLevel AutonomyLevel, evidence CompetenceEvidence) AutonomyTransition {
	var reason string
	var newLevel AutonomyLevel

	switch currentLevel {
	case AutonomyLevelFull:
		newLevel = AutonomyLevelHigh
		reason = fmt.Sprintf("success rate (%.1f%%) below acceptable threshold (%.1f%%)",
			evidence.SuccessRate*100, DowngradeMaxSuccessRate*100)
	case AutonomyLevelHigh:
		newLevel = AutonomyLevelMedium
		reason = fmt.Sprintf("success rate (%.1f%%) critically low (below %.1f%%)",
			evidence.SuccessRate*100, DowngradeMaxSuccessRate*100)
	case AutonomyLevelMedium:
		newLevel = AutonomyLevelLow
		reason = fmt.Sprintf("success rate (%.1f%%) indicates need for closer supervision",
			evidence.SuccessRate*100)
	case AutonomyLevelLow:
		newLevel = AutonomyLevelNone
		reason = fmt.Sprintf("success rate (%.1f%%) indicates fundamental competence issues",
			evidence.SuccessRate*100)
	default:
		// Already at minimum
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

// logNoTransition creates a no-change transition and logs the reason.
func (re *RuleEngine) logNoTransition(currentLevel AutonomyLevel, evidence CompetenceEvidence) AutonomyTransition {
	var reason string

	if evidence.TasksCompleted == 0 {
		reason = "no tasks completed yet"
	} else if evidence.TasksCompleted < UpgradeMinTasksCompleted {
		reason = fmt.Sprintf("insufficient tasks completed (%d < %d required)",
			evidence.TasksCompleted, UpgradeMinTasksCompleted)
	} else if evidence.SuccessRate <= UpgradeMinSuccessRate {
		reason = fmt.Sprintf("success rate (%.1f%%) below upgrade threshold (%.1f%%)",
			evidence.SuccessRate*100, UpgradeMinSuccessRate*100)
	} else {
		reason = fmt.Sprintf("validation pass rate (%.1f%%) below upgrade threshold (%.1f%%)",
			evidence.ValidationPassRate*100, UpgradeMinValidationRate*100)
	}

	re.logger("AUDIT: No transition from %s, tasks=%d, success=%.2f, validation=%.2f, reason=%s",
		levelName(currentLevel), evidence.TasksCompleted,
		evidence.SuccessRate, evidence.ValidationPassRate, reason)

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
	if fromLevel == AutonomyLevelNone {
		return false
	}
	return re.shouldDowngrade(fromLevel, evidence)
}

// CalculateAutonomyDelta returns the delta between two autonomy levels.
func CalculateAutonomyDelta(from, to AutonomyLevel) int {
	return int(to) - int(from)
}
