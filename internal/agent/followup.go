// Package agent provides follow-up task generation for agent execution results.
package agent

import (
	"fmt"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

// GenerateFollowUpTasks generates actual AgentTask objects from an AgentExecutionResult.
// It creates follow-up tasks with proper dependencies and metadata based on the result.
func GenerateFollowUpTasks(result *AgentExecutionResult, parentTask *types.AgentTask) []*types.AgentTask {
	if result == nil || len(result.FollowUpTasks) == 0 {
		return nil
	}

	tasks := make([]*types.AgentTask, 0, len(result.FollowUpTasks))
	now := time.Now()

	for i, followUp := range result.FollowUpTasks {
		taskID := followUp.TaskID
		if taskID == "" {
			taskID = fmt.Sprintf("%s-followup-%d", parentTask.TaskID, i+1)
		}

		// Create dependencies - must include parent task
		deps := make([]string, len(followUp.Dependencies)+1)
		deps[0] = parentTask.TaskID
		copy(deps[1:], followUp.Dependencies)

		// Merge input with parent context
		input := make(map[string]any)
		for k, v := range parentTask.Input {
			input[k] = v
		}
		for k, v := range followUp.Input {
			input[k] = v
		}
		// Add parent context
		input["parent_task_id"] = parentTask.TaskID
		input["parent_result"] = result.Output

		// Build metadata
		metadata := make(map[string]string)
		for k, v := range followUp.Metadata {
			metadata[k] = v
		}
		metadata["followup_index"] = fmt.Sprintf("%d", i+1)
		metadata["followup_count"] = fmt.Sprintf("%d", len(result.FollowUpTasks))
		if result.ValidationResult != nil {
			metadata["parent_validation_passed"] = fmt.Sprintf("%t", result.ValidationResult.IsAcceptable)
		}
		if result.AutonomyDelta.Delta != 0 {
			metadata["autonomy_delta"] = fmt.Sprintf("%d", result.AutonomyDelta.Delta)
		}

		task := &types.AgentTask{
			TaskID:       taskID,
			ContractID:   followUp.ContractID,
			Input:        input,
			Dependencies: deps,
			Status:       types.TaskStatusPending,
			CreatedAt:    now,
			Metadata:     metadata,
		}

		tasks = append(tasks, task)
	}

	return tasks
}

// GenerateFollowUpTasksFromMap generates follow-up tasks from a map-based result structure.
// This is used when the result comes from a simple map output without a full result object.
func GenerateFollowUpTasksFromMap(output map[string]any, parentTask *types.AgentTask) []*types.AgentTask {
	if output == nil {
		return nil
	}

	// Look for follow_ups key in output
	followUpsRaw, ok := output["follow_ups"]
	if !ok {
		return nil
	}

	followUpList, ok := followUpsRaw.([]string)
	if !ok {
		return nil
	}

	tasks := make([]*types.AgentTask, 0, len(followUpList))
	now := time.Now()

	for i, followUpType := range followUpList {
		taskID := fmt.Sprintf("%s-followup-%d", parentTask.TaskID, i+1)

		task := &types.AgentTask{
			TaskID:     taskID,
			ContractID: parentTask.ContractID, // Inherit parent contract
			Input: map[string]any{
				"parent_task":   parentTask.TaskID,
				"followup_type": followUpType,
				"input":         fmt.Sprintf("followup-for-%s", parentTask.TaskID),
			},
			Dependencies: []string{parentTask.TaskID},
			Status:       types.TaskStatusPending,
			CreatedAt:    now,
			Metadata: map[string]string{
				"parent_task_id": parentTask.TaskID,
				"followup_type":  followUpType,
				"followup_index": fmt.Sprintf("%d", i+1),
				"followup_count": fmt.Sprintf("%d", len(followUpList)),
			},
		}

		tasks = append(tasks, task)
	}

	return tasks
}

// BuildSpawnTaskInput creates the input map for a spawn contract task.
func BuildSpawnTaskInput(reviewResult map[string]any, currentTaskID string) map[string]any {
	return map[string]any{
		"review_result":   reviewResult,
		"current_task_id": currentTaskID,
	}
}

// ExtractTerminationReason determines the termination reason from a review result.
func ExtractTerminationReason(reviewResult map[string]any) string {
	if reviewResult == nil {
		return "unknown"
	}

	// Check approval_status field
	if status, ok := reviewResult["approval_status"].(string); ok {
		switch status {
		case "approved":
			return "task_approved"
		case "rejected":
			return "task_rejected"
		case "needs_changes":
			return "needs_modification"
		}
	}

	// Check termination_reason field
	if reason, ok := reviewResult["termination_reason"].(string); ok && reason != "" {
		return reason
	}

	return "completion"
}
