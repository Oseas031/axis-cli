// Package registry provides readiness registration for context bundles.
package registry

import (
	"context"
	"time"
)

// ReadinessStatus represents the readiness state of a bundle.
type ReadinessStatus string

const (
	StatusPending  ReadinessStatus = "pending"
	StatusReady    ReadinessStatus = "ready"
	StatusStale    ReadinessStatus = "stale"
	StatusNotFound ReadinessStatus = "not_found"
)

// ReadinessRecord represents the readiness state of a context bundle.
type ReadinessRecord struct {
	BundleID     string            `json:"bundle_id"`
	Status       ReadinessStatus   `json:"status"`
	RegisteredAt time.Time         `json:"registered_at"`
	ExpiresAt    *time.Time        `json:"expires_at,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Store defines the interface for readiness persistence.
type Store interface {
	// LoadAll loads all readiness records.
	LoadAll(ctx context.Context) (map[string]ReadinessRecord, error)

	// SaveAll persists all readiness records.
	SaveAll(ctx context.Context, records map[string]ReadinessRecord) error

	// DeleteAll removes all persisted records.
	DeleteAll(ctx context.Context) error
}

// Registry defines the interface for readiness registration operations.
type Registry interface {
	// Register marks a bundle as ready.
	Register(ctx context.Context, bundleID string, metadata map[string]string) error

	// IsReady checks if a bundle is ready.
	IsReady(ctx context.Context, bundleID string) bool

	// GetStatus returns the readiness status of a bundle.
	GetStatus(ctx context.Context, bundleID string) ReadinessStatus

	// List returns all registered bundles.
	List(ctx context.Context) ([]ReadinessRecord, error)

	// Unregister removes a bundle's readiness record.
	Unregister(ctx context.Context, bundleID string) error
}
