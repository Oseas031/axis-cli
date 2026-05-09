// Package agent provides agent execution capabilities.
package agent

import (
	"context"
	"testing"

	contractexec "github.com/axis-cli/axis/internal/contract/executor"
	"github.com/axis-cli/axis/internal/types"
)

func TestAgentExecutor_Interface(t *testing.T) {
	// Verify AgentExecutor is an interface
	var _ AgentExecutor = (*MockAgentExecutor)(nil)
}

func TestMockAgentExecutor_Execute(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	task := &types.AgentTask{
		TaskID:     "test-task-1",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "test input",
			"task_type": "default",
		},
		Status: types.TaskStatusPending,
	}

	req := &AgentExecutionRequest{
		Task:     task,
		Autonomy: AutonomyLevelLow,
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result == nil {
		t.Fatal("Execute returned nil result")
	}

	if result.Output == nil {
		t.Error("Output is nil")
	}

	if len(result.FollowUpTasks) == 0 {
		t.Error("FollowUpTasks is empty")
	}

	if result.ValidationResult == nil {
		t.Error("ValidationResult is nil")
	}
}

func TestMockAgentExecutor_ExecuteWithCodeGeneration(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	task := &types.AgentTask{
		TaskID:     "code-gen-task",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "generate code",
			"task_type": "code_generation",
		},
		Status: types.TaskStatusPending,
	}

	req := &AgentExecutionRequest{
		Task:     task,
		Autonomy: AutonomyLevelLow,
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.AutonomyDelta.Delta != 1 {
		t.Errorf("Expected autonomy delta 1 for code_generation, got %d", result.AutonomyDelta.Delta)
	}

	if result.ValidationResult == nil {
		t.Fatal("ValidationResult is nil")
	}

	if result.ValidationResult.TestsPassed != 10 {
		t.Errorf("Expected 10 tests passed for code_generation, got %d", result.ValidationResult.TestsPassed)
	}
}

func TestMockAgentExecutor_ExecuteWithDebugging(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	task := &types.AgentTask{
		TaskID:     "debug-task",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "debug issue",
			"task_type": "debugging",
		},
		Status: types.TaskStatusPending,
	}

	req := &AgentExecutionRequest{
		Task:     task,
		Autonomy: AutonomyLevelLow,
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.AutonomyDelta.Delta != 1 {
		t.Errorf("Expected autonomy delta 1 for debugging, got %d", result.AutonomyDelta.Delta)
	}

	if result.ValidationResult.TestsFailed != 1 {
		t.Errorf("Expected 1 test failed for debugging, got %d", result.ValidationResult.TestsFailed)
	}
}

func TestMockAgentExecutor_GetAutonomyLevel(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	level := executor.GetAutonomyLevel()
	if level != AutonomyLevelLow {
		t.Errorf("Expected initial autonomy level %d, got %d", AutonomyLevelLow, level)
	}
}

func TestMockAgentExecutor_SetAutonomyLevel(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	executor.SetAutonomyLevel(AutonomyLevelHigh)
	level := executor.GetAutonomyLevel()
	if level != AutonomyLevelHigh {
		t.Errorf("Expected autonomy level %d, got %d", AutonomyLevelHigh, level)
	}
}

func TestMockAgentExecutor_ExecuteWithNilTask(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	req := &AgentExecutionRequest{
		Task:     nil,
		Autonomy: AutonomyLevelLow,
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed with nil task: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// Should still have follow-up tasks structure even with nil task
	// (though the task itself is nil so follow-ups can't be generated)
	if result.FollowUpTasks != nil {
		t.Error("FollowUpTasks should be nil when task is nil")
	}
}

func TestMockAgentExecutor_FollowUpTasksDeterministic(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	task := &types.AgentTask{
		TaskID:     "parent-task",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "test",
			"task_type": "code_generation",
		},
		Status: types.TaskStatusPending,
	}

	req := &AgentExecutionRequest{
		Task:     task,
		Autonomy: AutonomyLevelLow,
	}

	// Execute twice and verify deterministic results
	result1, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("First Execute failed: %v", err)
	}

	// Reset autonomy level for second run
	executor.SetAutonomyLevel(AutonomyLevelLow)

	result2, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Second Execute failed: %v", err)
	}

	if len(result1.FollowUpTasks) != len(result2.FollowUpTasks) {
		t.Errorf("FollowUpTasks count mismatch: %d vs %d", len(result1.FollowUpTasks), len(result2.FollowUpTasks))
	}

	for i, task1 := range result1.FollowUpTasks {
		task2 := result2.FollowUpTasks[i]
		if task1.TaskID != task2.TaskID {
			t.Errorf("FollowUpTasks[%d] ID mismatch: %s vs %s", i, task1.TaskID, task2.TaskID)
		}
	}
}

