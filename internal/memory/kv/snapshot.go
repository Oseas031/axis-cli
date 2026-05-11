package kv

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	snapshotFileName = "snapshot.bin"
	magic            = "AXSN"
	headerVersion    = 1
	headerSize       = 64
)

// snapshotHeader is the 64-byte fixed header at the start of snapshot.bin.
type snapshotHeader struct {
	Magic                  [4]byte
	Version                uint32
	RecordCount            uint64
	DataOffset             uint64
	CreatedAtUnixNano      uint64
	CompactedHistoryOffset uint64
	_                      [24]byte // reserved
}

func newSnapshotHeader(recordCount, dataOffset, createdAt, compactedOffset uint64) snapshotHeader {
	h := snapshotHeader{
		Version:                headerVersion,
		RecordCount:            recordCount,
		DataOffset:             dataOffset,
		CreatedAtUnixNano:      createdAt,
		CompactedHistoryOffset: compactedOffset,
	}
	copy(h.Magic[:], magic)
	return h
}

func (h *snapshotHeader) encode() []byte {
	b := make([]byte, headerSize)
	copy(b[0:4], h.Magic[:])
	binary.LittleEndian.PutUint32(b[4:8], h.Version)
	binary.LittleEndian.PutUint64(b[8:16], h.RecordCount)
	binary.LittleEndian.PutUint64(b[16:24], h.DataOffset)
	binary.LittleEndian.PutUint64(b[24:32], h.CreatedAtUnixNano)
	binary.LittleEndian.PutUint64(b[32:40], h.CompactedHistoryOffset)
	return b
}

func decodeSnapshotHeader(b []byte) (snapshotHeader, error) {
	if len(b) < headerSize {
		return snapshotHeader{}, fmt.Errorf("kv: header too short (%d < %d)", len(b), headerSize)
	}
	h := snapshotHeader{}
	copy(h.Magic[:], b[0:4])
	if string(h.Magic[:]) != magic {
		return snapshotHeader{}, fmt.Errorf("kv: invalid magic %q", h.Magic)
	}
	h.Version = binary.LittleEndian.Uint32(b[4:8])
	if h.Version != headerVersion {
		return snapshotHeader{}, fmt.Errorf("kv: unsupported header version %d", h.Version)
	}
	h.RecordCount = binary.LittleEndian.Uint64(b[8:16])
	h.DataOffset = binary.LittleEndian.Uint64(b[16:24])
	h.CreatedAtUnixNano = binary.LittleEndian.Uint64(b[24:32])
	h.CompactedHistoryOffset = binary.LittleEndian.Uint64(b[32:40])
	return h, nil
}

// snapshotRecord is the JSON shape inside snapshot.bin (no op/ts).
type snapshotRecord struct {
	K string          `json:"k"`
	V json.RawMessage `json:"v,omitempty"`
}

