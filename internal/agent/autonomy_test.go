// Package agent provides agent autonomy and transition rules.
package agent

import (
	"testing"
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
		// Backward compatibility aliases
		{AutonomyLevelNone, "none"},
		{AutonomyLevelLow, "low"},
		{AutonomyLevelMedium, "medium"},
		{AutonomyLevelHigh, "high"},
		// Unknown
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
		// Primary levels (0-4)
		{AutonomyLevelExecute, true},
		{AutonomyLevelDecide, true},
		{AutonomyLevelPlan, true},
		{AutonomyLevelLearn, true},
		{AutonomyLevelFull, true},
		// Backward compatibility aliases (5-8)
		{AutonomyLevelNone, true},
		{AutonomyLevelLow, true},
		{AutonomyLevelMedium, true},
		{AutonomyLevelHigh, true},
		// Invalid
		{AutonomyLevel(-1), false},
		{AutonomyLevel(9), false},
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
		{AutonomyLevelExecute, AutonomyLevelFull, false},
		// Backward compatibility transitions
		{AutonomyLevelNone, AutonomyLevelLow, true},
		{AutonomyLevelLow, AutonomyLevelMedium, true},
		// Invalid (same values but different constants)
		{AutonomyLevelExecute, AutonomyLevelNone, false},
	}

	for _, tt := range tests {
		t.Run(tt.from.String()+"_to_"+tt.to.String(), func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.expected {
				t.Errorf("CanTransitionTo() = %v, want %v", got, tt.expected)
			}
		})
	}
}
