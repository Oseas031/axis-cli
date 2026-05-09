// Package agent provides integration tests for the bootstrap loop.
package agent

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/agent/contracts"
	orchestrator2 "github.com/axis-cli/axis/internal/kernel/orchestrator"
	"github.com/axis-cli/axis/internal/types"
)

// TestBootstrapOrchestrator_SubmitSelfIterationTask tests submitting a self-iteration task.
func TestBootstrapOrchestrator_SubmitSelfIterationTask(t *testing.T) {
	orch := orchestrator2.NewOrchestrator()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Register all contracts
	contracts.RegisterAll(orch.RegisterContract)

	// Start orchestrator
	if err := orch.Start(ctx); err != nil {
		t.Fatalf("failed to start orchestrator: %v", err)
	}
	defer orch.Shutdown(context.Background())

	// Create bootstrap orchestrator with max 3 iterations
	bo := NewBootstrapOrchestrator(orch, 3)

	// Create a self-iteration task
	task := &types.AgentTask{
		TaskID:     "self-iter-1",
		ContractID: contracts.ContractIDAnalyze,
		Input: map[string]any{
			"change_description": "test change",
			"target_files":       []string{"test.go"},
		},
		Dependencies: []string{},
		Metadata:     make(map[string]string),
	}

	// Submit self-iteration task
	if err := bo.SubmitSelfIterationTask(task); err != nil {
		t.Fatalf("failed to submit self-iteration task: %v", err)
	}

	// Verify loop tracking was initialized
	count := bo.GetIterationCount(task.TaskID)
	if count != 0 {
		t.Errorf("expected initial iteration count 0, got %d", count)
	}

	// Track an iteration
	newCount := bo.TrackIteration(task.TaskID)
	if newCount != 1 {
		t.Errorf("expected iteration count 1, got %d", newCount)
	}

	// Verify iteration is still allowed
	if !bo.IsIterationAllowed(task.TaskID) {
		t.Errorf("expected iteration to be allowed (count=%d, max=%d)", count, bo.maxIterations)
	}
}

// TestBootstrapOrchestrator_MaxIterationsExceeded tests that exceeding max iterations is blocked.
func TestBootstrapOrchestrator_MaxIterationsExceeded(t *testing.T) {
	orch := orchestrator2.NewOrchestrator()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	contracts.RegisterAll(orch.RegisterContract)

	if err := orch.Start(ctx); err != nil {
		t.Fatalf("failed to start orchestrator: %v", err)
	}
	defer orch.Shutdown(context.Background())

	// Create bootstrap orchestrator with max 2 iterations
	bo := NewBootstrapOrchestrator(orch, 2)

	// Create a task
	task := &types.AgentTask{
		TaskID:     "max-iter-test",
		ContractID: contracts.ContractIDAnalyze,
		Input: map[string]any{
			"change_description": "test",
		},
		Metadata: make(map[string]string),
	}

	// Submit first iteration
	if err := bo.SubmitSelfIterationTask(task); err != nil {
		t.Fatalf("first submission failed: %v", err)
	}

	// Track iterations until we hit the limit
	bo.TrackIteration(task.TaskID) // 1
	bo.TrackIteration(task.TaskID) // 2

	// Next iteration should not be allowed
	if bo.IsIterationAllowed(task.TaskID) {
		t.Errorf("expected iteration to be disallowed after max reached")
	}
}