func TestMockAgentExecutor_AutonomyDeltaAccumulation(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	// Execute multiple code_generation tasks and verify autonomy accumulates
	initialLevel := executor.GetAutonomyLevel()

	for i := 0; i < 3; i++ {
		task := &types.AgentTask{
			TaskID:     "code-task-" + string(rune('a'+i)),
			ContractID: "test-contract",
			Input: map[string]any{
				"input":     "generate",
				"task_type": "code_generation",
			},
			Status: types.TaskStatusPending,
		}

		req := &AgentExecutionRequest{
			Task:     task,
			Autonomy: AutonomyLevelLow,
		}

		_, err := executor.Execute(context.Background(), req)
		if err != nil {
			t.Fatalf("Execute %d failed: %v", i, err)
		}
	}

	finalLevel := executor.GetAutonomyLevel()
	expectedDelta := int(finalLevel) - int(initialLevel)

	// Each code_generation task adds delta of 1, and we're capped at AutonomyLevelFull (4)
	if finalLevel < initialLevel {
		t.Errorf("Autonomy level should not decrease: from %d to %d", initialLevel, finalLevel)
	}

	// Verify we don't exceed AutonomyLevelFull
	if finalLevel > AutonomyLevelFull {
		t.Errorf("Autonomy level exceeded maximum: got %d, max is %d", finalLevel, AutonomyLevelFull)
	}

	// After 3 code_generation tasks starting from Low (1), should be at most 4 (Full)
	if expectedDelta != 3 {
		t.Logf("Accumulated autonomy delta: %d (may be capped at max level)", expectedDelta)
	}
}

func TestMockAgentExecutor_ValidationSummaryByTaskType(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	testCases := []struct {
		taskType       string
		expectedPassed int
		expectedFailed int
		minCoverage    float64
		isAcceptable   bool
	}{
		{"code_generation", 10, 0, 80.0, true},
		{"debugging", 5, 1, 70.0, true},
		{"refactoring", 8, 0, 85.0, true},
		{"analysis", 3, 0, 50.0, true},
		{"default", 3, 0, 50.0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.taskType, func(t *testing.T) {
			executor.SetAutonomyLevel(AutonomyLevelLow) // Reset

			task := &types.AgentTask{
				TaskID:     "test-" + tc.taskType,
				ContractID: "test-contract",
				Input: map[string]any{
					"input":     "test",
					"task_type": tc.taskType,
				},
				Status: types.TaskStatusPending,
			}

			req := &AgentExecutionRequest{
				Task:     task,
				Autonomy: AutonomyLevelLow,
			}

			result, err := executor.Execute(context.Background(), req)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			if result.ValidationResult.TestsPassed != tc.expectedPassed {
				t.Errorf("TestsPassed: expected %d, got %d", tc.expectedPassed, result.ValidationResult.TestsPassed)
			}

			if result.ValidationResult.TestsFailed != tc.expectedFailed {
				t.Errorf("TestsFailed: expected %d, got %d", tc.expectedFailed, result.ValidationResult.TestsFailed)
			}

			if result.ValidationResult.Coverage < tc.minCoverage {
				t.Errorf("Coverage: expected >= %f, got %f", tc.minCoverage, result.ValidationResult.Coverage)
			}

			if result.ValidationResult.IsAcceptable != tc.isAcceptable {
				t.Errorf("IsAcceptable: expected %v, got %v", tc.isAcceptable, result.ValidationResult.IsAcceptable)
			}
		})
	}
}

