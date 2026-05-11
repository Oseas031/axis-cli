package longterm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const eventsFileName = "events.jsonl"

// Store defines the interface for Long-term Memory event operations.
type Store interface {
	Append(ctx context.Context, event EventRecord) error
	QueryEvents(ctx context.Context, filter EventFilter) ([]EventRecord, error)
	MarkForgotten(ctx context.Context, entityID string, at time.Time) error
	Close() error
}

// FileStore is the JSONL append-only implementation.
type FileStore struct {
	mu      sync.Mutex
	rootDir string
	file    *os.File
}

// Open creates or opens a FileStore at rootDir.
func Open(rootDir string) (*FileStore, error) {
	if err := os.MkdirAll(rootDir, 0750); err != nil {
		return nil, fmt.Errorf("longterm: create dir: %w", err)
	}
	path := filepath.Join(rootDir, eventsFileName)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		return nil, fmt.Errorf("longterm: open events file: %w", err)
	}
	return &FileStore{rootDir: rootDir, file: f}, nil
}

// Close closes the event log.
func (s *FileStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

// Append writes an immutable event record.
func (s *FileStore) Append(_ context.Context, event EventRecord) error {
	if event.EventType == "" {
		return ErrEventTypeEmpty
	}
	if event.EntityID == "" {
		return ErrEntityIDEmpty
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(event); err != nil {
		return fmt.Errorf("longterm: marshal: %w", err)
	}

	// Strip trailing newline from encoder, enforce LF.
	b := buf.Bytes()
	if len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}
	b = append(b, '\n')

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.file.Write(b); err != nil {
		return fmt.Errorf("longterm: write: %w", err)
	}
	if err := s.file.Sync(); err != nil {
		return fmt.Errorf("longterm: sync: %w", err)
	}
	return nil
}

// QueryEvents scans the event log with the given filter.
func (s *FileStore) QueryEvents(_ context.Context, filter EventFilter) ([]EventRecord, error) {
	path := filepath.Join(s.rootDir, eventsFileName)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("longterm: open for query: %w", err)
	}
	defer f.Close()

	var results []EventRecord
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("longterm: read: %w", err)
		}
		if len(line) == 0 {
			if err == io.EOF {
				break
			}
			continue
		}
		var rec EventRecord
		if jerr := json.Unmarshal(line, &rec); jerr != nil {
			if err == io.EOF {
				break
			}
			continue // skip malformed
		}
		if matchesFilter(rec, filter) {
			results = append(results, rec)
			if filter.Limit > 0 && len(results) >= filter.Limit {
				break
			}
		}
		if err == io.EOF {
			break
		}
	}
	return results, nil
}

// MarkForgotten appends a forgetting event. Original events are preserved.
func (s *FileStore) MarkForgotten(ctx context.Context, entityID string, at time.Time) error {
	if at.IsZero() {
		at = time.Now().UTC()
	}
	return s.Append(ctx, EventRecord{
		EventType:    EventMemoryForgotten,
		EntityID:     entityID,
		Timestamp:    at,
		DeprecatedAt: &at,
	})
}

func matchesFilter(rec EventRecord, filter EventFilter) bool {
	// Deprecated filtering.
	if !filter.IncludeDeprecated && rec.DeprecatedAt != nil {
		return false
	}
	// Event type filtering.
	if len(filter.EventTypes) > 0 {
		match := false
		for _, t := range filter.EventTypes {
			if rec.EventType == t {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	// Entity ID filtering.
	if filter.EntityID != "" && rec.EntityID != filter.EntityID {
		return false
	}
	// Time range.
	if !filter.After.IsZero() && !rec.Timestamp.After(filter.After) {
		return false
	}
	if !filter.Before.IsZero() && !rec.Timestamp.Before(filter.Before) {
		return false
	}
	return true
}