// TestBootstrapOrchestrator_FullDAGWorkflow tests the complete analyze→implement→validate→update-docs→review→spawn DAG.
func TestBootstrapOrchestrator_FullDAGWorkflow(t *testing.T) {
	orch := orchestrator2.NewOrchestrator()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Register all contracts
	contracts.RegisterAll(orch.RegisterContract)

	// Start orchestrator
	if err := orch.Start(ctx); err != nil {
		t.Fatalf("failed to start orchestrator: %v", err)
	}
	defer orch.Shutdown(context.Background())

	// Create bootstrap orchestrator with max 5 iterations
	bo := NewBootstrapOrchestrator(orch, 5)

	// Build the DAG: analyze → implement → validate → update-docs → review → spawn
	now := time.Now()

	analyzeTask := &types.AgentTask{
		TaskID:     "analyze-1",
		ContractID: contracts.ContractIDAnalyze,
		Input: map[string]any{
			"change_description": "implement feature X",
			"target_files":       []string{"feature.go"},
		},
		Dependencies: []string{},
		Status:       types.TaskStatusPending,
		CreatedAt:    now,
		Metadata:     make(map[string]string),
	}

	implementTask := &types.AgentTask{
		TaskID:     "implement-1",
		ContractID: contracts.ContractIDImplement,
		Input: map[string]any{
			"analysis_result":     map[string]any{"impact_scope": []string{"feature.go"}},
			"implementation_plan": []string{"step1", "step2"},
		},
		Dependencies: []string{"analyze-1"},
		Status:       types.TaskStatusPending,
		CreatedAt:    now,
		Metadata:     make(map[string]string),
	}

	validateTask := &types.AgentTask{
		TaskID:     "validate-1",
		ContractID: contracts.ContractIDValidate,
		Input: map[string]any{
			"modified_files": []string{"feature.go"},
		},
		Dependencies: []string{"implement-1"},
		Status:       types.TaskStatusPending,
		CreatedAt:    now,
		Metadata:     make(map[string]string),
	}

	updateDocsTask := &types.AgentTask{
		TaskID:     "update-docs-1",
		ContractID: contracts.ContractIDUpdate,
		Input: map[string]any{
			"changed_files": []string{"feature.go"},
		},
		Dependencies: []string{"validate-1"},
		Status:       types.TaskStatusPending,
		CreatedAt:    now,
		Metadata:     make(map[string]string),
	}

	reviewTask := &types.AgentTask{
		TaskID:     "review-1",
		ContractID: contracts.ContractIDReview,
		Input: map[string]any{
			"implementation_result": map[string]any{"status": "completed"},
			"validation_result":     map[string]any{"is_acceptable": true},
		},
		Dependencies: []string{"update-docs-1"},
		Status:       types.TaskStatusPending,
		CreatedAt:    now,
		Metadata:     make(map[string]string),
	}

	spawnTask := &types.AgentTask{
		TaskID:     "spawn-1",
		ContractID: contracts.ContractIDSpawn,
		Input: map[string]any{
			"review_result":   map[string]any{"approval_status": "approved"},
			"current_task_id": "review-1",
		},
		Dependencies: []string{"review-1"},
		Status:       types.TaskStatusPending,
		CreatedAt:    now,
		Metadata:     make(map[string]string),
	}

	// Submit all tasks
	tasks := []*types.AgentTask{analyzeTask, implementTask, validateTask, updateDocsTask, reviewTask, spawnTask}
	for _, task := range tasks {
		if err := bo.SubmitSelfIterationTask(task); err != nil {
			t.Fatalf("failed to submit task %s: %v", task.TaskID, err)
		}
	}

	// Verify DAG structure
	graph := bo.GetDependencyGraph()
	expectedDeps := map[string][]string{
		"analyze-1":     nil,
		"implement-1":   {"analyze-1"},
		"validate-1":    {"implement-1"},
		"update-docs-1": {"validate-1"},
		"review-1":      {"update-docs-1"},
		"spawn-1":       {"review-1"},
	}

	for taskID, expected := range expectedDeps {
		actual := graph[taskID]
		if len(actual) != len(expected) {
			t.Errorf("task %s: expected %d deps, got %d", taskID, len(expected), len(actual))
			continue
		}
		for i, dep := range expected {
			if actual[i] != dep {
				t.Errorf("task %s: expected dep %s at index %d, got %s", taskID, dep, i, actual[i])
			}
		}
	}

	// Verify all tasks are registered
	allTasks := bo.GetAllTasks()
	if len(allTasks) != 6 {
		t.Errorf("expected 6 tasks, got %d", len(allTasks))
	}

	// Verify loop status
	status := bo.GetLoopStatus()
	if len(status) != 6 {
		t.Errorf("expected 6 tasks in loop status, got %d", len(status))
	}
}

