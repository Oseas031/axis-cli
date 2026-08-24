// Package index provides TF-IDF document indexing and retrieval.
package index

import (
	"context"
	"time"
)

// Document represents a document in the index.
type Document struct {
	ID       string
	Source   string
	Content  string
	DocType  string // "code", "doc", "spec"
	ModTime  time.Time
	Checksum string
}

// SearchResult represents a single search result.
type SearchResult struct {
	Chunk *Chunk
	Score float64
}

// Chunk represents a document chunk for indexing.
type Chunk struct {
	Source  string
	Content string
	DocType string
}

// Store defines the interface for index persistence.
type Store interface {
	// Save persists the index to disk.
	Save(ctx context.Context, index *Index) error

	// Load loads the index from disk.
	Load(ctx context.Context) (*Index, error)
}

// Index defines the interface for document indexing operations.
type Index interface {
	// Add adds a document to the index.
	Add(doc *Document) error

	// Query searches the index with the given query.
	Query(query string, limit int) []SearchResult

	// Remove removes a document from the index.
	Remove(source string) error

	// Count returns the number of indexed documents.
	Count() int
}
