package contextpack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileStore persists readiness records as a single JSON file under
// <root>/.axis/context/readiness.json using atomic temp+rename writes.
type FileStore struct {
	root string
	mu   sync.Mutex
}

// NewFileStore creates a FileStore anchored at the given project root.
// If root is empty, it defaults to the current directory (".").
func NewFileStore(root string) *FileStore {
	if root == "" {
		root = "."
	}
	return &FileStore{root: root}
}

// Path returns the absolute path to the readiness JSON file.
func (s *FileStore) Path() string {
	return filepath.Join(s.root, ".axis", "context", "readiness.json")
}

// LoadAll reads the persisted readiness records.
// If the file does not exist, it returns an empty map without error.
func (s *FileStore) LoadAll() (map[string]ReadinessRecord, error) {
	data, err := os.ReadFile(s.Path())
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]ReadinessRecord), nil
		}
		return nil, fmt.Errorf("failed to load readiness store: %w", err)
	}
	var records map[string]ReadinessRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("invalid readiness store: %w", err)
	}
	if records == nil {
		return make(map[string]ReadinessRecord), nil
	}
	return records, nil
}

// SaveAll atomically writes all records to disk using a temp file + rename.
func (s *FileStore) SaveAll(records map[string]ReadinessRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create context store directory: %w", err)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal readiness store: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("failed to write readiness store temp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to commit readiness store: %w", err)
	}
	return nil
}

// DeleteAll removes the persisted store file.
// If the file does not exist, it returns nil.
func (s *FileStore) DeleteAll() error {
	err := os.Remove(s.Path())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
