package kv

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Open creates or opens a KV engine at rootDir.
// It replays history.jsonl (optionally from a snapshot checkpoint) to build
// the in-memory authoritative index.
func Open(rootDir string) (*Engine, error) {
	if err := os.MkdirAll(rootDir, 0750); err != nil {
		return nil, fmt.Errorf("kv: create root dir: %w", err)
	}

	e := &Engine{
		rootDir: rootDir,
		index:   make(map[string]RecordPos),
	}

	logPath := e.logPath()
	logExists := true
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		return nil, fmt.Errorf("kv: open log: %w", err)
	}
	e.logFile = f

	// Determine history size for replay boundary.
	logSize, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("kv: seek log end: %w", err)
	}

	var replayStart int64

	// Attempt to load snapshot + index.
	snapPath := filepath.Join(rootDir, snapshotFileName)
	if _, err := os.Stat(snapPath); err == nil {
		h, err := e.readSnapshotHeader()
		if err == nil {
			// Header valid. Try loading index.
			idx, err := e.loadIndex()
			if err == nil && idx != nil {
				// Validate that every offset points inside snapshot.bin.
				valid := true
				snapInfo, _ := os.Stat(snapPath)
				for _, pos := range idx {
					if pos.File != "snapshot" {
						valid = false
						break
					}
					if snapInfo != nil && pos.Offset+pos.Length > snapInfo.Size() {
						valid = false
						break
					}
				}
				if valid {
					e.index = idx
					replayStart = int64(h.CompactedHistoryOffset)
				} else {
					// Index corrupt; try rebuilding from snapshot scan.
					rebuilt := make(map[string]RecordPos)
					scanErr := e.scanSnapshot(func(key string, offset, length int64) bool {
						rebuilt[key] = RecordPos{File: "snapshot", Offset: offset, Length: length}
						return true
					})
					if scanErr == nil {
						e.index = rebuilt
						replayStart = int64(h.CompactedHistoryOffset)
					} // else fall through to full replay
				}
			} else if err == nil && idx == nil {
				// index.txt missing; rebuild from snapshot scan.
				rebuilt := make(map[string]RecordPos)
				scanErr := e.scanSnapshot(func(key string, offset, length int64) bool {
					rebuilt[key] = RecordPos{File: "snapshot", Offset: offset, Length: length}
					return true
				})
				if scanErr == nil {
					e.index = rebuilt
					replayStart = int64(h.CompactedHistoryOffset)
				}
			}
		} else {
			// Header invalid; discard snapshot entirely.
			_ = os.Remove(snapPath)
			_ = os.Remove(filepath.Join(rootDir, indexFileName))
		}
	}

	// Replay history.jsonl from replayStart to logSize.
	// We use bufio.Reader + ReadBytes instead of bufio.Scanner so that we can
	// track the exact byte offset of each line (Scanner buffers ahead and its
	// position tracking is unreliable).
	if logExists && logSize > replayStart {
		if _, err := f.Seek(replayStart, io.SeekStart); err != nil {
			return nil, fmt.Errorf("kv: seek log for replay: %w", err)
		}
		r := bufio.NewReaderSize(f, maxRecordLen+1024)
		var off int64 = replayStart
		for {
			line, err := r.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, fmt.Errorf("kv: read log line: %w", err)
			}
			if len(line) == 0 {
				continue
			}
			// line includes trailing '\n' (except possibly last line without newline)
			lineLen := int64(len(line))
			var rec logRecord
			if err := json.Unmarshal(line, &rec); err != nil {
				// Skip malformed lines.
				off += lineLen
				continue
			}
			switch rec.Op {
			case string(OpPut):
				e.index[rec.K] = RecordPos{File: "log", Offset: off, Length: lineLen}
			case string(OpDel):
				delete(e.index, rec.K)
			}
			off += lineLen
		}
	}

	return e, nil
}

// Close flushes and closes the log file.
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.logFile != nil {
		return e.logFile.Close()
	}
	return nil
}

// Get retrieves the raw value for key.
func (e *Engine) Get(_ context.Context, key string) ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	pos, ok := e.index[key]
	if !ok {
		return nil, ErrNotFound
	}
	return e.readAtPos(pos)
}