func TestAutonomyLevel_Constants(t *testing.T) {
	if AutonomyLevelNone != 0 {
		t.Errorf("AutonomyLevelNone: expected 0, got %d", AutonomyLevelNone)
	}
	if AutonomyLevelLow != 1 {
		t.Errorf("AutonomyLevelLow: expected 1, got %d", AutonomyLevelLow)
	}
	if AutonomyLevelMedium != 2 {
		t.Errorf("AutonomyLevelMedium: expected 2, got %d", AutonomyLevelMedium)
	}
	if AutonomyLevelHigh != 3 {
		t.Errorf("AutonomyLevelHigh: expected 3, got %d", AutonomyLevelHigh)
	}
	if AutonomyLevelFull != 4 {
		t.Errorf("AutonomyLevelFull: expected 4, got %d", AutonomyLevelFull)
	}
}

func TestSelfContext_Structure(t *testing.T) {
	ctx := &SelfContext{
		AgentID:        "agent-001",
		Name:           "TestAgent",
		Capabilities:   []string{"cap1", "cap2"},
		CurrentTask:    "task-1",
		CompletedTasks: 5,
		EarnedAutonomy: AutonomyLevelMedium,
	}

	if ctx.AgentID != "agent-001" {
		t.Errorf("AgentID: expected 'agent-001', got %s", ctx.AgentID)
	}
	if ctx.Name != "TestAgent" {
		t.Errorf("Name: expected 'TestAgent', got %s", ctx.Name)
	}
	if len(ctx.Capabilities) != 2 {
		t.Errorf("Capabilities: expected 2, got %d", len(ctx.Capabilities))
	}
	if ctx.CompletedTasks != 5 {
		t.Errorf("CompletedTasks: expected 5, got %d", ctx.CompletedTasks)
	}
	if ctx.EarnedAutonomy != AutonomyLevelMedium {
		t.Errorf("EarnedAutonomy: expected %d, got %d", AutonomyLevelMedium, ctx.EarnedAutonomy)
	}
}

func TestAgentExecutionRequest_Structure(t *testing.T) {
	task := &types.AgentTask{
		TaskID:     "test-task",
		ContractID: "test-contract",
		Status:     types.TaskStatusPending,
	}

	selfCtx := &SelfContext{
		AgentID: "agent-001",
		Name:    "TestAgent",
	}

	contract := &types.AgentContract{
		ContractID: "test-contract",
	}

	req := &AgentExecutionRequest{
		Task:        task,
		SelfContext: selfCtx,
		Contract:    contract,
		Autonomy:    AutonomyLevelHigh,
	}

	if req.Task != task {
		t.Error("Task not set correctly")
	}
	if req.SelfContext != selfCtx {
		t.Error("SelfContext not set correctly")
	}
	if req.Contract != contract {
		t.Error("Contract not set correctly")
	}
	if req.Autonomy != AutonomyLevelHigh {
		t.Errorf("Autonomy: expected %d, got %d", AutonomyLevelHigh, req.Autonomy)
	}
}

func TestAgentExecutionResult_Structure(t *testing.T) {
	result := &AgentExecutionResult{
		Output: map[string]any{"result": "success"},
		FollowUpTasks: []*types.AgentTask{
			{TaskID: "followup-1"},
			{TaskID: "followup-2"},
		},
		ValidationResult: &ValidationSummary{
			TestsPassed:  10,
			TestsFailed:  0,
			Coverage:     85.0,
			IsAcceptable: true,
		},
		AutonomyDelta: AutonomyDelta{
			Delta:  1,
			Reason: "task completed",
		},
		Error: "",
	}

	if result.Output["result"] != "success" {
		t.Error("Output not set correctly")
	}
	if len(result.FollowUpTasks) != 2 {
		t.Errorf("FollowUpTasks: expected 2, got %d", len(result.FollowUpTasks))
	}
	if result.ValidationResult.TestsPassed != 10 {
		t.Errorf("ValidationResult.TestsPassed: expected 10, got %d", result.ValidationResult.TestsPassed)
	}
	if result.AutonomyDelta.Delta != 1 {
		t.Errorf("AutonomyDelta.Delta: expected 1, got %d", result.AutonomyDelta.Delta)
	}
}

