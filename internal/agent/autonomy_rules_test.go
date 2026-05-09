// Package agent provides agent autonomy and transition rules.
package agent

import (
	"strings"
	"testing"
	"time"
)

// Tests for RuleEngine

func TestRuleEngine_EvaluateTransition_NoChange(t *testing.T) {
	re := NewRuleEngine(nil)

	evidence := CompetenceEvidence{
		TasksCompleted:     5,
		SuccessRate:        0.7,
		ValidationPassRate: 0.8,
	}

	transition := re.EvaluateTransition(AutonomyLevelDecide, evidence)

	if !transition.IsNoChange() {
		t.Error("Expected no change transition")
	}
	if transition.From != AutonomyLevelDecide {
		t.Errorf("From = %v, want %v", transition.From, AutonomyLevelDecide)
	}
	if transition.To != AutonomyLevelDecide {
		t.Errorf("To = %v, want %v", transition.To, AutonomyLevelDecide)
	}
}

func TestRuleEngine_EvaluateTransition_Upgrade(t *testing.T) {
	re := NewRuleEngine(nil)

	evidence := CompetenceEvidence{
		TasksCompleted:     15,
		SuccessRate:        0.85,
		ValidationPassRate: 0.95,
	}

	transition := re.EvaluateTransition(AutonomyLevelExecute, evidence)

	if !transition.IsUpgrade() {
		t.Error("Expected upgrade transition")
	}
	if transition.From != AutonomyLevelExecute {
		t.Errorf("From = %v, want %v", transition.From, AutonomyLevelExecute)
	}
	if transition.To != AutonomyLevelDecide {
		t.Errorf("To = %v, want %v", transition.To, AutonomyLevelDecide)
	}
}

func TestRuleEngine_EvaluateTransition_MultipleUpgrades(t *testing.T) {
	re := NewRuleEngine(nil)

	evidence := CompetenceEvidence{
		TasksCompleted:     50,
		SuccessRate:        0.95,
		ValidationPassRate: 0.98,
	}

	current := AutonomyLevelExecute

	// Execute -> Decide
	transition := re.EvaluateTransition(current, evidence)
	if transition.To != AutonomyLevelDecide {
		t.Errorf("Step 1: To = %v, want %v", transition.To, AutonomyLevelDecide)
	}
	current = transition.To

	// Decide -> Plan
	transition = re.EvaluateTransition(current, evidence)
	if transition.To != AutonomyLevelPlan {
		t.Errorf("Step 2: To = %v, want %v", transition.To, AutonomyLevelPlan)
	}
	current = transition.To

	// Plan -> Learn
	transition = re.EvaluateTransition(current, evidence)
	if transition.To != AutonomyLevelLearn {
		t.Errorf("Step 3: To = %v, want %v", transition.To, AutonomyLevelLearn)
	}
	current = transition.To

	// Learn -> Full
	transition = re.EvaluateTransition(current, evidence)
	if transition.To != AutonomyLevelFull {
		t.Errorf("Step 4: To = %v, want %v", transition.To, AutonomyLevelFull)
	}
}

func TestRuleEngine_EvaluateTransition_Downgrade(t *testing.T) {
	re := NewRuleEngine(nil)

	evidence := CompetenceEvidence{
		TasksCompleted:     15,
		SuccessRate:        0.4, // below DowngradeMaxSuccessRate
		ValidationPassRate: 0.5,
	}

	transition := re.EvaluateTransition(AutonomyLevelPlan, evidence)

	if !transition.IsDowngrade() {
		t.Error("Expected downgrade transition")
	}
	if transition.From != AutonomyLevelPlan {
		t.Errorf("From = %v, want %v", transition.From, AutonomyLevelPlan)
	}
	if transition.To != AutonomyLevelDecide {
		t.Errorf("To = %v, want %v", transition.To, AutonomyLevelDecide)
	}
}

func TestRuleEngine_EvaluateTransition_NoUpgradeAtMaxLevel(t *testing.T) {
	re := NewRuleEngine(nil)

	evidence := CompetenceEvidence{
		TasksCompleted:     100,
		SuccessRate:        0.99,
		ValidationPassRate: 0.99,
	}

	transition := re.EvaluateTransition(AutonomyLevelFull, evidence)

	if !transition.IsNoChange() {
		t.Error("Expected no change at max level")
	}
	if transition.Reason != "insufficient evidence for transition" {
		t.Errorf("Reason = %v, want 'insufficient evidence for transition'", transition.Reason)
	}
}