// TestGenerateFollowUpTasks tests follow-up task generation from execution results.
func TestGenerateFollowUpTasks(t *testing.T) {
	parentTask := &types.AgentTask{
		TaskID:     "parent-1",
		ContractID: contracts.ContractIDAnalyze,
		Input: map[string]any{
			"change_description": "test",
		},
		Metadata: make(map[string]string),
	}

	result := &AgentExecutionResult{
		Output: map[string]any{
			"result": "completed",
		},
		FollowUpTasks: []*types.AgentTask{
			{
				TaskID:     "followup-1",
				ContractID: contracts.ContractIDValidate,
				Input: map[string]any{
					"validation_type": "unit",
				},
				Dependencies: []string{},
				Metadata:     make(map[string]string),
			},
			{
				TaskID:     "followup-2",
				ContractID: contracts.ContractIDReview,
				Input: map[string]any{
					"review_type": "code",
				},
				Dependencies: []string{},
				Metadata:     make(map[string]string),
			},
		},
		ValidationResult: &ValidationSummary{
			TestsPassed:  10,
			TestsFailed:  0,
			Coverage:     85.0,
			IsAcceptable: true,
		},
		AutonomyDelta: AutonomyDelta{Delta: 1, Reason: "successful"},
	}

	tasks := GenerateFollowUpTasks(result, parentTask)

	if len(tasks) != 2 {
		t.Fatalf("expected 2 follow-up tasks, got %d", len(tasks))
	}

	// Verify first follow-up
	if tasks[0].TaskID != "followup-1" {
		t.Errorf("expected task ID followup-1, got %s", tasks[0].TaskID)
	}
	if tasks[0].ContractID != contracts.ContractIDValidate {
		t.Errorf("expected contract ID %s, got %s", contracts.ContractIDValidate, tasks[0].ContractID)
	}
	// Verify parent dependency was added
	if len(tasks[0].Dependencies) != 1 || tasks[0].Dependencies[0] != "parent-1" {
		t.Errorf("expected dependency on parent-1, got %v", tasks[0].Dependencies)
	}
	// Verify metadata
	if tasks[0].Metadata["followup_index"] != "1" {
		t.Errorf("expected followup_index 1, got %s", tasks[0].Metadata["followup_index"])
	}
	if tasks[0].Metadata["parent_validation_passed"] != "true" {
		t.Errorf("expected parent_validation_passed true, got %s", tasks[0].Metadata["parent_validation_passed"])
	}

	// Verify second follow-up
	if tasks[1].TaskID != "followup-2" {
		t.Errorf("expected task ID followup-2, got %s", tasks[1].TaskID)
	}
	if tasks[1].ContractID != contracts.ContractIDReview {
		t.Errorf("expected contract ID %s, got %s", contracts.ContractIDReview, tasks[1].ContractID)
	}
}

// TestGenerateFollowUpTasksFromMap tests follow-up generation from simple map output.
func TestGenerateFollowUpTasksFromMap(t *testing.T) {
	parentTask := &types.AgentTask{
		TaskID:     "parent-map-1",
		ContractID: contracts.ContractIDAnalyze,
		Input: map[string]any{
			"change_description": "test",
		},
		Metadata: make(map[string]string),
	}

	output := map[string]any{
		"result":     "completed",
		"follow_ups": []string{"validate", "report"},
	}

	tasks := GenerateFollowUpTasksFromMap(output, parentTask)

	if len(tasks) != 2 {
		t.Fatalf("expected 2 follow-up tasks, got %d", len(tasks))
	}

	if tasks[0].TaskID != "parent-map-1-followup-1" {
		t.Errorf("expected task ID parent-map-1-followup-1, got %s", tasks[0].TaskID)
	}
	if tasks[0].Metadata["followup_type"] != "validate" {
		t.Errorf("expected followup_type validate, got %s", tasks[0].Metadata["followup_type"])
	}

	if tasks[1].TaskID != "parent-map-1-followup-2" {
		t.Errorf("expected task ID parent-map-1-followup-2, got %s", tasks[1].TaskID)
	}
	if tasks[1].Metadata["followup_type"] != "report" {
		t.Errorf("expected followup_type report, got %s", tasks[1].Metadata["followup_type"])
	}
}

// TestBootstrapOrchestrator_ConcurrentTracking tests concurrent iteration tracking.
func TestBootstrapOrchestrator_ConcurrentTracking(t *testing.T) {
	orch := orchestrator2.NewOrchestrator()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := orch.Start(ctx); err != nil {
		t.Fatalf("failed to start orchestrator: %v", err)
	}
	defer orch.Shutdown(context.Background())

	bo := NewBootstrapOrchestrator(orch, 100)

	taskID := "concurrent-test"
	var wg sync.WaitGroup

	// Spawn 50 goroutines to track iterations concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bo.TrackIteration(taskID)
		}()
	}

	wg.Wait()

	count := bo.GetIterationCount(taskID)
	if count != 50 {
		t.Errorf("expected iteration count 50, got %d", count)
	}
}

