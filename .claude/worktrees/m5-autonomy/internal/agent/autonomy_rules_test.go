package agent

import (
	"testing"
	"time"
)

func TestAutonomyLevel_Constants(t *testing.T) {
	if AutonomyLevelNone != 0 {
		t.Errorf("AutonomyLevelNone: expected 0, got %d", AutonomyLevelNone)
	}
	if AutonomyLevelLow != 1 {
		t.Errorf("AutonomyLevelLow: expected 1, got %d", AutonomyLevelLow)
	}
	if AutonomyLevelMedium != 2 {
		t.Errorf("AutonomyLevelMedium: expected 2, got %d", AutonomyLevelMedium)
	}
	if AutonomyLevelHigh != 3 {
		t.Errorf("AutonomyLevelHigh: expected 3, got %d", AutonomyLevelHigh)
	}
	if AutonomyLevelFull != 4 {
		t.Errorf("AutonomyLevelFull: expected 4, got %d", AutonomyLevelFull)
	}
}

func TestAutonomyLevel_String(t *testing.T) {
	tests := []struct {
		level    AutonomyLevel
		expected string
	}{
		{AutonomyLevelNone, "none"},
		{AutonomyLevelLow, "low"},
		{AutonomyLevelMedium, "medium"},
		{AutonomyLevelHigh, "high"},
		{AutonomyLevelFull, "full"},
		{AutonomyLevel(5), "unknown"},
		{AutonomyLevel(-1), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("AutonomyLevel.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAutonomyLevel_IsValid(t *testing.T) {
	tests := []struct {
		level    AutonomyLevel
		expected bool
	}{
		{AutonomyLevelNone, true},
		{AutonomyLevelLow, true},
		{AutonomyLevelMedium, true},
		{AutonomyLevelHigh, true},
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
		// Valid adjacent transitions
		{AutonomyLevelNone, AutonomyLevelLow, true},
		{AutonomyLevelLow, AutonomyLevelMedium, true},
		{AutonomyLevelMedium, AutonomyLevelHigh, true},
		{AutonomyLevelHigh, AutonomyLevelFull, true},
		// Invalid non-adjacent transitions
		{AutonomyLevelNone, AutonomyLevelMedium, false},
		{AutonomyLevelNone, AutonomyLevelHigh, false},
		{AutonomyLevelNone, AutonomyLevelFull, false},
		{AutonomyLevelLow, AutonomyLevelHigh, false},
		{AutonomyLevelLow, AutonomyLevelFull, false},
		{AutonomyLevelMedium, AutonomyLevelFull, false},
		// Same level (no change)
		{AutonomyLevelLow, AutonomyLevelLow, true},
		// Downgrade transitions
		{AutonomyLevelFull, AutonomyLevelHigh, true},
		{AutonomyLevelHigh, AutonomyLevelMedium, true},
		// Invalid out of bounds
		{AutonomyLevelNone, AutonomyLevel(-1), false},
		{AutonomyLevelFull, AutonomyLevel(5), false},
	}

	for _, tt := range tests {
		t.Run(tt.from.String()+"_to_"+tt.to.String(), func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.expected {
				t.Errorf("CanTransitionTo(%v, %v) = %v, want %v", tt.from, tt.to, got, tt.expected)
			}
		})
	}
}

func TestAutonomyLevel_IsUpgrade(t *testing.T) {
	if !AutonomyLevelLow.IsUpgrade(AutonomyLevelMedium) {
		t.Error("Low should be upgrade to Medium")
	}
	if !AutonomyLevelNone.IsUpgrade(AutonomyLevelFull) {
		t.Error("None should be upgrade to Full")
	}
	if AutonomyLevelMedium.IsUpgrade(AutonomyLevelLow) {
		t.Error("Medium should not be upgrade to Low")
	}
	if AutonomyLevelHigh.IsUpgrade(AutonomyLevelHigh) {
		t.Error("Same level is not upgrade")
	}
}

func TestAutonomyLevel_IsDowngrade(t *testing.T) {
	if !AutonomyLevelMedium.IsDowngrade(AutonomyLevelLow) {
		t.Error("Medium should be downgrade to Low")
	}
	if !AutonomyLevelFull.IsDowngrade(AutonomyLevelNone) {
		t.Error("Full should be downgrade to None")
	}
	if AutonomyLevelLow.IsDowngrade(AutonomyLevelMedium) {
		t.Error("Low should not be downgrade to Medium")
	}
	if AutonomyLevelHigh.IsDowngrade(AutonomyLevelHigh) {
		t.Error("Same level is not downgrade")
	}
}

func TestNewCompetenceEvidence(t *testing.T) {
	ce := NewCompetenceEvidence()

	if ce.TasksCompleted != 0 {
		t.Errorf("TasksCompleted: expected 0, got %d", ce.TasksCompleted)
	}
	if ce.SuccessRate != 0.0 {
		t.Errorf("SuccessRate: expected 0.0, got %f", ce.SuccessRate)
	}
	if ce.ValidationPassRate != 0.0 {
		t.Errorf("ValidationPassRate: expected 0.0, got %f", ce.ValidationPassRate)
	}
	if ce.AvgExecutionTime != 0 {
		t.Errorf("AvgExecutionTime: expected 0, got %v", ce.AvgExecutionTime)
	}
}

func TestCompetenceEvidence_IsComplete(t *testing.T) {
	tests := []struct {
		tasks    int
		expected bool
	}{
		{0, false},
		{1, true},
		{10, true},
	}

	for _, tt := range tests {
		ce := CompetenceEvidence{TasksCompleted: tt.tasks}
		if got := ce.IsComplete(); got != tt.expected {
			t.Errorf("IsComplete() with tasks=%d = %v, want %v", tt.tasks, got, tt.expected)
		}
	}
}

func TestCompetenceEvidence_Clone(t *testing.T) {
	ce := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.85,
		ValidationPassRate: 0.95,
		AvgExecutionTime:   100 * time.Millisecond,
	}

	cloned := ce.Clone()

	if cloned.TasksCompleted != ce.TasksCompleted {
		t.Errorf("Clone TasksCompleted = %d, want %d", cloned.TasksCompleted, ce.TasksCompleted)
	}
	if cloned.SuccessRate != ce.SuccessRate {
		t.Errorf("Clone SuccessRate = %f, want %f", cloned.SuccessRate, ce.SuccessRate)
	}
	if cloned.ValidationPassRate != ce.ValidationPassRate {
		t.Errorf("Clone ValidationPassRate = %f, want %f", cloned.ValidationPassRate, ce.ValidationPassRate)
	}
	if cloned.AvgExecutionTime != ce.AvgExecutionTime {
		t.Errorf("Clone AvgExecutionTime = %v, want %v", cloned.AvgExecutionTime, ce.AvgExecutionTime)
	}

	// Verify it's a deep copy
	cloned.TasksCompleted = 999
	if ce.TasksCompleted == 999 {
		t.Error("Clone should be independent from original")
	}
}

func TestCompetenceEvidence_MeetsUpgradeThresholds(t *testing.T) {
	tests := []struct {
		name     string
		tasks    int
		success  float64
		valid    float64
		expected bool
	}{
		{"all thresholds met", 10, 0.81, 0.91, true},
		{"exactly at tasks", 10, 0.81, 0.91, true},
		{"below tasks", 9, 0.81, 0.91, false},
		{"at success boundary", 10, 0.8, 0.91, false},    // must be > 0.8
		{"at validation boundary", 10, 0.81, 0.9, false}, // must be > 0.9
		{"all boundaries", 10, 0.8, 0.9, false},
		{"zero values", 0, 0.0, 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce := CompetenceEvidence{
				TasksCompleted:     tt.tasks,
				SuccessRate:        tt.success,
				ValidationPassRate: tt.valid,
			}
			if got := ce.MeetsUpgradeThresholds(); got != tt.expected {
				t.Errorf("MeetsUpgradeThresholds() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCompetenceEvidence_IndicatesDowngrade(t *testing.T) {
	tests := []struct {
		name     string
		tasks    int
		success  float64
		expected bool
	}{
		{"below 50% with 10 tasks", 10, 0.49, true},
		{"at 50% boundary", 10, 0.5, false}, // must be < 0.5
		{"above 50%", 10, 0.51, false},
		{"below 50% with insufficient tasks", 9, 0.49, false},
		{"zero tasks", 0, 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce := CompetenceEvidence{
				TasksCompleted: tt.tasks,
				SuccessRate:    tt.success,
			}
			if got := ce.IndicatesDowngrade(); got != tt.expected {
				t.Errorf("IndicatesDowngrade() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAutonomyTransition_IsUpgrade(t *testing.T) {
	transition := AutonomyTransition{From: AutonomyLevelLow, To: AutonomyLevelMedium}
	if !transition.IsUpgrade() {
		t.Error("Low->Medium should be upgrade")
	}

	transition.To = AutonomyLevelLow
	if transition.IsUpgrade() {
		t.Error("Low->Low should not be upgrade")
	}

	transition.To = AutonomyLevelNone
	if transition.IsUpgrade() {
		t.Error("Low->None should not be upgrade (it's downgrade)")
	}
}

func TestAutonomyTransition_IsDowngrade(t *testing.T) {
	transition := AutonomyTransition{From: AutonomyLevelMedium, To: AutonomyLevelLow}
	if !transition.IsDowngrade() {
		t.Error("Medium->Low should be downgrade")
	}

	transition.To = AutonomyLevelMedium
	if transition.IsDowngrade() {
		t.Error("Medium->Medium should not be downgrade")
	}

	transition.To = AutonomyLevelFull
	if transition.IsDowngrade() {
		t.Error("Medium->Full should not be downgrade (it's upgrade)")
	}
}

func TestAutonomyTransition_IsNoChange(t *testing.T) {
	transition := AutonomyTransition{From: AutonomyLevelLow, To: AutonomyLevelLow}
	if !transition.IsNoChange() {
		t.Error("Low->Low should be no change")
	}

	transition.To = AutonomyLevelMedium
	if transition.IsNoChange() {
		t.Error("Low->Medium should not be no change")
	}
}

func TestAutonomyTransition_IsSignificant(t *testing.T) {
	transition := AutonomyTransition{From: AutonomyLevelLow, To: AutonomyLevelLow}
	if transition.IsSignificant() {
		t.Error("Low->Low should not be significant")
	}

	transition.To = AutonomyLevelMedium
	if !transition.IsSignificant() {
		t.Error("Low->Medium should be significant")
	}
}

func TestAutonomyTransition_GetDelta(t *testing.T) {
	transition := AutonomyTransition{From: AutonomyLevelLow, To: AutonomyLevelMedium}
	if transition.GetDelta() != 1 {
		t.Errorf("Low->Medium delta = %d, want 1", transition.GetDelta())
	}

	transition.To = AutonomyLevelNone
	if transition.GetDelta() != -1 {
		t.Errorf("Low->None delta = %d, want -1", transition.GetDelta())
	}

	transition.To = AutonomyLevelLow
	if transition.GetDelta() != 0 {
		t.Errorf("Low->Low delta = %d, want 0", transition.GetDelta())
	}
}

func TestAutonomyTransition_WithEvidence(t *testing.T) {
	transition := AutonomyTransition{
		From:   AutonomyLevelLow,
		To:     AutonomyLevelMedium,
		Reason: "test",
	}

	newEvidence := CompetenceEvidence{
		TasksCompleted:     15,
		SuccessRate:        0.9,
		ValidationPassRate: 0.95,
	}

	newTransition := transition.WithEvidence(newEvidence)

	if newTransition.From != transition.From {
		t.Error("WithEvidence should preserve From")
	}
	if newTransition.To != transition.To {
		t.Error("WithEvidence should preserve To")
	}
	if newTransition.BasedOn != newEvidence {
		t.Error("WithEvidence should update BasedOn")
	}
}

func TestAutonomyTransition_EvidenceMetrics(t *testing.T) {
	transition := AutonomyTransition{
		BasedOn: CompetenceEvidence{
			TasksCompleted:     10,
			SuccessRate:        0.85,
			ValidationPassRate: 0.92,
			AvgExecutionTime:   100 * time.Millisecond,
		},
	}

	metrics := transition.EvidenceMetrics()

	if metrics["tasks_completed"].(int) != 10 {
		t.Errorf("tasks_completed = %v, want 10", metrics["tasks_completed"])
	}
	if metrics["success_rate"].(float64) != 0.85 {
		t.Errorf("success_rate = %v, want 0.85", metrics["success_rate"])
	}
	if metrics["validation_pass_rate"].(float64) != 0.92 {
		t.Errorf("validation_pass_rate = %v, want 0.92", metrics["validation_pass_rate"])
	}
	if metrics["avg_execution_time_ms"].(int64) != 100 {
		t.Errorf("avg_execution_time_ms = %v, want 100", metrics["avg_execution_time_ms"])
	}
}

func TestAutonomyTransition_String(t *testing.T) {
	transition := AutonomyTransition{
		From:   AutonomyLevelLow,
		To:     AutonomyLevelMedium,
		Reason: "test reason",
	}

	s := transition.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
	// Just verify it doesn't panic and contains key info
	_ = s
}

func TestNewAutonomyRecord(t *testing.T) {
	evidence := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.85,
		ValidationPassRate: 0.92,
	}

	record := NewAutonomyRecord(AutonomyLevelMedium, evidence)

	if record.Level != AutonomyLevelMedium {
		t.Errorf("Level = %v, want %v", record.Level, AutonomyLevelMedium)
	}
	if record.Evidence != evidence {
		t.Error("Evidence not set correctly")
	}
	if record.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestAutonomyRecord_Clone(t *testing.T) {
	record := NewAutonomyRecord(AutonomyLevelHigh, CompetenceEvidence{
		TasksCompleted: 10,
		SuccessRate:    0.9,
	})

	cloned := record.Clone()

	if cloned.Level != record.Level {
		t.Error("Clone should preserve Level")
	}

	// Verify it's a deep copy
	cloned.Level = AutonomyLevelNone
	if record.Level == AutonomyLevelNone {
		t.Error("Clone should be independent from original")
	}
}

func TestAutonomyRecord_CloneNil(t *testing.T) {
	var record *AutonomyRecord
	cloned := record.Clone()

	if cloned != nil {
		t.Error("Clone of nil should return nil")
	}
}

func TestAutonomyRecord_IsStale(t *testing.T) {
	record := &AutonomyRecord{
		Level:     AutonomyLevelMedium,
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}

	if !record.IsStale(1 * time.Hour) {
		t.Error("Record from 24h ago should be stale with 1h threshold")
	}

	if record.IsStale(48 * time.Hour) {
		t.Error("Record from 24h ago should not be stale with 48h threshold")
	}
}

func TestAutonomyRecord_IsStaleNil(t *testing.T) {
	var record *AutonomyRecord

	if !record.IsStale(time.Hour) {
		t.Error("Nil record should be considered stale")
	}
}

func TestAutonomyRecord_Update(t *testing.T) {
	record := NewAutonomyRecord(AutonomyLevelLow, CompetenceEvidence{TasksCompleted: 5})

	newEvidence := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.85,
		ValidationPassRate: 0.92,
	}

	oldUpdatedAt := record.UpdatedAt
	time.Sleep(time.Millisecond) // Ensure time advances
	record.Update(AutonomyLevelMedium, newEvidence)

	if record.Level != AutonomyLevelMedium {
		t.Errorf("Level = %v, want %v", record.Level, AutonomyLevelMedium)
	}
	if record.Evidence != newEvidence {
		t.Error("Evidence not updated correctly")
	}
	if !record.UpdatedAt.After(oldUpdatedAt) {
		t.Error("UpdatedAt should be updated to a newer time")
	}
}

// RuleEngine tests

func TestNewRuleEngine(t *testing.T) {
	engine := NewRuleEngine(nil)
	if engine == nil {
		t.Fatal("NewRuleEngine should not return nil")
	}

	// Test with custom logger - verify it doesn't panic
	customLoggerCalled := false
	customLogger := func(format string, args ...interface{}) {
		customLoggerCalled = true
	}
	engine = NewRuleEngine(customLogger)
	// Trigger a transition to ensure logger is called
	evidence := CompetenceEvidence{TasksCompleted: 10, SuccessRate: 0.85, ValidationPassRate: 0.95}
	engine.EvaluateTransition(AutonomyLevelLow, evidence)
	if !customLoggerCalled {
		t.Error("Custom logger should have been called")
	}
}

func TestRuleEngine_EvaluateTransition_NoChange_InsufficientTasks(t *testing.T) {
	engine := NewRuleEngine(nil)
	evidence := CompetenceEvidence{
		TasksCompleted:     5, // Below minimum 10
		SuccessRate:        0.9,
		ValidationPassRate: 0.95,
	}

	transition := engine.EvaluateTransition(AutonomyLevelLow, evidence)

	if !transition.IsNoChange() {
		t.Error("Should not transition with insufficient tasks")
	}
}

func TestRuleEngine_EvaluateTransition_NoChange_InsufficientSuccessRate(t *testing.T) {
	engine := NewRuleEngine(nil)
	evidence := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.8, // Not > 0.8
		ValidationPassRate: 0.95,
	}

	transition := engine.EvaluateTransition(AutonomyLevelLow, evidence)

	if !transition.IsNoChange() {
		t.Error("Should not transition with insufficient success rate")
	}
}

func TestRuleEngine_EvaluateTransition_NoChange_InsufficientValidationRate(t *testing.T) {
	engine := NewRuleEngine(nil)
	evidence := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.85,
		ValidationPassRate: 0.9, // Not > 0.9
	}

	transition := engine.EvaluateTransition(AutonomyLevelLow, evidence)

	if !transition.IsNoChange() {
		t.Error("Should not transition with insufficient validation rate")
	}
}

func TestRuleEngine_EvaluateTransition_Upgrade(t *testing.T) {
	engine := NewRuleEngine(nil)
	evidence := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.85,
		ValidationPassRate: 0.95,
	}

	transition := engine.EvaluateTransition(AutonomyLevelLow, evidence)

	if !transition.IsUpgrade() {
		t.Error("Should upgrade with sufficient evidence")
	}
	if transition.To != AutonomyLevelMedium {
		t.Errorf("Should upgrade to Medium, got %v", transition.To)
	}
}

func TestRuleEngine_EvaluateTransition_UpgradeChain(t *testing.T) {
	engine := NewRuleEngine(nil)
	evidence := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.85,
		ValidationPassRate: 0.95,
	}

	// Test full upgrade chain
	levels := []AutonomyLevel{
		AutonomyLevelNone,
		AutonomyLevelLow,
		AutonomyLevelMedium,
		AutonomyLevelHigh,
	}

	for i := 0; i < len(levels)-1; i++ {
		transition := engine.EvaluateTransition(levels[i], evidence)
		if !transition.IsUpgrade() {
			t.Errorf("Should upgrade from %v", levels[i])
		}
		if transition.To != levels[i+1] {
			t.Errorf("Should upgrade to %v, got %v", levels[i+1], transition.To)
		}
	}
}

func TestRuleEngine_EvaluateTransition_NoUpgradeFromFull(t *testing.T) {
	engine := NewRuleEngine(nil)
	evidence := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.85,
		ValidationPassRate: 0.95,
	}

	transition := engine.EvaluateTransition(AutonomyLevelFull, evidence)

	if transition.To != AutonomyLevelFull {
		t.Error("Should not upgrade from Full")
	}
}

func TestRuleEngine_EvaluateTransition_Downgrade(t *testing.T) {
	engine := NewRuleEngine(nil)
	evidence := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.4, // Below 0.5
		ValidationPassRate: 0.95,
	}

	transition := engine.EvaluateTransition(AutonomyLevelMedium, evidence)

	if !transition.IsDowngrade() {
		t.Error("Should downgrade with low success rate")
	}
	if transition.To != AutonomyLevelLow {
		t.Errorf("Should downgrade to Low, got %v", transition.To)
	}
}

func TestRuleEngine_EvaluateTransition_DowngradeChain(t *testing.T) {
	engine := NewRuleEngine(nil)
	evidence := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.4,
		ValidationPassRate: 0.95,
	}

	// Test full downgrade chain
	levels := []AutonomyLevel{
		AutonomyLevelFull,
		AutonomyLevelHigh,
		AutonomyLevelMedium,
		AutonomyLevelLow,
		AutonomyLevelNone,
	}

	for i := 0; i < len(levels)-1; i++ {
		transition := engine.EvaluateTransition(levels[i], evidence)
		if !transition.IsDowngrade() {
			t.Errorf("Should downgrade from %v", levels[i])
		}
		if transition.To != levels[i+1] {
			t.Errorf("Should downgrade to %v, got %v", levels[i+1], transition.To)
		}
	}
}

func TestRuleEngine_EvaluateTransition_NoDowngradeFromNone(t *testing.T) {
	engine := NewRuleEngine(nil)
	evidence := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.4,
		ValidationPassRate: 0.95,
	}

	transition := engine.EvaluateTransition(AutonomyLevelNone, evidence)

	if transition.To != AutonomyLevelNone {
		t.Error("Should not downgrade from None")
	}
}

