package contextpack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// IndexStore persists and loads TFIDFIndex as JSON.
type IndexStore struct{}

// Save writes the index to path using atomic temp+rename.
func (s *IndexStore) Save(path string, idx *TFIDFIndex) error {
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "index-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		cleanup()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

// Load reads the index from path.
func (s *IndexStore) Load(path string) (*TFIDFIndex, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read index: %w", err)
	}

	var idx TFIDFIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("unmarshal index: %w", err)
	}
	return &idx, nil
}
