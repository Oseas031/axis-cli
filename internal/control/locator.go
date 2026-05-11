package control

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var ErrRuntimeLocatorNotFound = errors.New("runtime locator not found")

type RuntimeRecord struct {
	PID         int       `json:"pid"`
	Protocol    string    `json:"protocol"`
	Address     string    `json:"address"`
	StartedAt   time.Time `json:"started_at"`
	ProjectRoot string    `json:"project_root"`
}

type RuntimeLocator struct {
	root string
}

func NewRuntimeLocator(root string) *RuntimeLocator {
	if root == "" {
		root = "."
	}
	return &RuntimeLocator{root: root}
}

func (l *RuntimeLocator) Path() string {
	return filepath.Join(l.root, ".axis", "runtime.json")
}

func (l *RuntimeLocator) Save(record RuntimeRecord) error {
	path := l.Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (l *RuntimeLocator) Load() (RuntimeRecord, error) {
	data, err := os.ReadFile(l.Path())
	if err != nil {
		if os.IsNotExist(err) {
			return RuntimeRecord{}, ErrRuntimeLocatorNotFound
		}
		return RuntimeRecord{}, err
	}
	var record RuntimeRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return RuntimeRecord{}, fmt.Errorf("invalid runtime locator: %w", err)
	}
	return record, nil
}

func (l *RuntimeLocator) Delete() error {
	err := os.Remove(l.Path())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
