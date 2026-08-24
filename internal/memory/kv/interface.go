package kv

import "context"

// Store defines the interface for key-value storage operations.
// This interface enables dependency injection and testing.
type Store interface {
	// Get retrieves the raw value for key.
	Get(ctx context.Context, key string) ([]byte, error)

	// Put writes a key-value pair.
	Put(ctx context.Context, key string, value []byte) error

	// Delete writes a tombstone for key.
	Delete(ctx context.Context, key string) error

	// ScanPrefix returns an iterator over all keys matching prefix.
	ScanPrefix(ctx context.Context, prefix string) (Iterator, error)

	// Compact rebuilds snapshot and index from the in-memory index.
	Compact() error

	// Close flushes and closes the engine.
	Close() error
}

// Ensure Engine implements Store at compile time.
var _ Store = (*Engine)(nil)
