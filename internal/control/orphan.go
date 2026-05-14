package control

import (
	"bufio"
	"encoding/json"
	"os"
)

// MarkOrphanedTasks scans the event log and appends "abandoned" events
// for any tasks whose last known status is "running" or "pending".
// This handles the case where a previous runtime process died without
// completing its in-flight tasks.
func MarkOrphanedTasks(log *TaskEventLog) (int, error) {
	path := log.Path()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // no event log yet
		}
		return 0, err
	}
	defer f.Close()

	// Build last-known status per task
	lastStatus := make(map[string]string)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 256*1024), 256*1024) // 256KB max line
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var evt struct {
			TaskID string `json:"task_id"`
			Status string `json:"status"`
		}
		if json.Unmarshal(line, &evt) == nil && evt.TaskID != "" && evt.Status != "" {
			lastStatus[evt.TaskID] = evt.Status
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}

	// Mark orphans
	count := 0
	for taskID, status := range lastStatus {
		if status == "running" || status == "pending" {
			if err := log.Append(TaskEvent{
				TaskID:    taskID,
				EventType: "abandoned",
				Status:    "abandoned",
				Actor:     "runtime-startup",
				Message:   "marked abandoned: previous runtime exited without completing this task",
			}); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}
