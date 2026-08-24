// Package storeutil provides shared storage utilities for all store implementations.
// It eliminates code duplication for JSONL append, atomic writes, and file operations.
package storeutil

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// JSONLOptions configures JSONL write behavior.
type JSONLOptions struct {
	Sync       bool // Call fsync after write
	EscapeHTML bool // Set EscapeHTML on encoder
}

// DefaultJSONLOptions returns default options (sync enabled).
func DefaultJSONLOptions() JSONLOptions {
	return JSONLOptions{
		Sync:       true,
		EscapeHTML: false,
	}
}

// AppendJSONL appends a JSON record to a JSONL file.
// It creates the file and parent directories if they don't exist.
func AppendJSONL(path string, record any, opts JSONLOptions) error {
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("storeutil: marshal: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("storeutil: mkdir: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return fmt.Errorf("storeutil: open: %w", err)
	}
	defer file.Close()

	// Add trailing newline
	if len(data) == 0 || data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("storeutil: write: %w", err)
	}

	if opts.Sync {
		if err := file.Sync(); err != nil {
			return fmt.Errorf("storeutil: sync: %w", err)
		}
	}

	return nil
}

// ReadJSONL reads all records from a JSONL file.
// It returns the decoded records as a slice.
func ReadJSONL[T any](path string) ([]T, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("storeutil: open: %w", err)
	}
	defer file.Close()

	var results []T
	r := bufio.NewReader(file)

	for {
		line, err := r.ReadBytes('\n')
		if len(line) == 0 {
			if err == io.EOF {
				break
			}
			continue
		}

		// Strip trailing newline
		line = bytes.TrimSuffix(line, []byte("\n"))

		var record T
		if jerr := json.Unmarshal(line, &record); jerr != nil {
			// Skip malformed lines
			if err == io.EOF {
				break
			}
			continue
		}
		results = append(results, record)

		if err == io.EOF {
			break
		}
	}

	return results, nil
}

// ReadJSONLWithCallback reads a JSONL file and calls a callback for each record.
// This is memory-efficient for large files.
func ReadJSONLWithCallback[T any](path string, fn func(T) bool) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("storeutil: open: %w", err)
	}
	defer file.Close()

	r := bufio.NewReader(file)

	for {
		line, err := r.ReadBytes('\n')
		if len(line) == 0 {
			if err == io.EOF {
				break
			}
			continue
		}

		line = bytes.TrimSuffix(line, []byte("\n"))

		var record T
		if jerr := json.Unmarshal(line, &record); jerr != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		if !fn(record) {
			break
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

// AtomicWriteJSON atomically writes a JSON value to a file.
// It writes to a temp file first, then renames.
func AtomicWriteJSON(path string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("storeutil: marshal: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("storeutil: mkdir: %w", err)
	}

	// Write to temp file
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0640); err != nil {
		return fmt.Errorf("storeutil: write tmp: %w", err)
	}

	// Atomic rename
	if err := AtomicReplace(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}

	return nil
}

// AtomicReplace performs a cross-platform safe atomic file replacement.
// On Unix it uses a single os.Rename. On Windows it uses a two-phase rename.
func AtomicReplace(src, dst string) error {
	// Try direct rename first (works on Unix, and on Windows if dst doesn't exist)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// Fallback for Windows when dst exists
	old := dst + ".old"
	os.Remove(old)

	if err := os.Rename(dst, old); err != nil {
		return fmt.Errorf("storeutil: rename old: %w", err)
	}

	if err := os.Rename(src, dst); err != nil {
		// Attempt rollback
		os.Rename(old, dst)
		return fmt.Errorf("storeutil: rename new: %w", err)
	}

	os.Remove(old)
	return nil
}

// EnsureDir creates a directory if it doesn't exist.
func EnsureDir(dir string, perm os.FileMode) error {
	return os.MkdirAll(dir, perm)
}

// TimestampedID generates an ID based on current timestamp.
func TimestampedID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// Mutex is a named mutex for store-level locking.
type Mutex struct {
	mu sync.Mutex
}

// NewMutex creates a new named mutex.
func NewMutex() *Mutex {
	return &Mutex{}
}

// Lock acquires the mutex.
func (m *Mutex) Lock() {
	m.mu.Lock()
}

// Unlock releases the mutex.
func (m *Mutex) Unlock() {
	m.mu.Unlock()
}

// RWMutex is a named read-write mutex for store-level locking.
type RWMutex struct {
	mu sync.RWMutex
}

// NewRWMutex creates a new named read-write mutex.
func NewRWMutex() *RWMutex {
	return &RWMutex{}
}

// RLock acquires a read lock.
func (m *RWMutex) RLock() {
	m.mu.RLock()
}

// RUnlock releases a read lock.
func (m *RWMutex) RUnlock() {
	m.mu.RUnlock()
}

// Lock acquires a write lock.
func (m *RWMutex) Lock() {
	m.mu.Lock()
}

// Unlock releases a write lock.
func (m *RWMutex) Unlock() {
	m.mu.Unlock()
}
