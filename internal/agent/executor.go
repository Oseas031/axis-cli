// Package agent provides agent execution capabilities.
package agent

import (
	"context"

	"github.com/axis-cli/axis/internal/types"
)

// AgentExecutor is the interface for agent-based task execution.
// It handles autonomous task execution with self-context and autonomy levels.
type AgentExecutor interface {
	Execute(ctx context.Context, req *AgentExecutionRequest) (*AgentExecutionResult, error)
	GetAutonomyLevel() AutonomyLevel
}

// AgentExecutionRequest contains all context needed for agent execution.
type AgentExecutionRequest struct {
	Task        *types.AgentTask
	SelfContext *SelfContext
	Contract    *types.AgentContract
	Autonomy    AutonomyLevel
}

// AgentExecutionResult is the output of an agent execution.
type AgentExecutionResult struct {
	Output           map[string]any
	FollowUpTasks    []*types.AgentTask
	ValidationResult *ValidationSummary
	AutonomyDelta    AutonomyDelta
	Error            string
}

// AutonomyDelta represents a change in autonomy level.
type AutonomyDelta struct {
	Delta  int // positive for earned, negative for lost
	Reason string
}

// ValidationSummary contains test and coverage results from execution.
type ValidationSummary struct {
	TestsPassed  int
	TestsFailed  int
	Coverage     float64
	IsAcceptable bool
}
