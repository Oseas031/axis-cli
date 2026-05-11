package kv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"
)

// logFileName is the append-only event log.
const logFileName = "history.jsonl"

// appendLog writes a logRecord as a single JSONL line to history.jsonl.
// It returns the byte offset and length of the written line.
// The write is atomic at the line level: either the full line is persisted
// or nothing is. Partial lines are detectable on replay and skipped.
func (e *Engine) appendLog(op OpType, key string, value []byte) (offset int64, length int64, err error) {
	rec := logRecord{
		Op: string(op),
		K:  key,
		Ts: time.Now().UTC(),
	}
	if op == OpPut && value != nil {
		rec.V = json.RawMessage(value)
	}

	// Compact JSON: no unnecessary whitespace. Use json.Encoder with
	// SetEscapeHTML(false) for smart escaping, then strip trailing newline
	// so we can enforce our own LF.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "") // compact; Encoder uses no indent by default
	if err := enc.Encode(rec); err != nil {
		return 0, 0, fmt.Errorf("kv: marshal log record: %w", err)
	}

	// enc.Encode adds a trailing newline. Replace it with a guaranteed LF.
	b := buf.Bytes()
	if len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}
	b = append(b, '\n')

	// Atomic line write: full line in one Write + Sync.
	off, err := e.logFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, 0, fmt.Errorf("kv: seek log: %w", err)
	}
	if _, err := e.logFile.Write(b); err != nil {
		return 0, 0, fmt.Errorf("kv: write log: %w", err)
	}
	if err := e.logFile.Sync(); err != nil {
		return 0, 0, fmt.Errorf("kv: sync log: %w", err)
	}

	return off, int64(len(b)), nil
}

// logPath returns the full path to history.jsonl.
func (e *Engine) logPath() string {
	return filepath.Join(e.rootDir, logFileName)
}
