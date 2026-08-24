package compactor

import (
	"strings"
	"time"
)

// RecoveryContext holds information needed to restore agent state after compaction.
type RecoveryContext struct {
	ActivePlan  string               // current plan/goal summary
	Feedback    string               // correction feedback from previous attempts
	FileStates  map[string]FileState // file state snapshot
	ExtractedAt time.Time            // extraction timestamp
}

// FileState tracks the state of a single file.
type FileState struct {
	Path      string
	Operation string // "created", "modified", "deleted"
	Timestamp int64
}

// Message represents a chat message for extraction.
type Message struct {
	Role    string
	Content string
	Name    string
	Path    string
}

// ExtractRecoveryContext extracts recovery context from chat history.
func ExtractRecoveryContext(history []Message) *RecoveryContext {
	if len(history) == 0 {
		return &RecoveryContext{
			FileStates:  make(map[string]FileState),
			ExtractedAt: time.Now(),
		}
	}

	ctx := &RecoveryContext{
		FileStates:  make(map[string]FileState),
		ExtractedAt: time.Now(),
	}

	ctx.ActivePlan = extractActivePlan(history)
	ctx.Feedback = extractFeedback(history)
	ctx.FileStates = extractFileStates(history)

	return ctx
}

// extractActivePlan extracts the current execution plan from history.
func extractActivePlan(history []Message) string {
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		if msg.Role == "assistant" {
			content := msg.Content
			if idx := strings.Index(content, "## Plan"); idx != -1 {
				return content[idx:]
			}
			if idx := strings.Index(content, "## 执行计划"); idx != -1 {
				return content[idx:]
			}
			if len(content) > 500 {
				return content[:500] + "..."
			}
			return content
		}
	}
	return ""
}

// extractFeedback extracts the most recent user feedback from history.
func extractFeedback(history []Message) string {
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		if msg.Role == "user" {
			content := msg.Content
			if len(content) > 1000 {
				return content[:1000] + "..."
			}
			return content
		}
	}
	return ""
}

// extractFileStates extracts file states from tool messages in history.
func extractFileStates(history []Message) map[string]FileState {
	states := make(map[string]FileState)

	for _, msg := range history {
		if msg.Role == "tool" {
			path := msg.Path
			if path == "" {
				continue
			}

			operation := "modified"
			content := msg.Content

			if strings.Contains(content, "created") || strings.Contains(content, "创建") {
				operation = "created"
			} else if strings.Contains(content, "deleted") || strings.Contains(content, "删除") {
				operation = "deleted"
			}

			states[path] = FileState{
				Path:      path,
				Operation: operation,
				Timestamp: time.Now().Unix(),
			}
		}
	}

	return states
}

// MergeRecoveryContexts merges multiple recovery contexts, later ones win.
func MergeRecoveryContexts(contexts ...*RecoveryContext) *RecoveryContext {
	merged := &RecoveryContext{
		FileStates: make(map[string]FileState),
	}

	for _, ctx := range contexts {
		if ctx == nil {
			continue
		}

		if ctx.ActivePlan != "" {
			merged.ActivePlan = ctx.ActivePlan
		}

		if ctx.Feedback != "" {
			merged.Feedback = ctx.Feedback
		}

		for path, state := range ctx.FileStates {
			merged.FileStates[path] = state
		}

		if ctx.ExtractedAt.After(merged.ExtractedAt) {
			merged.ExtractedAt = ctx.ExtractedAt
		}
	}

	return merged
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
		paths := make([]string, 0, len(rc.FileStates))
		for p := range rc.FileStates {
			paths = append(paths, p)
		}
		parts = append(parts, "[file_state] Recently modified: "+strings.Join(paths, ", "))
	}
	return strings.Join(parts, "\n")
}
