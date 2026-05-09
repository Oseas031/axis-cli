// Package agent provides self-context management for agent autonomy.
package agent

import (
	"github.com/axis-cli/axis/internal/types"
)

// AutonomyLevel represents the agent's current autonomy level.
type AutonomyLevel int

const (
	AutonomyLevelExecute AutonomyLevel = iota // Can execute tasks autonomously
	AutonomyLevelDecide                       // Can decide on task approach
	AutonomyLevelPlan                         // Can plan and create tasks
	AutonomyLevelLearn                        // Can learn and improve
)

// SelfContext represents the complete context of an agent at a point in time.
type SelfContext struct {
	TaskID          string         `json:"task_id"`
	TaskLineage     []string       `json:"task_lineage"` // parent task IDs
	CodeSnapshot    *CodeSnapshot  `json:"code_snapshot"`
	DocSnapshot     *DocSnapshot   `json:"doc_snapshot"`
	StateSnapshot   *StateSnapshot `json:"state_snapshot"`
	AutonomyLevel   AutonomyLevel  `json:"autonomy_level"`
	CompetenceScore float64        `json:"competence_score"`
}

// CodeSnapshot captures the current state of the codebase.
type CodeSnapshot struct {
	ModifiedFiles []string `json:"modified_files"`
	SpecVersion   string   `json:"spec_version"`
	TaskCount     int      `json:"task_count"`
	ToolCount     int      `json:"tool_count"`
}

// DocSnapshot captures the current state of documentation.
type DocSnapshot struct {
	SpecFiles []string `json:"spec_files"`
	DocFiles  []string `json:"doc_files"`
	Reports   []string `json:"reports"`
}

// StateSnapshot captures the current task execution state.
type StateSnapshot struct {
	RunningTasks   int `json:"running_tasks"`
	PendingTasks   int `json:"pending_tasks"`
	CompletedTasks int `json:"completed_tasks"`
	FailedTasks    int `json:"failed_tasks"`
}

// SchedulerProvider interface for accessing scheduler operations.
type SchedulerProvider interface {
	GetAllTasks() []*types.AgentTask
}

// StateStoreProvider interface for accessing state store operations.
type StateStoreProvider interface {
	Load(taskID string) (types.TaskState, error)
}

// NewSelfContext creates a new SelfContext with default values.
func NewSelfContext(taskID string) *SelfContext {
	return &SelfContext{
		TaskID:      taskID,
		TaskLineage: make([]string, 0),
		CodeSnapshot: &CodeSnapshot{
			ModifiedFiles: make([]string, 0),
		},
		DocSnapshot: &DocSnapshot{
			SpecFiles: make([]string, 0),
			DocFiles:  make([]string, 0),
			Reports:   make([]string, 0),
		},
		StateSnapshot:   &StateSnapshot{},
		AutonomyLevel:   AutonomyLevelExecute,
		CompetenceScore: 0.5,
	}
}

// NewCodeSnapshot creates a new CodeSnapshot with default values.
func NewCodeSnapshot() *CodeSnapshot {
	return &CodeSnapshot{
		ModifiedFiles: make([]string, 0),
	}
}

// NewDocSnapshot creates a new DocSnapshot with default values.
func NewDocSnapshot() *DocSnapshot {
	return &DocSnapshot{
		SpecFiles: make([]string, 0),
		DocFiles:  make([]string, 0),
		Reports:   make([]string, 0),
	}
}

// NewStateSnapshot creates a new StateSnapshot with default values.
func NewStateSnapshot() *StateSnapshot {
	return &StateSnapshot{}
}

// Clone creates a deep copy of SelfContext.
func (sc *SelfContext) Clone() *SelfContext {
	if sc == nil {
		return nil
	}
	clone := &SelfContext{
		TaskID:          sc.TaskID,
		AutonomyLevel:   sc.AutonomyLevel,
		CompetenceScore: sc.CompetenceScore,
	}
	if sc.TaskLineage != nil {
		clone.TaskLineage = make([]string, len(sc.TaskLineage))
		copy(clone.TaskLineage, sc.TaskLineage)
	}
	if sc.CodeSnapshot != nil {
		clone.CodeSnapshot = &CodeSnapshot{
			SpecVersion: sc.CodeSnapshot.SpecVersion,
			TaskCount:   sc.CodeSnapshot.TaskCount,
			ToolCount:   sc.CodeSnapshot.ToolCount,
		}
		if sc.CodeSnapshot.ModifiedFiles != nil {
			clone.CodeSnapshot.ModifiedFiles = make([]string, len(sc.CodeSnapshot.ModifiedFiles))
			copy(clone.CodeSnapshot.ModifiedFiles, sc.CodeSnapshot.ModifiedFiles)
		}
	}
	if sc.DocSnapshot != nil {
		clone.DocSnapshot = &DocSnapshot{
			SpecFiles: make([]string, len(sc.DocSnapshot.SpecFiles)),
			DocFiles:  make([]string, len(sc.DocSnapshot.DocFiles)),
			Reports:   make([]string, len(sc.DocSnapshot.Reports)),
		}
		copy(clone.DocSnapshot.SpecFiles, sc.DocSnapshot.SpecFiles)
		copy(clone.DocSnapshot.DocFiles, sc.DocSnapshot.DocFiles)
		copy(clone.DocSnapshot.Reports, sc.DocSnapshot.Reports)
	}
	if sc.StateSnapshot != nil {
		clone.StateSnapshot = &StateSnapshot{
			RunningTasks:   sc.StateSnapshot.RunningTasks,
			PendingTasks:   sc.StateSnapshot.PendingTasks,
			CompletedTasks: sc.StateSnapshot.CompletedTasks,
			FailedTasks:    sc.StateSnapshot.FailedTasks,
		}
	}
	return clone
}

// AddLineage adds a parent task ID to the lineage.
func (sc *SelfContext) AddLineage(parentTaskID string) {
	if sc.TaskLineage == nil {
		sc.TaskLineage = make([]string, 0)
	}
	sc.TaskLineage = append(sc.TaskLineage, parentTaskID)
}

// SetAutonomyLevel sets the autonomy level based on competence score.
func (sc *SelfContext) SetAutonomyLevel(level AutonomyLevel) {
	sc.AutonomyLevel = level
}

// UpdateCompetenceScore updates the competence score and adjusts autonomy level.
func (sc *SelfContext) UpdateCompetenceScore(score float64) {
	sc.CompetenceScore = score
	if score >= 0.9 {
		sc.AutonomyLevel = AutonomyLevelLearn
	} else if score >= 0.7 {
		sc.AutonomyLevel = AutonomyLevelPlan
	} else if score >= 0.5 {
		sc.AutonomyLevel = AutonomyLevelDecide
	} else {
		sc.AutonomyLevel = AutonomyLevelExecute
	}
}
