// Package contextpack provides context assembly for Axis agent tasks.
//
// Architecture:
//   - model/    — Core data types (ContextPacket, ContextBundle, etc.)
//   - assemble/ — Assembly algorithm (budget-aware packet selection)
//   - index/    — TF-IDF document indexing and retrieval
//   - rank/     — Multi-strategy packet ranking
//   - registry/ — Readiness registration and persistence
//   - consumer/ — Execution context summary generation
//
// This top-level package provides backward-compatible type aliases and
// convenience constructors. New code should import sub-packages directly.
package contextpack

import (
	"github.com/axis-cli/axis/internal/contextpack/model"
)

// Type aliases for backward compatibility.
// New code should use model.ContextPacket directly.
type ContextPacket = model.ContextPacket
type ContextBundle = model.ContextBundle
type AssemblyTrace = model.AssemblyTrace
type TraceItem = model.TraceItem
type ContextBudget = model.ContextBudget
type PacketType = model.PacketType

// Packet type constants.
const (
	PacketTypeSpec      = model.PacketTypeSpec
	PacketTypeCode      = model.PacketTypeCode
	PacketTypeDoc       = model.PacketTypeDoc
	PacketTypePrinciple = model.PacketTypePrinciple
	PacketTypeTool      = model.PacketTypeTool
)

// DefaultBudget returns a default budget configuration.
func DefaultBudget() ContextBudget {
	return model.DefaultBudget()
}
