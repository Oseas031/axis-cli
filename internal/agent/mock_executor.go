// Package agent provides agent execution capabilities.
package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	contractexec "github.com/axis-cli/axis/internal/contract/executor"
	"github.com/axis-cli/axis/internal/types"
)

// MockAgentExecutor is a deterministic mock implementation of AgentExecutor.
type MockAgentExecutor struct {
	mu              sync.RWMutex
	contractExec    contractexec.ContractExecutor
	autonomyLevel   AutonomyLevel
	defaultContract *types.AgentContract
}

// NewMockAgentExecutor creates a new mock agent executor.
func NewMockAgentExecutor(contractExec contractexec.ContractExecutor) *MockAgentExecutor {
	// Create a default echo contract if none is provided
	defaultContract := &types.AgentContract{
		ContractID: "agent-echo-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{Name: "input", Type: types.FieldTypeString, Required: true},
				{Name: "task_type", Type: types.FieldTypeString, Required: false},
			},
		},
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{
				{Name: "result", Type: types.FieldTypeString, Required: true},
				{Name: "follow_ups", Type: types.FieldTypeArray, Required: false},
				{Name: "autonomy_delta", Type: types.FieldTypeInt, Required: false},
			},
		},
	}

	return &MockAgentExecutor{
		contractExec:    contractExec,
		autonomyLevel:   AutonomyLevelLow,
		defaultContract: defaultContract,
	}
}

// Execute runs the task using the underlying contract executor.
// It generates deterministic follow-up tasks and validation summaries.
func (e *MockAgentExecutor) Execute(ctx context.Context, req *AgentExecutionRequest) (*AgentExecutionResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Use default contract if none provided
	contract := req.Contract
	if contract == nil {
		contract = e.defaultContract
	}

	// Register the default contract if using it
	if contract == e.defaultContract && e.contractExec != nil {
		_ = e.contractExec.RegisterContract(contract)
	}

	// Extract task_type from input for deterministic behavior
	taskType := "default"
	if req.Task != nil && req.Task.Input != nil {
		if tt, ok := req.Task.Input["task_type"].(string); ok {
			taskType = tt
		}
	}

	// Build input for contract execution
	input := make(map[string]any)
	if req.Task != nil {
		for k, v := range req.Task.Input {
			input[k] = v
		}
		if input["input"] == nil {
			input["input"] = fmt.Sprintf("task-%s", req.Task.TaskID)
		}
	}

	var output map[string]any
	var err error

	// Use contract executor if available
	if e.contractExec != nil {
		execResult, execErr := e.contractExec.Execute(ctx, contract.ContractID, input)
		if execErr != nil {
			err = fmt.Errorf("contract execution failed: %w", execErr)
		} else if execResult != nil {
			output = execResult.Output
			if execResult.Error != "" && err == nil {
				err = fmt.Errorf("%s", execResult.Error)
			}
		}
	} else {
		// Fallback to deterministic mock output
		output = e.generateDeterministicOutput(req.Task, taskType)
	}

	// Generate deterministic follow-up tasks
	followUpTasks := e.generateFollowUpTasks(req.Task, taskType)

	// Calculate autonomy delta based on task type
	autonomyDelta := e.calculateAutonomyDelta(taskType)

	// Generate validation summary
	validationSummary := e.generateValidationSummary(taskType)

	result := &AgentExecutionResult{
		Output:           output,
		FollowUpTasks:    followUpTasks,
		ValidationResult: validationSummary,
		AutonomyDelta:    autonomyDelta,
	}

	if err != nil {
		result.Error = err.Error()
	}

	// Update autonomy level based on delta
	e.autonomyLevel = e.applyAutonomyDelta(e.autonomyLevel, autonomyDelta.Delta)

	return result, nil
}

// GetAutonomyLevel returns the current autonomy level.
func (e *MockAgentExecutor) GetAutonomyLevel() AutonomyLevel {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.autonomyLevel
}

// SetAutonomyLevel sets the autonomy level (for testing).
func (e *MockAgentExecutor) SetAutonomyLevel(level AutonomyLevel) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.autonomyLevel = level
}

