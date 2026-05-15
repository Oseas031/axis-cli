package compactor

import "strings"

// v1: RecoveryContext is the data structure only. Population logic not yet implemented.
// TODO: extract ActivePlan/Feedback/FileStates from history before compaction.

// RecoveryContext holds semantic state that must survive compaction.
type RecoveryContext struct {
	ActivePlan string   // current plan/goal summary
	Feedback   string   // correction feedback from previous attempts
	FileStates []string // recently read file paths
}

// BuildRecoveryMessage creates a system-level message to inject after compaction.
func (rc *RecoveryContext) BuildRecoveryMessage() string {
	if rc == nil {
		return ""
	}
	var parts []string
	parts = append(parts, "[compact_boundary] Context was compacted. Preserved state:")
	if rc.ActivePlan != "" {
		parts = append(parts, "[active_plan] "+rc.ActivePlan)
	}
	if rc.Feedback != "" {
		parts = append(parts, "[feedback] "+rc.Feedback)
	}
	if len(rc.FileStates) > 0 {
		parts = append(parts, "[file_state] Recently accessed: "+strings.Join(rc.FileStates, ", "))
	}
	return strings.Join(parts, "\n")
}