// writeSnapshot builds a new snapshot.bin and index.txt from the authoritative
// in-memory index. It returns the number of records written.
func (e *Engine) writeSnapshot(historyOffset uint64) (recordCount int, err error) {
	tmpSnapshot := filepath.Join(e.rootDir, "."+snapshotFileName+".tmp")
	tmpIndex := filepath.Join(e.rootDir, "."+indexFileName+".tmp")

	// Gather and sort keys for deterministic output.
	keys := make([]string, 0, len(e.index))
	for k, pos := range e.index {
		if pos.File == "log" || pos.File == "snapshot" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	snapFile, err := os.Create(tmpSnapshot)
	if err != nil {
		return 0, fmt.Errorf("kv: create temp snapshot: %w", err)
	}
	defer snapFile.Close()

	// Write placeholder header; we will overwrite it after we know dataOffset.
	if _, err := snapFile.Write(make([]byte, headerSize)); err != nil {
		return 0, fmt.Errorf("kv: write snapshot placeholder: %w", err)
	}

	idxFile, err := os.Create(tmpIndex)
	if err != nil {
		return 0, fmt.Errorf("kv: create temp index: %w", err)
	}
	defer idxFile.Close()

	enc := json.NewEncoder(snapFile)
	enc.SetEscapeHTML(false)

	var written int
	for _, key := range keys {
		pos := e.index[key]
		val, err := e.readAtPos(pos)
		if err != nil {
			return 0, fmt.Errorf("kv: read value for key %q: %w", key, err)
		}
		rec := snapshotRecord{K: key, V: json.RawMessage(val)}

		off, err := snapFile.Seek(0, io.SeekCurrent)
		if err != nil {
			return 0, fmt.Errorf("kv: seek snapshot: %w", err)
		}

		if err := enc.Encode(rec); err != nil {
			return 0, fmt.Errorf("kv: encode snapshot record: %w", err)
		}

		// enc.Encode writes a trailing newline. Determine exact length.
		after, err := snapFile.Seek(0, io.SeekCurrent)
		if err != nil {
			return 0, fmt.Errorf("kv: seek snapshot after encode: %w", err)
		}
		length := after - off

		// Write index line: key offset length
		if _, err := fmt.Fprintf(idxFile, "%s %d %d\n", key, off, length); err != nil {
			return 0, fmt.Errorf("kv: write index line: %w", err)
		}
		written++
	}

	// Determine dataOffset (position after header).
	dataOff, err := snapFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, fmt.Errorf("kv: seek snapshot end: %w", err)
	}

	// Overwrite header.
	h := newSnapshotHeader(uint64(written), uint64(dataOff), uint64(time.Now().UnixNano()), historyOffset)
	if _, err := snapFile.WriteAt(h.encode(), 0); err != nil {
		return 0, fmt.Errorf("kv: write snapshot header: %w", err)
	}

	if err := snapFile.Sync(); err != nil {
		return 0, fmt.Errorf("kv: sync snapshot: %w", err)
	}
	if err := idxFile.Close(); err != nil {
		return 0, fmt.Errorf("kv: close index: %w", err)
	}
	if err := snapFile.Close(); err != nil {
		return 0, fmt.Errorf("kv: close snapshot: %w", err)
	}

	// Atomic replacement (cross-platform safe).
	if err := e.atomicReplace(tmpSnapshot, filepath.Join(e.rootDir, snapshotFileName)); err != nil {
		return 0, fmt.Errorf("kv: replace snapshot: %w", err)
	}
	if err := e.atomicReplace(tmpIndex, filepath.Join(e.rootDir, indexFileName)); err != nil {
		return 0, fmt.Errorf("kv: replace index: %w", err)
	}

	return written, nil
}

// readSnapshotHeader reads and validates the snapshot header.
func (e *Engine) readSnapshotHeader() (snapshotHeader, error) {
	path := filepath.Join(e.rootDir, snapshotFileName)
	f, err := os.Open(path)
	if err != nil {
		return snapshotHeader{}, err
	}
	defer f.Close()

	b := make([]byte, headerSize)
	if _, err := io.ReadFull(f, b); err != nil {
		return snapshotHeader{}, fmt.Errorf("kv: read header: %w", err)
	}
	return decodeSnapshotHeader(b)
}

// scanSnapshot sequentially reads snapshot.bin JSONL records and invokes
// cb for each valid record. Used for index rebuild fallback.
func (e *Engine) scanSnapshot(cb func(key string, offset, length int64) bool) error {
	path := filepath.Join(e.rootDir, snapshotFileName)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Skip header.
	if _, err := f.Seek(int64(headerSize), io.SeekStart); err != nil {
		return fmt.Errorf("kv: seek past header: %w", err)
	}

	dec := json.NewDecoder(f)
	for {
		// Record offset before decode.
		off, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			return fmt.Errorf("kv: seek snapshot: %w", err)
		}

		var rec snapshotRecord
		if err := dec.Decode(&rec); err != nil {
			if err == io.EOF {
				break
			}
			// Skip corrupt line; try to continue.
			continue
		}

		after, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			return fmt.Errorf("kv: seek after decode: %w", err)
		}
		length := after - off

		if !cb(rec.K, off, length) {
			break
		}
	}
	return nil
}
