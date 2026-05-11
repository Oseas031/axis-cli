package kv

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const indexFileName = "index.txt"

// loadIndex reads index.txt into an in-memory map.
func (e *Engine) loadIndex() (map[string]RecordPos, error) {
	path := filepath.Join(e.rootDir, indexFileName)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("kv: open index: %w", err)
	}
	defer f.Close()

	idx := make(map[string]RecordPos)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 3 {
			continue // skip malformed lines
		}
		key := parts[0]
		off, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}
		length, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			continue
		}
		idx[key] = RecordPos{File: "snapshot", Offset: off, Length: length}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("kv: scan index: %w", err)
	}
	return idx, nil
}