// readAtPos reads the raw record bytes at the given position.
func (e *Engine) readAtPos(pos RecordPos) ([]byte, error) {
	var path string
	if pos.File == "snapshot" {
		path = filepath.Join(e.rootDir, snapshotFileName)
	} else {
		path = e.logPath()
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("kv: open %s: %w", pos.File, err)
	}
	defer f.Close()

	if _, err := f.Seek(pos.Offset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("kv: seek %s: %w", pos.File, err)
	}
	b := make([]byte, pos.Length)
	if _, err := io.ReadFull(f, b); err != nil {
		return nil, fmt.Errorf("kv: read %s: %w", pos.File, err)
	}

	if pos.File == "log" {
		// Parse the log line and extract just the value.
		var rec logRecord
		if err := json.Unmarshal(b, &rec); err != nil {
			return nil, fmt.Errorf("kv: unmarshal log record: %w", err)
		}
		if rec.Op == string(OpDel) {
			return nil, ErrNotFound
		}
		return []byte(rec.V), nil
	}

	// Snapshot line: parse and extract value.
	var rec snapshotRecord
	if err := json.Unmarshal(b, &rec); err != nil {
		return nil, fmt.Errorf("kv: unmarshal snapshot record: %w", err)
	}
	return []byte(rec.V), nil
}

// Put writes a key-value pair. The value must be valid JSON (or any byte slice
// that the caller considers a value); it is stored as-is inside the JSON envelope.
func (e *Engine) Put(_ context.Context, key string, value []byte) error {
	if err := e.validateKeyValue(key, value); err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	off, length, err := e.appendLog(OpPut, key, value)
	if err != nil {
		return err
	}
	e.index[key] = RecordPos{File: "log", Offset: off, Length: length}
	return nil
}

// Delete writes a tombstone for key.
func (e *Engine) Delete(_ context.Context, key string) error {
	if key == "" {
		return ErrKeyEmpty
	}
	if len(key) > maxKeyLen {
		return ErrKeyTooLong
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	_, _, err := e.appendLog(OpDel, key, nil)
	if err != nil {
		return err
	}
	delete(e.index, key)
	return nil
}

// scanIterator implements Iterator over the in-memory index.
type scanIterator struct {
	keys   []string
	values [][]byte
	pos    int
	err    error
}

func (it *scanIterator) Next() bool {
	if it.pos >= len(it.keys) {
		return false
	}
	it.pos++
	return it.pos <= len(it.keys)
}

func (it *scanIterator) Key() string {
	if it.pos == 0 || it.pos > len(it.keys) {
		return ""
	}
	return it.keys[it.pos-1]
}

func (it *scanIterator) Value() []byte {
	if it.pos == 0 || it.pos > len(it.values) {
		return nil
	}
	return it.values[it.pos-1]
}

func (it *scanIterator) Err() error { return it.err }

func (it *scanIterator) Close() error {
	it.keys = nil
	it.values = nil
	return nil
}

// ScanPrefix returns an iterator over all keys matching prefix.
func (e *Engine) ScanPrefix(_ context.Context, prefix string) (Iterator, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var keys []string
	for k := range e.index {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}

	// Pre-load values for iterator simplicity (P0 data volumes are small).
	values := make([][]byte, 0, len(keys))
	for _, k := range keys {
		pos := e.index[k]
		val, err := e.readAtPos(pos)
		if err != nil {
			return nil, fmt.Errorf("kv: read key %q: %w", k, err)
		}
		values = append(values, val)
	}

	return &scanIterator{keys: keys, values: values}, nil
}

// Compact rebuilds snapshot.bin and index.txt from the authoritative
// in-memory index without modifying history.jsonl.
func (e *Engine) Compact() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Record current history offset.
	historyOffset, err := e.logFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("kv: seek log for compact: %w", err)
	}

	_, err = e.writeSnapshot(uint64(historyOffset))
	if err != nil {
		return fmt.Errorf("kv: compact: %w", err)
	}
	return nil
}

// atomicReplace performs a cross-platform safe atomic file replacement.
// On Unix it uses a single os.Rename. On Windows it uses a two-phase
// rename (old → .old, new → target, remove .old) to handle the
// "cannot replace existing file" limitation.
func (e *Engine) atomicReplace(src, dst string) error {
	// On all platforms, try a direct rename first (works on Unix, and on
	// Windows if dst does not exist).
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// Fallback for Windows when dst exists.
	old := dst + ".old"
	_ = os.Remove(old) // clean up any stale .old
	if err := os.Rename(dst, old); err != nil {
		return fmt.Errorf("kv: rename old %q to %q: %w", dst, old, err)
	}
	if err := os.Rename(src, dst); err != nil {
		// Attempt rollback.
		_ = os.Rename(old, dst)
		return fmt.Errorf("kv: rename new %q to %q: %w", src, dst, err)
	}
	_ = os.Remove(old)
	return nil
}