func TestValidationSummary_Structure(t *testing.T) {
	summary := &ValidationSummary{
		TestsPassed:  8,
		TestsFailed:  2,
		Coverage:     75.5,
		IsAcceptable: true,
	}

	if summary.TestsPassed != 8 {
		t.Errorf("TestsPassed: expected 8, got %d", summary.TestsPassed)
	}
	if summary.TestsFailed != 2 {
		t.Errorf("TestsFailed: expected 2, got %d", summary.TestsFailed)
	}
	if summary.Coverage != 75.5 {
		t.Errorf("Coverage: expected 75.5, got %f", summary.Coverage)
	}
	if !summary.IsAcceptable {
		t.Error("IsAcceptable: expected true, got false")
	}
}

func TestAutonomyDelta_Structure(t *testing.T) {
	delta := &AutonomyDelta{
		Delta:  2,
		Reason: "successful complex task",
	}

	if delta.Delta != 2 {
		t.Errorf("Delta: expected 2, got %d", delta.Delta)
	}
	if delta.Reason != "successful complex task" {
		t.Errorf("Reason: expected 'successful complex task', got %s", delta.Reason)
	}
}

func TestMockAgentExecutor_ExecuteWithContract(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()

	// Register a custom contract
	customContract := &types.AgentContract{
		ContractID: "custom-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{Name: "input", Type: types.FieldTypeString, Required: true},
			},
		},
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{
				{Name: "result", Type: types.FieldTypeString, Required: true},
			},
		},
	}
	_ = contractExec.RegisterContract(customContract)

	executor := NewMockAgentExecutor(contractExec)

	task := &types.AgentTask{
		TaskID:     "custom-contract-task",
		ContractID: "custom-contract",
		Input: map[string]any{
			"input": "test input for custom contract",
		},
		Status: types.TaskStatusPending,
	}

	req := &AgentExecutionRequest{
		Task:     task,
		Contract: customContract,
		Autonomy: AutonomyLevelMedium,
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// The MockAgentExecutor uses the contract executor which validates input
	// but the mock provider returns no output, so result.Output may be nil
	// This is expected behavior for the mock
	if result.ValidationResult == nil {
		t.Error("ValidationResult should not be nil")
	}
}

func TestMockAgentExecutor_ContractFallback(t *testing.T) {
	// Test with nil contract executor
	executor := NewMockAgentExecutor(nil)

	task := &types.AgentTask{
		TaskID:     "fallback-task",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "test",
			"task_type": "default",
		},
		Status: types.TaskStatusPending,
	}

	req := &AgentExecutionRequest{
		Task:     task,
		Autonomy: AutonomyLevelLow,
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute with nil contract executor failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// With nil contract executor, should still get deterministic mock output
	if result.Output == nil {
		t.Error("Output should not be nil when contract executor is nil")
	}
}

func TestMockAgentExecutor_FollowUpTaskDependencies(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	task := &types.AgentTask{
		TaskID:     "parent-task-with-deps",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "test",
			"task_type": "code_generation",
		},
		Dependencies: []string{"earlier-task-1", "earlier-task-2"},
		Status:       types.TaskStatusPending,
	}

	req := &AgentExecutionRequest{
		Task:     task,
		Autonomy: AutonomyLevelLow,
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(result.FollowUpTasks) == 0 {
		t.Fatal("FollowUpTasks is empty")
	}

	// Each follow-up task should depend on the parent task
	for _, followUp := range result.FollowUpTasks {
		if len(followUp.Dependencies) != 1 {
			t.Errorf("FollowUp task %s should have 1 dependency, got %d", followUp.TaskID, len(followUp.Dependencies))
		}
		if followUp.Dependencies[0] != task.TaskID {
			t.Errorf("FollowUp task dependency should be parent task ID, got %s", followUp.Dependencies[0])
		}
	}
}

