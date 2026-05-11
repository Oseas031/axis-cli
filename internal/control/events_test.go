package control

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTaskEventLogAppendWritesJSONL(t *testing.T) {
	root := t.TempDir()
	log := NewTaskEventLog(root)
	event := TaskEvent{
		EventID:   "evt-1",
		TaskID:    "task-1",
		EventType: "submitted",
		Actor:     "local-control",
		Status:    "pending",
		Message:   "task submitted",
		CreatedAt: time.Date(2026, 5, 10, 10, 0, 0, 0, time.UTC),
	}
	if err := log.Append(event); err != nil {
		t.Fatalf("append event: %v", err)
	}
	data, err := os.ReadFile(log.Path())
	if err != nil {
		t.Fatalf("read event log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected one jsonl event, got %d: %q", len(lines), string(data))
	}
	var got TaskEvent
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("event line should be json: %v", err)
	}
	if got.EventID != event.EventID || got.TaskID != event.TaskID || got.EventType != event.EventType || got.Status != event.Status {
		t.Fatalf("unexpected event: %#v", got)
	}
}

func TestTaskEventLogAppendAddsCreatedAtAndEventID(t *testing.T) {
	log := NewTaskEventLog(t.TempDir())
	if err := log.Append(TaskEvent{TaskID: "task-1", EventType: "status_requested"}); err != nil {
		t.Fatalf("append event: %v", err)
	}
	data, err := os.ReadFile(log.Path())
	if err != nil {
		t.Fatalf("read event log: %v", err)
	}
	var got TaskEvent
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(data))), &got); err != nil {
		t.Fatalf("event line should be json: %v", err)
	}
	if got.EventID == "" {
		t.Fatal("expected generated event id")
	}
	if got.CreatedAt.IsZero() {
		t.Fatal("expected generated created_at")
	}
}

func TestTaskEventLogDoesNotWriteSecrets(t *testing.T) {
	log := NewTaskEventLog(t.TempDir())
	if err := log.Append(TaskEvent{TaskID: "task-1", EventType: "submitted", Message: "ok"}); err != nil {
		t.Fatalf("append event: %v", err)
	}
	data, err := os.ReadFile(log.Path())
	if err != nil {
		t.Fatalf("read event log: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(data))), &got); err != nil {
		t.Fatalf("event line should be json: %v", err)
	}
	for _, key := range []string{"api_key", "apikey", "token", "secret", "authorization", "password"} {
		if _, ok := got[key]; ok {
			t.Fatalf("event log should not contain secret key %q: %#v", key, got)
		}
	}
}

func TestTaskEventLogAppendConcurrentWritesValidJSONL(t *testing.T) {
	log := NewTaskEventLog(t.TempDir())
	var wg sync.WaitGroup
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := log.Append(TaskEvent{TaskID: "task-1", EventType: "status_requested"}); err != nil {
				t.Errorf("append event: %v", err)
			}
		}()
	}
	wg.Wait()
	data, err := os.ReadFile(log.Path())
	if err != nil {
		t.Fatalf("read event log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 25 {
		t.Fatalf("expected 25 jsonl events, got %d", len(lines))
	}
	for _, line := range lines {
		var got TaskEvent
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Fatalf("event line should be json: %v; line=%q", err, line)
		}
	}
}
