// Package consumer provides execution context summary generation.
package consumer

import (
	"context"

	"github.com/axis-cli/axis/internal/contextpack/model"
)

// SourceType represents the type of context source.
type SourceType string

const (
	SourceRules     SourceType = "rules"
	SourceRetrieval SourceType = "retrieval"
	SourceMemory    SourceType = "memory"
	SourceSkill     SourceType = "skill"
)

// RequestedSource represents a context source requested by the task.
type RequestedSource struct {
	Type    SourceType `json:"type"`
	Pattern string     `json:"pattern"`
	Reason  string     `json:"reason"`
}

// ExecutionContextSummary summarizes the execution context for a task.
type ExecutionContextSummary struct {
	Goal             string                `json:"goal"`
	ContractID       string                `json:"contract_id"`
	RequestedSources []RequestedSource     `json:"requested_sources"`
	AvailablePackets []model.ContextPacket `json:"available_packets"`
	Budget           model.ContextBudget   `json:"budget"`
}

// Consumer defines the interface for execution context generation.
type Consumer interface {
	// Summarize generates an execution context summary for a task.
	Summarize(ctx context.Context, taskID string, goal string, contractID string) (*ExecutionContextSummary, error)

	// GetRequestedSources extracts requested sources from task metadata.
	GetRequestedSources(ctx context.Context, metadata map[string]string) []RequestedSource
}