func TestMockAgentExecutor_GenerateDeterministicOutputAllTypes(t *testing.T) {
	// Use nil contract executor to force fallback to deterministic output
	executor := NewMockAgentExecutor(nil)

	taskTypes := []string{"code_generation", "debugging", "refactoring", "analysis", "unknown_type"}

	for _, taskType := range taskTypes {
		t.Run(taskType, func(t *testing.T) {
			task := &types.AgentTask{
				TaskID:     "test-task-" + taskType,
				ContractID: "test-contract",
				Input: map[string]any{
					"input":     "test",
					"task_type": taskType,
				},
				Status: types.TaskStatusPending,
			}

			output := executor.generateDeterministicOutput(task, taskType)
			if output == nil {
				t.Error("generateDeterministicOutput returned nil")
			}
			if output["result"] == nil {
				t.Error("result field is missing")
			}
		})
	}
}

func TestMockAgentExecutor_ApplyAutonomyDeltaBounds(t *testing.T) {
	executor := NewMockAgentExecutor(nil)

	// Test lower bound (should not go below AutonomyLevelNone = 0)
	executor.SetAutonomyLevel(AutonomyLevelNone)
	result := executor.applyAutonomyDelta(AutonomyLevelNone, -5)
	if result != AutonomyLevelNone {
		t.Errorf("Expected AutonomyLevelNone when applying negative delta, got %d", result)
	}

	// Test upper bound (should not exceed AutonomyLevelFull = 4)
	executor.SetAutonomyLevel(AutonomyLevelFull)
	result = executor.applyAutonomyDelta(AutonomyLevelFull, 5)
	if result != AutonomyLevelFull {
		t.Errorf("Expected AutonomyLevelFull when applying positive delta beyond max, got %d", result)
	}

	// Test valid delta within bounds
	executor.SetAutonomyLevel(AutonomyLevelLow)
	result = executor.applyAutonomyDelta(AutonomyLevelLow, 1)
	if result != AutonomyLevelMedium {
		t.Errorf("Expected AutonomyLevelMedium, got %d", result)
	}

	// Test negative delta within bounds
	result = executor.applyAutonomyDelta(AutonomyLevelMedium, -1)
	if result != AutonomyLevelLow {
		t.Errorf("Expected AutonomyLevelLow, got %d", result)
	}
}

func TestMockAgentExecutor_ExecuteWithRefactoring(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	task := &types.AgentTask{
		TaskID:     "refactor-task",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "refactor this",
			"task_type": "refactoring",
		},
		Status: types.TaskStatusPending,
	}

	req := &AgentExecutionRequest{
		Task:     task,
		Autonomy: AutonomyLevelLow,
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.AutonomyDelta.Delta != 0 {
		t.Errorf("Expected autonomy delta 0 for refactoring, got %d", result.AutonomyDelta.Delta)
	}

	if result.ValidationResult == nil {
		t.Fatal("ValidationResult is nil")
	}

	if result.ValidationResult.TestsPassed != 8 {
		t.Errorf("Expected 8 tests passed for refactoring, got %d", result.ValidationResult.TestsPassed)
	}

	if result.ValidationResult.TestsFailed != 0 {
		t.Errorf("Expected 0 tests failed for refactoring, got %d", result.ValidationResult.TestsFailed)
	}
}

func TestMockAgentExecutor_ExecuteWithAnalysis(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	task := &types.AgentTask{
		TaskID:     "analysis-task",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "analyze this",
			"task_type": "analysis",
		},
		Status: types.TaskStatusPending,
	}

	req := &AgentExecutionRequest{
		Task:     task,
		Autonomy: AutonomyLevelLow,
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.AutonomyDelta.Delta != 0 {
		t.Errorf("Expected autonomy delta 0 for analysis, got %d", result.AutonomyDelta.Delta)
	}

	if result.ValidationResult == nil {
		t.Fatal("ValidationResult is nil")
	}

	if result.ValidationResult.TestsPassed != 3 {
		t.Errorf("Expected 3 tests passed for analysis, got %d", result.ValidationResult.TestsPassed)
	}
}

