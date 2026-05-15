package vigil

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LockInfo represents an active lock on a vigil item.
type LockInfo struct {
	Holder    string    `json:"holder"`
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"started_at"`
}

// Locker manages file-based locks for vigil items.
type Locker struct {
	dir string
}

// NewLocker creates a Locker that stores locks in the given vigil directory.
func NewLocker(vigilDir string) *Locker {
	return &Locker{dir: filepath.Join(vigilDir, "locks")}
}

func (l *Locker) lockPath(id string) string {
	return filepath.Join(l.dir, id+".lock")
}

// Lock attempts to acquire a lock for the given item ID.
// Returns an error if the item is already locked by a live process.
func (l *Locker) Lock(id, holder string) error {
	if id == "" {
		return errors.New("id is empty")
	}
	if err := os.MkdirAll(l.dir, 0o755); err != nil {
		return err
	}
	info, err := l.Read(id)
	if err == nil && info != nil {
		if processAlive(info.PID) {
			return fmt.Errorf("item %s locked by %s (PID %d) since %s",
				id, info.Holder, info.PID, info.StartedAt.Format(time.RFC3339))
		}
		// Stale lock — reclaim
	}
	lock := LockInfo{
		Holder:    holder,
		PID:       os.Getpid(),
		StartedAt: time.Now(),
	}
	data, err := json.Marshal(lock)
	if err != nil {
		return err
	}
	return os.WriteFile(l.lockPath(id), data, 0o644)
}

// Unlock removes the lock for the given item ID.
func (l *Locker) Unlock(id string) error {
	err := os.Remove(l.lockPath(id))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// Read returns the lock info for an item, or nil if not locked.
func (l *Locker) Read(id string) (*LockInfo, error) {
	data, err := os.ReadFile(l.lockPath(id))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var info LockInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// IsLocked returns true if the item is locked by a live process.
func (l *Locker) IsLocked(id string) bool {
	info, err := l.Read(id)
	if err != nil || info == nil {
		return false
	}
	return processAlive(info.PID)
}
