package control

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type TaskEvent struct {
	EventID   string    `json:"event_id"`
	TaskID    string    `json:"task_id"`
	EventType string    `json:"event_type"`
	CreatedAt time.Time `json:"created_at"`
	Actor     string    `json:"actor,omitempty"`
	Status    string    `json:"status,omitempty"`
	Message   string    `json:"message,omitempty"`
}

type TaskEventLog struct {
	root string
	mu   sync.Mutex
}

func NewTaskEventLog(root string) *TaskEventLog {
	if root == "" {
		root = "."
	}
	return &TaskEventLog{root: root}
}

func (l *TaskEventLog) Path() string {
	return filepath.Join(l.root, ".axis", "events", "tasks.jsonl")
}

func (l *TaskEventLog) Append(event TaskEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	if event.EventID == "" {
		event.EventID = fmt.Sprintf("evt-%d", event.CreatedAt.UnixNano())
	}
	path := l.Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}