// generateDeterministicOutput creates deterministic output based on task type.
func (e *MockAgentExecutor) generateDeterministicOutput(task *types.AgentTask, taskType string) map[string]any {
	result := fmt.Sprintf("completed-%s", task.TaskID)

	var followUps []string
	autonomyDelta := 0

	switch taskType {
	case "code_generation":
		result = fmt.Sprintf("generated-code-for-%s", task.TaskID)
		followUps = []string{"review", "test", "deploy"}
		autonomyDelta = 1
	case "debugging":
		result = fmt.Sprintf("debugged-%s-successfully", task.TaskID)
		followUps = []string{"verify-fix", "regression-test"}
		autonomyDelta = 1
	case "refactoring":
		result = fmt.Sprintf("refactored-%s", task.TaskID)
		followUps = []string{"validate", "benchmark"}
		autonomyDelta = 0
	case "analysis":
		result = fmt.Sprintf("analysis-complete-%s", task.TaskID)
		followUps = []string{"present", "archive"}
		autonomyDelta = 0
	default:
		followUps = []string{"validate", "report"}
		autonomyDelta = 0
	}

	return map[string]any{
		"result":         result,
		"follow_ups":     followUps,
		"autonomy_delta": autonomyDelta,
	}
}

// generateFollowUpTasks creates deterministic follow-up tasks.
func (e *MockAgentExecutor) generateFollowUpTasks(task *types.AgentTask, taskType string) []*types.AgentTask {
	if task == nil {
		return nil
	}

	followUpDefs := map[string][]string{
		"code_generation": {"validate-code", "run-tests"},
		"debugging":       {"verify-fix", "check-regression"},
		"refactoring":     {"validate-output", "benchmark"},
		"analysis":        {"present-results", "archive-data"},
		"default":         {"validate", "report"},
	}

	followUpList, ok := followUpDefs[taskType]
	if !ok {
		followUpList = followUpDefs["default"]
	}

	tasks := make([]*types.AgentTask, 0, len(followUpList))
	now := time.Now()

	for i, followUpType := range followUpList {
		followUpID := fmt.Sprintf("%s-followup-%d", task.TaskID, i+1)
		newTask := &types.AgentTask{
			TaskID:     followUpID,
			ContractID: task.ContractID,
			Input: map[string]any{
				"parent_task": task.TaskID,
				"task_type":   followUpType,
				"input":       fmt.Sprintf("followup-for-%s", task.TaskID),
			},
			Dependencies: []string{task.TaskID},
			Status:       types.TaskStatusPending,
			CreatedAt:    now,
			Metadata: map[string]string{
				"parent_task_id": task.TaskID,
				"followup_type":  followUpType,
			},
		}
		tasks = append(tasks, newTask)
	}

	return tasks
}

// calculateAutonomyDelta returns deterministic autonomy changes.
func (e *MockAgentExecutor) calculateAutonomyDelta(taskType string) AutonomyDelta {
	switch taskType {
	case "code_generation":
		return AutonomyDelta{Delta: 1, Reason: "successful code generation"}
	case "debugging":
		return AutonomyDelta{Delta: 1, Reason: "successful debugging"}
	case "refactoring":
		return AutonomyDelta{Delta: 0, Reason: "refactoring maintains autonomy"}
	case "analysis":
		return AutonomyDelta{Delta: 0, Reason: "analysis completed"}
	default:
		return AutonomyDelta{Delta: 0, Reason: "default task completed"}
	}
}

// generateValidationSummary creates deterministic validation results.
func (e *MockAgentExecutor) generateValidationSummary(taskType string) *ValidationSummary {
	switch taskType {
	case "code_generation":
		return &ValidationSummary{
			TestsPassed:  10,
			TestsFailed:  0,
			Coverage:     85.5,
			IsAcceptable: true,
		}
	case "debugging":
		return &ValidationSummary{
			TestsPassed:  5,
			TestsFailed:  1,
			Coverage:     72.0,
			IsAcceptable: true,
		}
	case "refactoring":
		return &ValidationSummary{
			TestsPassed:  8,
			TestsFailed:  0,
			Coverage:     90.0,
			IsAcceptable: true,
		}
	default:
		return &ValidationSummary{
			TestsPassed:  3,
			TestsFailed:  0,
			Coverage:     60.0,
			IsAcceptable: true,
		}
	}
}

// applyAutonomyDelta applies a delta to the current autonomy level.
func (e *MockAgentExecutor) applyAutonomyDelta(current AutonomyLevel, delta int) AutonomyLevel {
	newLevel := int(current) + delta
	if newLevel < int(AutonomyLevelNone) {
		return AutonomyLevelNone
	}
	if newLevel > int(AutonomyLevelFull) {
		return AutonomyLevelFull
	}
	return AutonomyLevel(newLevel)
}
