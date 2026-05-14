package control

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMarkOrphanedTasks_NoFile(t *testing.T) {
	dir := t.TempDir()
	log := NewTaskEventLog(dir)
	count, err := MarkOrphanedTasks(log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 orphans, got %d", count)
	}
}

func TestMarkOrphanedTasks_MarksRunningAndPending(t *testing.T) {
	dir := t.TempDir()
	log := NewTaskEventLog(dir)

	// Simulate events from a previous runtime
	_ = log.Append(TaskEvent{TaskID: "task-1", EventType: "submitted", Status: "pending"})
	_ = log.Append(TaskEvent{TaskID: "task-2", EventType: "submitted", Status: "pending"})
	_ = log.Append(TaskEvent{TaskID: "task-2", EventType: "started", Status: "running"})
	_ = log.Append(TaskEvent{TaskID: "task-3", EventType: "submitted", Status: "pending"})
	_ = log.Append(TaskEvent{TaskID: "task-3", EventType: "completed", Status: "completed"})

	// task-1: last status = pending (orphan)
	// task-2: last status = running (orphan)
	// task-3: last status = completed (not orphan)

	count, err := MarkOrphanedTasks(log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 orphans, got %d", count)
	}

	// Verify abandoned events were appended
	data, _ := os.ReadFile(filepath.Join(dir, ".axis", "events", "tasks.jsonl"))
	lines := 0
	var lastStatuses []string
	for _, line := range splitLines(data) {
		if len(line) == 0 {
			continue
		}
		lines++
		var evt TaskEvent
		_ = json.Unmarshal(line, &evt)
		if evt.EventType == "abandoned" {
			lastStatuses = append(lastStatuses, evt.TaskID)
		}
	}
	if len(lastStatuses) != 2 {
		t.Errorf("expected 2 abandoned events, got %d", len(lastStatuses))
	}
}

func TestMarkOrphanedTasks_Idempotent(t *testing.T) {
	dir := t.TempDir()
	log := NewTaskEventLog(dir)

	_ = log.Append(TaskEvent{TaskID: "task-x", EventType: "submitted", Status: "pending"})

	// First call marks it
	count1, _ := MarkOrphanedTasks(log)
	if count1 != 1 {
		t.Fatalf("first call: expected 1, got %d", count1)
	}

	// Second call should find 0 orphans (task-x is now "abandoned")
	count2, _ := MarkOrphanedTasks(log)
	if count2 != 0 {
		t.Errorf("second call: expected 0 (idempotent), got %d", count2)
	}
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