func TestRuleEngine_EvaluateTransition_NoDowngrade_InsufficientTasks(t *testing.T) {
	engine := NewRuleEngine(nil)
	evidence := CompetenceEvidence{
		TasksCompleted:     5, // Below 10
		SuccessRate:        0.4,
		ValidationPassRate: 0.95,
	}

	transition := engine.EvaluateTransition(AutonomyLevelMedium, evidence)

	// Should not downgrade without minimum tasks
	if transition.IsDowngrade() {
		t.Error("Should not downgrade with insufficient tasks even with low success rate")
	}
}

func TestRuleEngine_EvaluateTransition_DowngradePriority(t *testing.T) {
	// Downgrade should take priority over upgrade
	engine := NewRuleEngine(nil)
	evidence := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.4,  // Low - triggers downgrade
		ValidationPassRate: 0.95, // High - would trigger upgrade
	}

	transition := engine.EvaluateTransition(AutonomyLevelHigh, evidence)

	if !transition.IsDowngrade() {
		t.Error("Downgrade should take priority over upgrade")
	}
}

func TestRuleEngine_ValidateTransition(t *testing.T) {
	engine := NewRuleEngine(nil)

	tests := []struct {
		from     AutonomyLevel
		to       AutonomyLevel
		hasError bool
	}{
		{AutonomyLevelLow, AutonomyLevelMedium, false},
		{AutonomyLevelLow, AutonomyLevelHigh, true},    // Non-adjacent
		{AutonomyLevelLow, AutonomyLevel(-1), true},    // Invalid target
		{AutonomyLevel(-1), AutonomyLevelMedium, true}, // Invalid source
	}

	for _, tt := range tests {
		t.Run(tt.from.String()+"_to_"+tt.to.String(), func(t *testing.T) {
			err := engine.ValidateTransition(tt.from, tt.to)
			if tt.hasError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.hasError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestRuleEngine_GetRequiredEvidence(t *testing.T) {
	engine := NewRuleEngine(nil)

	required := engine.GetRequiredEvidence(AutonomyLevelLow)

	if required["min_tasks_completed"].(int) != UpgradeMinTasksCompleted {
		t.Errorf("min_tasks_completed = %v, want %d", required["min_tasks_completed"], UpgradeMinTasksCompleted)
	}
	if required["min_success_rate"].(float64) != UpgradeMinSuccessRate {
		t.Errorf("min_success_rate = %v, want %f", required["min_success_rate"], UpgradeMinSuccessRate)
	}
	if required["min_validation_pass_rate"].(float64) != UpgradeMinValidationRate {
		t.Errorf("min_validation_pass_rate = %v, want %f", required["min_validation_pass_rate"], UpgradeMinValidationRate)
	}
}

func TestRuleEngine_RecordTransition(t *testing.T) {
	engine := NewRuleEngine(nil)
	transition := AutonomyTransition{
		From:   AutonomyLevelLow,
		To:     AutonomyLevelMedium,
		Reason: "test",
		BasedOn: CompetenceEvidence{
			TasksCompleted:     10,
			SuccessRate:        0.85,
			ValidationPassRate: 0.95,
		},
	}

	record := engine.RecordTransition(transition)

	if record.Level != transition.To {
		t.Errorf("Level = %v, want %v", record.Level, transition.To)
	}
}

func TestRuleEngine_CanUpgrade(t *testing.T) {
	engine := NewRuleEngine(nil)
	evidence := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.85,
		ValidationPassRate: 0.95,
	}

	if !engine.CanUpgrade(AutonomyLevelLow, evidence) {
		t.Error("Should be able to upgrade with sufficient evidence")
	}

	if engine.CanUpgrade(AutonomyLevelFull, evidence) {
		t.Error("Should not be able to upgrade from Full")
	}

	// Insufficient evidence
	evidence.SuccessRate = 0.5
	if engine.CanUpgrade(AutonomyLevelLow, evidence) {
		t.Error("Should not upgrade with insufficient evidence")
	}
}

func TestRuleEngine_CanDowngrade(t *testing.T) {
	engine := NewRuleEngine(nil)
	evidence := CompetenceEvidence{
		TasksCompleted:     10,
		SuccessRate:        0.4,
		ValidationPassRate: 0.95,
	}

	if !engine.CanDowngrade(AutonomyLevelMedium, evidence) {
		t.Error("Should be able to downgrade with low success rate")
	}

	if engine.CanDowngrade(AutonomyLevelNone, evidence) {
		t.Error("Should not be able to downgrade from None")
	}

	// Insufficient tasks
	evidence.TasksCompleted = 5
	if engine.CanDowngrade(AutonomyLevelMedium, evidence) {
		t.Error("Should not downgrade with insufficient tasks")
	}

	// High success rate
	evidence.SuccessRate = 0.6
	evidence.TasksCompleted = 10
	if engine.CanDowngrade(AutonomyLevelMedium, evidence) {
		t.Error("Should not downgrade with high success rate")
	}
}

func TestCalculateAutonomyDelta(t *testing.T) {
	tests := []struct {
		from     AutonomyLevel
		to       AutonomyLevel
		expected int
	}{
		{AutonomyLevelLow, AutonomyLevelMedium, 1},
		{AutonomyLevelMedium, AutonomyLevelLow, -1},
		{AutonomyLevelNone, AutonomyLevelFull, 4},
		{AutonomyLevelFull, AutonomyLevelNone, -4},
		{AutonomyLevelMedium, AutonomyLevelMedium, 0},
	}

	for _, tt := range tests {
		t.Run(tt.from.String()+"_"+tt.to.String(), func(t *testing.T) {
			if got := CalculateAutonomyDelta(tt.from, tt.to); got != tt.expected {
				t.Errorf("CalculateAutonomyDelta(%v, %v) = %d, want %d", tt.from, tt.to, got, tt.expected)
			}
		})
	}
}

func TestLevelName(t *testing.T) {
	// Helper function for logging
	tests := []struct {
		level    AutonomyLevel
		expected string
	}{
		{AutonomyLevelNone, "none"},
		{AutonomyLevelLow, "low"},
		{AutonomyLevelMedium, "medium"},
		{AutonomyLevelHigh, "high"},
		{AutonomyLevelFull, "full"},
	}

	for _, tt := range tests {
		if got := levelName(tt.level); got != tt.expected {
			t.Errorf("levelName(%v) = %q, want %q", tt.level, got, tt.expected)
		}
	}
}