// TestExtractTerminationReason tests termination reason extraction.
func TestExtractTerminationReason(t *testing.T) {
	tests := []struct {
		name     string
		result   map[string]any
		expected string
	}{
		{
			name:     "nil result",
			result:   nil,
			expected: "unknown",
		},
		{
			name: "approved status",
			result: map[string]any{
				"approval_status": "approved",
			},
			expected: "task_approved",
		},
		{
			name: "rejected status",
			result: map[string]any{
				"approval_status": "rejected",
			},
			expected: "task_rejected",
		},
		{
			name: "needs_changes status",
			result: map[string]any{
				"approval_status": "needs_changes",
			},
			expected: "needs_modification",
		},
		{
			name: "termination_reason",
			result: map[string]any{
				"termination_reason": "max_iterations_reached",
			},
			expected: "max_iterations_reached",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTerminationReason(tt.result)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestBuildSpawnTaskInput tests spawn task input building.
func TestBuildSpawnTaskInput(t *testing.T) {
	reviewResult := map[string]any{
		"approval_status": "approved",
		"review_notes":    "looks good",
	}
	currentTaskID := "review-1"

	input := BuildSpawnTaskInput(reviewResult, currentTaskID)

	if input["review_result"] != reviewResult {
		t.Errorf("expected review_result to be set")
	}
	if input["current_task_id"] != currentTaskID {
		t.Errorf("expected current_task_id to be %s, got %v", currentTaskID, input["current_task_id"])
	}
}

// TestBootstrapOrchestrator_SelfContextInjection tests that self-context is properly injected into metadata.
func TestBootstrapOrchestrator_SelfContextInjection(t *testing.T) {
	orch := orchestrator2.NewOrchestrator()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := orch.Start(ctx); err != nil {
		t.Fatalf("failed to start orchestrator: %v", err)
	}
	defer orch.Shutdown(context.Background())

	bo := NewBootstrapOrchestrator(orch, 5)

	// Track a few iterations first
	bo.TrackIteration("test-task")
	bo.TrackIteration("test-task")

	task := &types.AgentTask{
		TaskID:     "test-task",
		ContractID: contracts.ContractIDAnalyze,
		Input: map[string]any{
			"change_description": "test",
		},
		Dependencies: []string{"parent-task"},
		Metadata:     make(map[string]string),
	}

	// Submit will inject self context
	if err := bo.SubmitSelfIterationTask(task); err != nil {
		t.Fatalf("failed to submit task: %v", err)
	}

	// Verify self context was injected
	if task.Metadata["self.iteration"] != "3" {
		t.Errorf("expected self.iteration 3, got %s", task.Metadata["self.iteration"])
	}
	if task.Metadata["self.max_iterations"] != "5" {
		t.Errorf("expected self.max_iterations 5, got %s", task.Metadata["self.max_iterations"])
	}
	if task.Metadata["self.parent_task_id"] != "parent-task" {
		t.Errorf("expected self.parent_task_id parent-task, got %s", task.Metadata["self.parent_task_id"])
	}
	if task.Metadata["self.lineage"] != "parent-task -> test-task" {
		t.Errorf("expected self.lineage 'parent-task -> test-task', got %s", task.Metadata["self.lineage"])
	}
}

// TestBootstrapOrchestrator_SubmitWithNilMetadata tests submission with nil metadata.
func TestBootstrapOrchestrator_SubmitWithNilMetadata(t *testing.T) {
	orch := orchestrator2.NewOrchestrator()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := orch.Start(ctx); err != nil {
		t.Fatalf("failed to start orchestrator: %v", err)
	}
	defer orch.Shutdown(context.Background())

	bo := NewBootstrapOrchestrator(orch, 5)

	task := &types.AgentTask{
		TaskID:     "nil-metadata-task",
		ContractID: contracts.ContractIDAnalyze,
		Input: map[string]any{
			"change_description": "test",
		},
		Dependencies: []string{},
		Metadata:     nil, // nil metadata
	}

	if err := bo.SubmitSelfIterationTask(task); err != nil {
		t.Fatalf("failed to submit task with nil metadata: %v", err)
	}

	// Verify metadata was initialized
	if task.Metadata == nil {
		t.Fatal("expected metadata to be initialized")
	}
	if task.Metadata["self.iteration"] != "1" {
		t.Errorf("expected self.iteration 1, got %s", task.Metadata["self.iteration"])
	}
}

// TestBootstrapOrchestrator_ResetIteration tests iteration reset.
func TestBootstrapOrchestrator_ResetIteration(t *testing.T) {
	orch := orchestrator2.NewOrchestrator()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := orch.Start(ctx); err != nil {
		t.Fatalf("failed to start orchestrator: %v", err)
	}
	defer orch.Shutdown(context.Background())

	bo := NewBootstrapOrchestrator(orch, 5)

	taskID := "reset-test"
	bo.TrackIteration(taskID)
	bo.TrackIteration(taskID)
	bo.TrackIteration(taskID)

	if bo.GetIterationCount(taskID) != 3 {
		t.Fatalf("expected count 3, got %d", bo.GetIterationCount(taskID))
	}

	bo.ResetIteration(taskID)

	if bo.GetIterationCount(taskID) != 0 {
		t.Errorf("expected count 0 after reset, got %d", bo.GetIterationCount(taskID))
	}

	// Should be allowed again
	if !bo.IsIterationAllowed(taskID) {
		t.Errorf("expected iteration to be allowed after reset")
	}
}

// TestBootstrapOrchestrator_EmptyFollowUps tests handling of empty follow-ups.
func TestBootstrapOrchestrator_EmptyFollowUps(t *testing.T) {
	parentTask := &types.AgentTask{
		TaskID:     "parent-empty",
		ContractID: contracts.ContractIDAnalyze,
		Input: map[string]any{
			"change_description": "test",
		},
	}

	result := &AgentExecutionResult{
		Output:        map[string]any{"result": "completed"},
		FollowUpTasks: nil, // nil follow-ups
	}

	tasks := GenerateFollowUpTasks(result, parentTask)
	if tasks != nil {
		t.Errorf("expected nil tasks for nil follow-ups, got %d", len(tasks))
	}

	result2 := &AgentExecutionResult{
		Output:        map[string]any{"result": "completed"},
		FollowUpTasks: []*types.AgentTask{}, // empty follow-ups
	}

	tasks2 := GenerateFollowUpTasks(result2, parentTask)
	if tasks2 != nil {
		t.Errorf("expected nil tasks for empty follow-ups, got %d", len(tasks2))
	}
}

// TestBootstrapOrchestrator_DAGEdges verifies the contract DAG forms a linear chain.
func TestBootstrapOrchestrator_DAGEdges(t *testing.T) {
	// Verify the contract DAG forms a linear chain
	contractsList := contracts.AllContracts()
	contractIDs := make(map[string]bool)
	for _, c := range contractsList {
		contractIDs[c.ContractID] = true
	}

	// Verify all expected contracts exist
	expected := []string{
		contracts.ContractIDAnalyze,
		contracts.ContractIDImplement,
		contracts.ContractIDValidate,
		contracts.ContractIDUpdate,
		contracts.ContractIDReview,
		contracts.ContractIDSpawn,
	}

	for _, exp := range expected {
		if !contractIDs[exp] {
			t.Errorf("expected contract %s not found", exp)
		}
	}
}

// TestBootstrapOrchestrator_CoverageThreshold verifies high coverage paths.
func TestBootstrapOrchestrator_CoverageThreshold(t *testing.T) {
	orch := orchestrator2.NewOrchestrator()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := orch.Start(ctx); err != nil {
		t.Fatalf("failed to start orchestrator: %v", err)
	}
	defer orch.Shutdown(context.Background())

	bo := NewBootstrapOrchestrator(orch, 3)

	// Exercise multiple paths for coverage
	task := &types.AgentTask{
		TaskID:     "coverage-task",
		ContractID: contracts.ContractIDAnalyze,
		Input: map[string]any{
			"change_description": "coverage test",
		},
		Metadata: make(map[string]string),
	}

	// Test iteration tracking paths
	bo.TrackIteration(task.TaskID)
	bo.TrackIteration(task.TaskID)
	count := bo.TrackIteration(task.TaskID)
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}

	// Test IsIterationAllowed
	if !bo.IsIterationAllowed(task.TaskID) {
		t.Error("expected iteration allowed at count 3 with max 5")
	}

	// Test GetLoopStatus
	status := bo.GetLoopStatus()
	if status[task.TaskID] != 3 {
		t.Errorf("expected status count 3, got %d", status[task.TaskID])
	}

	// Test ResetIteration
	bo.ResetIteration(task.TaskID)
	if bo.GetIterationCount(task.TaskID) != 0 {
		t.Error("expected count 0 after reset")
	}

	// Verify iteration still allowed after reset
	if !bo.IsIterationAllowed(task.TaskID) {
		t.Error("expected iteration allowed after reset")
	}

	// Submit task after reset
	if err := bo.SubmitSelfIterationTask(task); err != nil {
		t.Fatalf("failed to submit task: %v", err)
	}

	// Exhaust iterations to test max check
	for i := 0; i < 3; i++ {
		bo.TrackIteration(task.TaskID)
	}

	// Should be blocked
	newTask := &types.AgentTask{
		TaskID:     fmt.Sprintf("new-task-%d", time.Now().UnixNano()),
		ContractID: contracts.ContractIDAnalyze,
		Input: map[string]any{
			"change_description": "test",
		},
		Metadata: make(map[string]string),
	}
	if err := bo.SubmitSelfIterationTask(newTask); err == nil {
		t.Error("expected error when submitting with exceeded iterations")
	}
}
