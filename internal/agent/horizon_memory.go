package agent

import (
	"fmt"
	"time"

	"github.com/axis-cli/axis/internal/memory/horizon"
)

// HorizonMemory adapts horizon.Store to the ExecutionMemory interface.
// Stores lessons as patterns; recalls by keyword search.
type HorizonMemory struct {
	store *horizon.Store
}

func NewHorizonMemory(store *horizon.Store) *HorizonMemory {
	return &HorizonMemory{store: store}
}

func (m *HorizonMemory) StoreLessson(taskID, lesson string) error {
	return m.store.Store(horizon.Entry{
		ID:        fmt.Sprintf("lesson-%s-%d", taskID, time.Now().UnixMilli()),
		Category:  horizon.CategoryPatterns,
		Title:     "Execution lesson: " + taskID,
		Tags:      []string{"lesson", "auto"},
		CreatedAt: time.Now(),
		Body:      lesson,
	})
}

func (m *HorizonMemory) RecallLessons(query string) []string {
	entries, err := m.store.Recall(query, horizon.CategoryPatterns)
	if err != nil || len(entries) == 0 {
		return nil
	}
	// Return at most 3 most recent lessons
	max := 3
	if len(entries) < max {
		max = len(entries)
	}
	lessons := make([]string, max)
	for i := 0; i < max; i++ {
		lessons[i] = entries[len(entries)-1-i].Body
	}
	return lessons
}
