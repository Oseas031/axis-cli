// Package assemble provides the context assembly algorithm.
// It selects, ranks, and truncates context packets within budget constraints.
package assemble

import (
	"context"

	"github.com/axis-cli/axis/internal/contextpack/model"
)

// Assembler defines the interface for context assembly operations.
type Assembler interface {
	// Assemble selects and ranks context packets for a task goal.
	Assemble(ctx context.Context, goal string, contractID string) (*model.ContextBundle, error)
}

// Ranker defines the interface for packet ranking strategies.
type Ranker interface {
	// Rank orders packets by relevance to the goal.
	Rank(packets []model.ContextPacket, goal string) []model.ContextPacket
}

// Truncater defines the interface for packet truncation.
type Truncater interface {
	// Truncate fits a packet into the given byte budget.
	Truncate(packet model.ContextPacket, budget int) (model.ContextPacket, bool)
}