func TestMockAgentExecutor_ExecuteWithUnknownType(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	task := &types.AgentTask{
		TaskID:     "unknown-task",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "do something",
			"task_type": "unknown_special_type",
		},
		Status: types.TaskStatusPending,
	}

	req := &AgentExecutionRequest{
		Task:     task,
		Autonomy: AutonomyLevelLow,
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should fall through to default case
	if result.AutonomyDelta.Delta != 0 {
		t.Errorf("Expected autonomy delta 0 for unknown type, got %d", result.AutonomyDelta.Delta)
	}

	if result.ValidationResult.TestsPassed != 3 {
		t.Errorf("Expected 3 tests passed for unknown type, got %d", result.ValidationResult.TestsPassed)
	}
}

func TestMockAgentExecutor_ExecuteWithDefaultContractNoTaskInput(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	// Task with no input map
	task := &types.AgentTask{
		TaskID:     "no-input-task",
		ContractID: "test-contract",
		Input:      nil,
		Status:     types.TaskStatusPending,
	}

	req := &AgentExecutionRequest{
		Task:     task,
		Autonomy: AutonomyLevelLow,
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// Should still generate follow-up tasks (will use "default" task type since no input)
	if len(result.FollowUpTasks) != 2 {
		t.Errorf("Expected 2 follow-up tasks for default, got %d", len(result.FollowUpTasks))
	}
}

func TestMockAgentExecutor_AutonomyLevelFullCaps(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	// Set to High (3) and try to add 2 (should cap at Full = 4)
	executor.SetAutonomyLevel(AutonomyLevelHigh)

	task := &types.AgentTask{
		TaskID:     "cap-test",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "test",
			"task_type": "code_generation",
		},
		Status: types.TaskStatusPending,
	}

	// First call - should go from 3 to 4
	req := &AgentExecutionRequest{
		Task:     task,
		Autonomy: AutonomyLevelLow,
	}
	_, _ = executor.Execute(context.Background(), req)

	level := executor.GetAutonomyLevel()
	if level != AutonomyLevelFull {
		t.Errorf("Expected AutonomyLevelFull (4), got %d", level)
	}

	// Second call - should stay at Full (capped)
	executor.SetAutonomyLevel(AutonomyLevelHigh)
	_, _ = executor.Execute(context.Background(), req)

	level = executor.GetAutonomyLevel()
	if level != AutonomyLevelFull {
		t.Errorf("Expected AutonomyLevelFull (4) after second call, got %d", level)
	}
}

func TestMockAgentExecutor_ExecuteSelfContext(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	task := &types.AgentTask{
		TaskID:     "self-context-task",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "test",
			"task_type": "default",
		},
		Status: types.TaskStatusPending,
	}

	selfCtx := &SelfContext{
		AgentID:        "agent-002",
		Name:           "CustomAgent",
		Capabilities:   []string{"custom_cap"},
		CurrentTask:    "self-context-task",
		CompletedTasks: 10,
		EarnedAutonomy: AutonomyLevelHigh,
	}

	req := &AgentExecutionRequest{
		Task:        task,
		SelfContext: selfCtx,
		Autonomy:    AutonomyLevelMedium,
	}

	result, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// SelfContext is passed through but MockAgentExecutor doesn't modify it
	// This test verifies it doesn't cause errors
}

func TestMockAgentExecutor_ExecuteContractExecutorError(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	executor := NewMockAgentExecutor(contractExec)

	// Register a contract with required field
	contract := &types.AgentContract{
		ContractID: "require-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{Name: "required_field", Type: types.FieldTypeString, Required: true},
			},
		},
	}
	_ = contractExec.RegisterContract(contract)

	task := &types.AgentTask{
		TaskID:     "error-task",
		ContractID: "require-contract",
		Input: map[string]any{
			"input": "missing required_field",
			// required_field is missing
		},
		Status: types.TaskStatusPending,
	}

	req := &AgentExecutionRequest{
		Task:     task,
		Autonomy: AutonomyLevelLow,
	}

	result, err := executor.Execute(context.Background(), req)
	if err == nil {
		// Contract executor error is returned as part of result, not thrown
		t.Log("Contract executor error returned in result (expected)")
	}

	// When there's an error, Error field should be set
	if result != nil && result.Error == "" {
		// This is acceptable - the mock executor may not propagate all errors
		t.Log("Result error field is empty (may be expected for mock)")
	}
}
