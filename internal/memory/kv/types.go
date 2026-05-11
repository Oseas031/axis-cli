// Package kv provides a pure Go, standard-library-only append-only key-value
// engine for Axis Working Memory. It uses a three-file model:
//
//	history.jsonl — immutable append-only log
//	snapshot.bin  — compacted checkpoint with tiny header
//	index.txt     — plain-text offset index
package kv

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

// Engine limits (defensive boundaries).
const (
	maxKeyLen    = 256        // bytes
	maxValueLen  = 256 * 1024 // 256 KiB
	maxRecordLen = maxKeyLen + maxValueLen + 1024
)

var (
	// ErrKeyEmpty is returned when the key is empty.
	ErrKeyEmpty = errors.New("kv: key is empty")
	// ErrKeyTooLong is returned when the key exceeds maxKeyLen.
	ErrKeyTooLong = errors.New("kv: key too long")
	// ErrValueTooLong is returned when the value exceeds maxValueLen.
	ErrValueTooLong = errors.New("kv: value too long")
	// ErrNotFound is returned when a key does not exist.
	ErrNotFound = errors.New("kv: key not found")
	// ErrCorrupt is returned when snapshot or index is corrupt.
	ErrCorrupt = errors.New("kv: snapshot/index corrupt")
)

// OpType represents the operation recorded in the log.
type OpType string

const (
	OpPut OpType = "put"
	OpDel OpType = "del"
)

// logRecord is the on-disk JSON shape for history.jsonl.
type logRecord struct {
	Op string          `json:"op"`
	K  string          `json:"k"`
	V  json.RawMessage `json:"v,omitempty"`
	Ts time.Time       `json:"ts"`
}

// RecordPos identifies where a record lives on disk.
type RecordPos struct {
	File   string // "log" or "snapshot"
	Offset int64
	Length int64
}

// Iterator provides sequential access over KV entries.
type Iterator interface {
	// Next advances to the next entry. Returns false when exhausted or on error.
	Next() bool
	// Key returns the current key. Valid only after Next returns true.
	Key() string
	// Value returns the current raw value bytes. Valid only after Next returns true.
	Value() []byte
	// Err returns any iteration error.
	Err() error
	// Close releases the iterator.
	Close() error
}

// Engine defines the IndexedKV interface.
type Engine struct {
	mu      sync.Mutex
	rootDir string
	index   map[string]RecordPos // in-memory authoritative index
	logFile *os.File             // history.jsonl handle
}

func (e *Engine) validateKeyValue(key string, value []byte) error {
	if key == "" {
		return ErrKeyEmpty
	}
	if len(key) > maxKeyLen {
		return ErrKeyTooLong
	}
	if len(value) > maxValueLen {
		return ErrValueTooLong
	}
	return nil
}