func TestRuleEngine_EvaluateTransition_NoDowngradeAtMinLevel(t *testing.T) {
	re := NewRuleEngine(nil)

	evidence := CompetenceEvidence{
		TasksCompleted:     50,
		SuccessRate:        0.1, // very low
		ValidationPassRate: 0.2,
	}

	transition := re.EvaluateTransition(AutonomyLevelExecute, evidence)

	if !transition.IsNoChange() {
		t.Error("Expected no change at min level")
	}
	if transition.Reason != "already at minimum autonomy level" {
		t.Errorf("Reason = %v, want 'already at minimum autonomy level'", transition.Reason)
	}
}

func TestRuleEngine_ValidateTransition(t *testing.T) {
	re := NewRuleEngine(nil)

	tests := []struct {
		name    string
		from    AutonomyLevel
		to      AutonomyLevel
		wantErr bool
	}{
		{"valid upgrade", AutonomyLevelDecide, AutonomyLevelPlan, false},
		{"valid downgrade", AutonomyLevelLearn, AutonomyLevelPlan, false},
		{"no change", AutonomyLevelPlan, AutonomyLevelPlan, false},
		{"invalid multi-level upgrade", AutonomyLevelExecute, AutonomyLevelPlan, true},
		{"invalid multi-level downgrade", AutonomyLevelLearn, AutonomyLevelExecute, true},
		{"invalid from level", AutonomyLevel(99), AutonomyLevelDecide, true},
		{"invalid to level", AutonomyLevelDecide, AutonomyLevel(99), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := re.ValidateTransition(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTransition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRuleEngine_CanUpgrade(t *testing.T) {
	re := NewRuleEngine(nil)

	evidence := CompetenceEvidence{
		TasksCompleted:     15,
		SuccessRate:        0.85,
		ValidationPassRate: 0.95,
	}

	if !re.CanUpgrade(AutonomyLevelDecide, evidence) {
		t.Error("Expected CanUpgrade to be true with sufficient evidence")
	}

	if re.CanUpgrade(AutonomyLevelFull, evidence) {
		t.Error("Expected CanUpgrade to be false at max level")
	}

	insufficientEvidence := CompetenceEvidence{
		TasksCompleted:     5,
		SuccessRate:        0.6,
		ValidationPassRate: 0.7,
	}
	if re.CanUpgrade(AutonomyLevelDecide, insufficientEvidence) {
		t.Error("Expected CanUpgrade to be false with insufficient evidence")
	}
}

func TestRuleEngine_CanDowngrade(t *testing.T) {
	re := NewRuleEngine(nil)

	evidence := CompetenceEvidence{
		TasksCompleted:     15,
		SuccessRate:        0.3, // below threshold
		ValidationPassRate: 0.5,
	}

	if !re.CanDowngrade(AutonomyLevelPlan, evidence) {
		t.Error("Expected CanDowngrade to be true with low success rate")
	}

	if re.CanDowngrade(AutonomyLevelExecute, evidence) {
		t.Error("Expected CanDowngrade to be false at min level")
	}

	// With sufficient success rate
	goodEvidence := CompetenceEvidence{
		TasksCompleted:     15,
		SuccessRate:        0.7,
		ValidationPassRate: 0.8,
	}
	if re.CanDowngrade(AutonomyLevelPlan, goodEvidence) {
		t.Error("Expected CanDowngrade to be false with good success rate")
	}
}

func TestRuleEngine_GetRequiredEvidence(t *testing.T) {
	re := NewRuleEngine(nil)

	required := re.GetRequiredEvidence(AutonomyLevelDecide)

	if required["min_tasks_completed"] != UpgradeMinTasksCompleted {
		t.Errorf("min_tasks_completed = %v, want %v", required["min_tasks_completed"], UpgradeMinTasksCompleted)
	}
	if required["min_success_rate"] != UpgradeMinSuccessRate {
		t.Errorf("min_success_rate = %v, want %v", required["min_success_rate"], UpgradeMinSuccessRate)
	}
	if required["min_validation_pass_rate"] != UpgradeMinValidationRate {
		t.Errorf("min_validation_pass_rate = %v, want %v", required["min_validation_pass_rate"], UpgradeMinValidationRate)
	}
}

func TestRuleEngine_RecordTransition(t *testing.T) {
	re := NewRuleEngine(nil)

	transition := AutonomyTransition{
		From:   AutonomyLevelDecide,
		To:     AutonomyLevelPlan,
		Reason: "test upgrade",
		BasedOn: CompetenceEvidence{
			TasksCompleted:     20,
			SuccessRate:        0.9,
			ValidationPassRate: 0.95,
		},
	}

	record := re.RecordTransition(transition)

	if record.Level != transition.To {
		t.Errorf("Level = %v, want %v", record.Level, transition.To)
	}
	if record.Evidence.TasksCompleted != transition.BasedOn.TasksCompleted {
		t.Errorf("Evidence.TasksCompleted = %v, want %v", record.Evidence.TasksCompleted, transition.BasedOn.TasksCompleted)
	}
}

func TestCalculateAutonomyDelta(t *testing.T) {
	tests := []struct {
		from     AutonomyLevel
		to       AutonomyLevel
		expected int
	}{
		{AutonomyLevelExecute, AutonomyLevelDecide, 1},
		{AutonomyLevelDecide, AutonomyLevelPlan, 1},
		{AutonomyLevelPlan, AutonomyLevelLearn, 1},
		{AutonomyLevelLearn, AutonomyLevelFull, 1},
		{AutonomyLevelFull, AutonomyLevelLearn, -1},
		{AutonomyLevelPlan, AutonomyLevelPlan, 0},
		{AutonomyLevelExecute, AutonomyLevelFull, 4},
	}

	for _, tt := range tests {
		t.Run(tt.from.String()+"_to_"+tt.to.String(), func(t *testing.T) {
			if got := CalculateAutonomyDelta(tt.from, tt.to); got != tt.expected {
				t.Errorf("CalculateAutonomyDelta() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewAutonomyTransition_Valid(t *testing.T) {
	evidence := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.85,
		ValidationPassRate: 0.92,
	}

	transition := NewAutonomyTransition(AutonomyLevelDecide, AutonomyLevelPlan, "test", evidence)

	if transition.From != AutonomyLevelDecide {
		t.Errorf("From = %v, want %v", transition.From, AutonomyLevelDecide)
	}
	if transition.To != AutonomyLevelPlan {
		t.Errorf("To = %v, want %v", transition.To, AutonomyLevelPlan)
	}
	if transition.GetDelta() != 1 {
		t.Errorf("GetDelta() = %v, want %v", transition.GetDelta(), 1)
	}
}

func TestNewAutonomyTransition_Invalid(t *testing.T) {
	evidence := CompetenceEvidence{}

	// Multi-level jump should be rejected
	transition := NewAutonomyTransition(AutonomyLevelExecute, AutonomyLevelPlan, "invalid", evidence)

	if transition.To != AutonomyLevelExecute {
		t.Errorf("To should remain at From for invalid transition, got %v", transition.To)
	}
	if !strings.Contains(transition.Reason, "invalid transition") {
		t.Errorf("Reason should contain 'invalid transition', got %q", transition.Reason)
	}
}

func TestAutonomyTransition_IsUpgrade(t *testing.T) {
	transition := AutonomyTransition{
		From:   AutonomyLevelDecide,
		To:     AutonomyLevelPlan,
		Reason: "test",
	}

	if !transition.IsUpgrade() {
		t.Error("Expected IsUpgrade to be true")
	}
	if transition.IsDowngrade() {
		t.Error("Expected IsDowngrade to be false")
	}
}

func TestAutonomyTransition_IsDowngrade(t *testing.T) {
	transition := AutonomyTransition{
		From:   AutonomyLevelPlan,
		To:     AutonomyLevelDecide,
		Reason: "test",
	}

	if !transition.IsDowngrade() {
		t.Error("Expected IsDowngrade to be true")
	}
	if transition.IsUpgrade() {
		t.Error("Expected IsUpgrade to be false")
	}
}

func TestAutonomyTransition_IsNoChange(t *testing.T) {
	transition := AutonomyTransition{
		From:   AutonomyLevelPlan,
		To:     AutonomyLevelPlan,
		Reason: "no change",
	}

	if !transition.IsNoChange() {
		t.Error("Expected IsNoChange to be true")
	}
}

func TestAutonomyTransition_IsSignificant(t *testing.T) {
	noChange := AutonomyTransition{
		From:   AutonomyLevelPlan,
		To:     AutonomyLevelPlan,
		Reason: "no change",
	}

	upgrade := AutonomyTransition{
		From:   AutonomyLevelDecide,
		To:     AutonomyLevelPlan,
		Reason: "upgrade",
	}

	if noChange.IsSignificant() {
		t.Error("Expected IsSignificant to be false for no change")
	}
	if !upgrade.IsSignificant() {
		t.Error("Expected IsSignificant to be true for upgrade")
	}
}

func TestAutonomyTransition_WithEvidence(t *testing.T) {
	transition := AutonomyTransition{
		From:   AutonomyLevelDecide,
		To:     AutonomyLevelPlan,
		Reason: "test",
	}

	newEvidence := CompetenceEvidence{
		TasksCompleted:     25,
		SuccessRate:        0.88,
		ValidationPassRate: 0.93,
		AvgExecutionTime:   150 * time.Millisecond,
	}

	newTransition := transition.WithEvidence(newEvidence)

	if newTransition.From != transition.From {
		t.Errorf("From should be preserved")
	}
	if newTransition.To != transition.To {
		t.Errorf("To should be preserved")
	}
	if newTransition.BasedOn.TasksCompleted != newEvidence.TasksCompleted {
		t.Errorf("BasedOn.TasksCompleted = %v, want %v", newTransition.BasedOn.TasksCompleted, newEvidence.TasksCompleted)
	}
}

func TestAutonomyTransition_EvidenceMetrics(t *testing.T) {
	transition := AutonomyTransition{
		From:   AutonomyLevelDecide,
		To:     AutonomyLevelPlan,
		Reason: "test",
		BasedOn: CompetenceEvidence{
			TasksCompleted:     20,
			SuccessRate:        0.85,
			ValidationPassRate: 0.92,
			AvgExecutionTime:   100 * time.Millisecond,
		},
	}

	metrics := transition.EvidenceMetrics()

	if metrics["tasks_completed"] != 20 {
		t.Errorf("tasks_completed = %v, want 20", metrics["tasks_completed"])
	}
	if metrics["success_rate"] != 0.85 {
		t.Errorf("success_rate = %v, want 0.85", metrics["success_rate"])
	}
	if metrics["validation_pass_rate"] != 0.92 {
		t.Errorf("validation_pass_rate = %v, want 0.92", metrics["validation_pass_rate"])
	}
	if metrics["avg_execution_time_ms"] != int64(100) {
		t.Errorf("avg_execution_time_ms = %v, want 100", metrics["avg_execution_time_ms"])
	}
}

func TestAutonomyTransition_String(t *testing.T) {
	transition := AutonomyTransition{
		From:   AutonomyLevelDecide,
		To:     AutonomyLevelPlan,
		Reason: "test reason",
	}

	str := transition.String()

	if !strings.Contains(str, "decide") {
		t.Errorf("String should contain 'decide', got %q", str)
	}
	if !strings.Contains(str, "plan") {
		t.Errorf("String should contain 'plan', got %q", str)
	}
	if !strings.Contains(str, "test reason") {
		t.Errorf("String should contain reason, got %q", str)
	}
	if !strings.Contains(str, "delta=1") {
		t.Errorf("String should contain 'delta=1', got %q", str)
	}
}

func TestRuleEngine_WithAuditLogger(t *testing.T) {
	var auditLogs []string
	re := NewRuleEngine(func(format string, args ...interface{}) {
		auditLogs = append(auditLogs, "[AUDIT]")
	})

	evidence := CompetenceEvidence{
		TasksCompleted:     15,
		SuccessRate:        0.85,
		ValidationPassRate: 0.95,
	}

	re.EvaluateTransition(AutonomyLevelExecute, evidence)

	if len(auditLogs) == 0 {
		t.Error("Expected audit logs to be written")
	}
}

func TestRuleEngine_EdgeCases(t *testing.T) {
	re := NewRuleEngine(nil)

	// Exactly at threshold
	atThreshold := CompetenceEvidence{
		TasksCompleted:     10,  // exactly UpgradeMinTasksCompleted
		SuccessRate:        0.8, // exactly UpgradeMinSuccessRate
		ValidationPassRate: 0.9, // exactly UpgradeMinValidationRate
	}

	transition := re.EvaluateTransition(AutonomyLevelExecute, atThreshold)
	if transition.IsNoChange() {
		t.Error("Should upgrade when exactly at thresholds")
	}

	// Just below threshold
	belowThreshold := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.79, // just below
		ValidationPassRate: 0.9,
	}

	transition = re.EvaluateTransition(AutonomyLevelExecute, belowThreshold)
	if !transition.IsNoChange() {
		t.Error("Should not upgrade when success rate is below threshold")
	}
}
