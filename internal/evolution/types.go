// Package evolution provides data models and storage for the Sandboxed Evolution Protocol.
package evolution

import (
	"time"
)

// EvolutionStatus represents the lifecycle status of an evolution entity.
type EvolutionStatus string

const (
	StatusPending   EvolutionStatus = "pending"
	StatusRunning   EvolutionStatus = "running"
	StatusCompleted EvolutionStatus = "completed"
	StatusFailed    EvolutionStatus = "failed"
	StatusDiscarded EvolutionStatus = "discarded"
	StatusPromoted  EvolutionStatus = "promoted"
	StatusPaused    EvolutionStatus = "paused"
)

// RiskLevel represents the assessed risk of an evolution intent.
type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

// EvolutionIntent captures the initial proposal for a system change.
type EvolutionIntent struct {
	ID           string          `json:"id"`
	CreatedAt    time.Time       `json:"created_at"`
	Actor        string          `json:"actor"`
	Summary      string          `json:"summary"`
	TargetDomain string          `json:"target_domain"`
	RiskLevel    RiskLevel       `json:"risk_level"`
	Status       EvolutionStatus `json:"status"`
}

// EvolutionRun represents a single isolated evolution attempt.
type EvolutionRun struct {
	RunID         string          `json:"run_id"`
	IntentID      string          `json:"intent_id"`
	Status        EvolutionStatus `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
	WorkspacePath string          `json:"workspace_path,omitempty"`
}

// StepAction describes what an evolution step does.
type StepAction string

const (
	StepActionPatch   StepAction = "patch"
	StepActionCreate  StepAction = "create"
	StepActionDelete  StepAction = "delete"
	StepActionVerify  StepAction = "verify"
	StepActionPromote StepAction = "promote"
	StepActionDiscard StepAction = "discard"
)

// EvolutionStep is an atomic, inspectable unit of change within a run.
type EvolutionStep struct {
	StepID      string          `json:"step_id"`
	RunID       string          `json:"run_id"`
	Sequence    int             `json:"sequence"`
	TargetPath  string          `json:"target_path"`
	Action      StepAction      `json:"action"`
	PatchRef    string          `json:"patch_ref,omitempty"`
	Status      EvolutionStatus `json:"status"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	Error       string          `json:"error,omitempty"`
}

// VerificationStatus represents the outcome of a verification command.
type VerificationStatus string

const (
	VerificationPending   VerificationStatus = "pending"
	VerificationPassed    VerificationStatus = "passed"
	VerificationFailed    VerificationStatus = "failed"
	VerificationCancelled VerificationStatus = "cancelled"
)

// VerificationRecord captures the evidence from running a verification command.
type VerificationRecord struct {
	RunID       string             `json:"run_id"`
	Command     string             `json:"command"`
	StartedAt   time.Time          `json:"started_at"`
	CompletedAt *time.Time         `json:"completed_at,omitempty"`
	ExitCode    int                `json:"exit_code"`
	StdoutRef   string             `json:"stdout_ref,omitempty"`
	StderrRef   string             `json:"stderr_ref,omitempty"`
	Status      VerificationStatus `json:"status"`
}

// DecisionType is the explicit outcome of an evolution run.
type DecisionType string

const (
	DecisionPromoted  DecisionType = "promoted"
	DecisionDiscarded DecisionType = "discarded"
	DecisionPaused    DecisionType = "paused"
)

// EvolutionDecision is the final explicit gate for an evolution run.
type EvolutionDecision struct {
	RunID     string       `json:"run_id"`
	Decision  DecisionType `json:"decision"`
	Actor     string       `json:"actor"`
	Reason    string       `json:"reason"`
	CreatedAt time.Time    `json:"created_at"`
}
