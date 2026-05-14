package judgement

import "sync"

// TypeScore tracks pass/fail statistics for a criterion within a single task type.
type TypeScore struct {
	Total    int
	Passed   int
	PassRate float64
}

// GeneralizationScore tracks how well a judgement criterion performs across task types.
type GeneralizationScore struct {
	CriteriaName string
	TaskTypes    map[string]*TypeScore
}

// GeneralizationTracker accumulates judgement results per task type.
type GeneralizationTracker struct {
	scores map[string]*GeneralizationScore
	mu     sync.RWMutex
}

// NewGeneralizationTracker creates a new GeneralizationTracker.
func NewGeneralizationTracker() *GeneralizationTracker {
	return &GeneralizationTracker{scores: make(map[string]*GeneralizationScore)}
}

// Record registers a judgement result for a criterion and task type.
func (gt *GeneralizationTracker) Record(criteriaName, taskType string, passed bool) {
	gt.mu.Lock()
	defer gt.mu.Unlock()

	gs, ok := gt.scores[criteriaName]
	if !ok {
		gs = &GeneralizationScore{CriteriaName: criteriaName, TaskTypes: make(map[string]*TypeScore)}
		gt.scores[criteriaName] = gs
	}
	ts, ok := gs.TaskTypes[taskType]
	if !ok {
		ts = &TypeScore{}
		gs.TaskTypes[taskType] = ts
	}
	ts.Total++
	if passed {
		ts.Passed++
	}
	ts.PassRate = float64(ts.Passed) / float64(ts.Total)
}

// GetScore returns the GeneralizationScore for a criterion, or nil if not found.
func (gt *GeneralizationTracker) GetScore(criteriaName string) *GeneralizationScore {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	return gt.scores[criteriaName]
}

// IsGeneralized returns true if the criterion has been tested on >= minTypes task types
// AND maintains >= minPassRate across all of them.
func (gt *GeneralizationTracker) IsGeneralized(criteriaName string, minTypes int, minPassRate float64) bool {
	gt.mu.RLock()
	defer gt.mu.RUnlock()

	gs, ok := gt.scores[criteriaName]
	if !ok || len(gs.TaskTypes) < minTypes {
		return false
	}
	for _, ts := range gs.TaskTypes {
		if ts.PassRate < minPassRate {
			return false
		}
	}
	return true
}
