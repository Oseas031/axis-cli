// Package working provides the Working Memory layer for Axis.
// It manages retained context bundles across a task chain or session,
// backed by the kv.IndexedKV engine.
package working

import (
	"context"
	"errors"
	"time"

	"github.com/axis-cli/axis/internal/contextpack/model"
)

var (
	// ErrBundleIDEmpty is returned when bundle_id is empty.
	ErrBundleIDEmpty = errors.New("working: bundle_id is empty")
	// ErrBundleNotFound is returned when a bundle is not in the working set.
	ErrBundleNotFound = errors.New("working: bundle not found")
	// ErrReasonEmpty is returned when retain reason is empty.
	ErrReasonEmpty = errors.New("working: retain reason is empty")
)

// bundleKeyPrefix is the mandatory key prefix for all Working Memory entries.
const bundleKeyPrefix = "wm:bundle:"

// Memory defines the interface for Working Memory operations.
type Memory interface {
	// Retain adds a bundle to the working set with a reason.
	Retain(ctx context.Context, bundleID string, reason string) error

	// Release removes a bundle from the working set.
	Release(ctx context.Context, bundleID string) error

	// Recall retrieves relevant packets from retained bundles by keyword.
	Recall(ctx context.Context, query string, limit int) ([]PacketHit, error)

	// List returns all retained bundles in the working set.
	List(ctx context.Context) ([]WorkingSetItem, error)

	// Clear empties the entire working set.
	Clear(ctx context.Context) error

	// Compact triggers explicit snapshot rebuild.
	Compact() error
}

// WorkingSetItem represents a retained bundle in the working set.
type WorkingSetItem struct {
	BundleID    string    `json:"bundle_id"`
	RetainedAt  time.Time `json:"retained_at"`
	Reason      string    `json:"reason"`
	AccessCount int       `json:"access_count"`
}

// PacketHit represents a single packet retrieved by Recall.
type PacketHit struct {
	BundleID  string           `json:"bundle_id"`
	PacketID  string           `json:"packet_id"`
	Type      model.PacketType `json:"type"`
	Source    string           `json:"source"`
	Summary   string           `json:"summary"`
	Relevance float64          `json:"relevance"`
}

// WorkingBundle is the self-describing JSON value stored in the KV engine.
// It uses model types to avoid duplication.
type WorkingBundle struct {
	BundleID     string                `json:"bundle_id"`
	TaskID       string                `json:"task_id"`
	ContractID   string                `json:"contract_id"`
	Goal         string                `json:"goal"`
	Packets      []model.ContextPacket `json:"packets"`
	Trace        model.AssemblyTrace   `json:"trace"`
	Budget       model.ContextBudget   `json:"budget"`
	RetainedAt   time.Time             `json:"retained_at"`
	AccessCount  int                   `json:"access_count"`
	Reason       string                `json:"reason"`
	SourceDigest string                `json:"source_digest"`
}

// ContextPacket is an alias for model.ContextPacket for backward compatibility.
type ContextPacket = model.ContextPacket

// AssemblyTrace is an alias for model.AssemblyTrace for backward compatibility.
type AssemblyTrace = model.AssemblyTrace

// TraceItem is an alias for model.TraceItem for backward compatibility.
type TraceItem = model.TraceItem

// ContextBudget is an alias for model.ContextBudget for backward compatibility.
type ContextBudget = model.ContextBudget

// makeBundleKey returns the canonical KV key for a bundle.
func makeBundleKey(bundleID string) string {
	return bundleKeyPrefix + bundleID
}
