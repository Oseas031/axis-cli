package dispatcher

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/agent"
	"github.com/axis-cli/axis/internal/contextpack"
	contractexec "github.com/axis-cli/axis/internal/contract/executor"
	humanexec "github.com/axis-cli/axis/internal/human/executor"
	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/types"
)

func TestDispatcher_NewDispatcher(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := NewDispatcher(contractExec, humanExec)

	if dispatch == nil {
		t.Fatal("NewDispatcher should return non-nil")
	}
}

func TestDispatcher_Dispatch(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := NewDispatcher(contractExec, humanExec)

	// Register a contract
	contract := &types.AgentContract{
		ContractID: "test-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{
					Name:     "name",
					Type:     types.FieldTypeString,
					Required: true,
				},
			},
		},
	}
	contractExec.RegisterContract(contract)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "test-contract",
		Input:      map[string]any{"name": "test"},
	}

	ctx := context.Background()
	result, err := dispatch.Dispatch(ctx, task)

	if err != nil {
		t.Errorf("Dispatch should succeed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.TaskID != task.TaskID {
		t.Errorf("Expected task ID %s, got %s", task.TaskID, result.TaskID)
	}
}

func TestDispatcher_DispatchInvalidInput(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := NewDispatcher(contractExec, humanExec)

	// Register a contract with required field
	contract := &types.AgentContract{
		ContractID: "test-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{
					Name:     "name",
					Type:     types.FieldTypeString,
					Required: true,
				},
			},
		},
	}
	contractExec.RegisterContract(contract)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "test-contract",
		Input:      map[string]any{}, // Missing required field
	}

	ctx := context.Background()
	result, _ := dispatch.Dispatch(ctx, task)

	// Should return result with error status
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Status != types.TaskStatusFailed {
		t.Errorf("Expected status %s, got %s", types.TaskStatusFailed, result.Status)
	}
}

func TestDispatcher_DispatchParentContextCancelled(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := NewDispatcher(contractExec, humanExec)

	contract := &types.AgentContract{
		ContractID: "test-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{{Name: "name", Type: types.FieldTypeString, Required: true}},
		},
	}
	contractExec.RegisterContract(contract)

	task := &types.AgentTask{
		TaskID:     "ctx-cancelled",
		ContractID: "test-contract",
		Input:      map[string]any{"name": "test"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := dispatch.Dispatch(ctx, task)
	if err == nil {
		t.Fatal("Dispatch should return error when parent context is cancelled")
	}
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	if result.Status != types.TaskStatusFailed {
		t.Errorf("Expected failed status, got %s", result.Status)
	}
}

func TestDispatcher_DispatchErrChan(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := NewDispatcher(contractExec, humanExec)

	// Set a provider so Execute goes through the full path
	contractExec.SetProvider(provider.NewMockModelProvider())

	// Contract whose output schema requires a field the mock won't provide
	contract := &types.AgentContract{
		ContractID: "err-chan",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{{Name: "x", Type: types.FieldTypeString, Required: false}},
		},
		OutputSchema: &types.OutputSchema{
			Fields: []types.FieldDef{{Name: "missing_field", Type: types.FieldTypeString, Required: true}},
		},
	}
	contractExec.RegisterContract(contract)

	task := &types.AgentTask{
		TaskID:     "err-chan-task",
		ContractID: "err-chan",
		Input:      map[string]any{"x": "y"},
	}

	ctx := context.Background()
	result, err := dispatch.Dispatch(ctx, task)
	if err == nil {
		t.Fatal("Dispatch should return error when executeTask fails output validation")
	}
	if result == nil {
		t.Fatal("Result should not be nil")
	}
	if result.Status != types.TaskStatusFailed {
		t.Errorf("Expected failed status, got %s", result.Status)
	}
}

func TestDispatcher_DispatchTimeout(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()

	// Create dispatcher with very short timeout
	dispatch := NewDispatcher(contractExec, humanExec)
	dispatch.timeout = 10 * time.Millisecond

	contract := &types.AgentContract{
		ContractID: "test-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{
				{
					Name:     "name",
					Type:     types.FieldTypeString,
					Required: true,
				},
			},
		},
	}
	contractExec.RegisterContract(contract)

	task := &types.AgentTask{
		TaskID:     "task-1",
		ContractID: "test-contract",
		Input:      map[string]any{"name": "test"},
	}

	ctx := context.Background()
	result, _ := dispatch.Dispatch(ctx, task)

	// Should still succeed quickly (task execution is fast)

	if result == nil {
		t.Error("Result should not be nil")
	}
}

func TestDispatcher_HumanExecutorRoute(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := NewDispatcher(contractExec, humanExec)
	dispatch.timeout = 500 * time.Millisecond

	task := &types.AgentTask{
		TaskID:     "human-task",
		ContractID: "any",
		Input:      map[string]any{"prompt": "hello"},
		Metadata:   map[string]string{types.TaskMetadataKeyExecutor: types.ExecutorTypeHuman},
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		humanExec.ResolveCall("human-task", map[string]any{"answer": "hi"}, nil)
	}()

	ctx := context.Background()
	result, err := dispatch.Dispatch(ctx, task)
	if err != nil {
		t.Fatalf("Human dispatch should succeed after resolution: %v", err)
	}
	if result.Status != types.TaskStatusCompleted {
		t.Errorf("Expected completed, got %s", result.Status)
	}
}

func TestDispatcher_HumanExecutorTimeout(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := NewDispatcher(contractExec, humanExec)
	dispatch.timeout = 100 * time.Millisecond

	task := &types.AgentTask{
		TaskID:     "human-timeout",
		ContractID: "any",
		Input:      map[string]any{"prompt": "hello"},
		Metadata:   map[string]string{types.TaskMetadataKeyExecutor: types.ExecutorTypeHuman},
	}

	ctx := context.Background()
	result, err := dispatch.Dispatch(ctx, task)
	if err == nil {
		t.Fatal("Human dispatch should timeout")
	}
	if result.Status != types.TaskStatusFailed {
		t.Errorf("Expected failed status, got %s", result.Status)
	}
}

func TestDispatcher_SetAgentExecutor(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)

	agentExec := agent.NewMockAgentExecutor(contractExec)
	dispatcher.SetAgentExecutor(agentExec)

	if dispatcher.agentExecutor == nil {
		t.Error("agentExecutor should be set after SetAgentExecutor")
	}
}

func TestDispatcher_Dispatch_AgentExecutor(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)

	agentExec := agent.NewMockAgentExecutor(contractExec)
	dispatcher.SetAgentExecutor(agentExec)

	task := &types.AgentTask{
		TaskID:     "agent-task-1",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "test input",
			"task_type": "default",
		},
		Metadata: map[string]string{
			"executor": "agent",
		},
		Status: types.TaskStatusPending,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := dispatcher.Dispatch(ctx, task)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.Status != types.TaskStatusCompleted {
		t.Errorf("Expected status completed, got %s", result.Status)
	}
}

func TestDispatcher_Dispatch_AgentExecutorNotConfigured(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)
	// Don't set agent executor

	task := &types.AgentTask{
		TaskID:     "agent-task-unconfigured",
		ContractID: "test-contract",
		Input: map[string]any{
			"input": "test input",
		},
		Metadata: map[string]string{
			"executor": "agent",
		},
		Status: types.TaskStatusPending,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := dispatcher.Dispatch(ctx, task)
	if err == nil {
		t.Error("Expected error when agent executor not configured")
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.Status != types.TaskStatusFailed {
		t.Errorf("Expected status failed, got %s", result.Status)
	}

	if result.Error == "" {
		t.Error("Error should be set when agent executor not configured")
	}
}

func TestDispatcher_Dispatch_AgentExecutorCodeGen(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)

	agentExec := agent.NewMockAgentExecutor(contractExec)
	dispatcher.SetAgentExecutor(agentExec)

	task := &types.AgentTask{
		TaskID:     "agent-code-gen",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "generate code",
			"task_type": "code_generation",
		},
		Metadata: map[string]string{
			"executor": "agent",
		},
		Status: types.TaskStatusPending,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := dispatcher.Dispatch(ctx, task)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.Status != types.TaskStatusCompleted {
		t.Errorf("Expected status completed, got %s", result.Status)
	}
}

func TestDispatcher_Dispatch_AgentExecutorDebugging(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)

	agentExec := agent.NewMockAgentExecutor(contractExec)
	dispatcher.SetAgentExecutor(agentExec)

	task := &types.AgentTask{
		TaskID:     "agent-debug",
		ContractID: "test-contract",
		Input: map[string]any{
			"input":     "debug issue",
			"task_type": "debugging",
		},
		Metadata: map[string]string{
			"executor": "agent",
		},
		Status: types.TaskStatusPending,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := dispatcher.Dispatch(ctx, task)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.Status != types.TaskStatusCompleted {
		t.Errorf("Expected status completed, got %s", result.Status)
	}
}

// detectingAgentExecutor records whether its Execute context was cancelled.
type detectingAgentExecutor struct {
	mu        sync.Mutex
	cancelled bool
}

func (e *detectingAgentExecutor) Execute(ctx context.Context, req *agent.AgentExecutionRequest) (*agent.AgentExecutionResult, error) {
	select {
	case <-ctx.Done():
		e.mu.Lock()
		e.cancelled = true
		e.mu.Unlock()
		return nil, ctx.Err()
	case <-time.After(300 * time.Millisecond):
		return &agent.AgentExecutionResult{Output: map[string]any{"status": "ok"}}, nil
	}
}
func (e *detectingAgentExecutor) WasCancelled() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.cancelled
}
func (e *detectingAgentExecutor) GetAutonomyLevel() agent.AutonomyLevel {
	return agent.AutonomyLevelLow
}

func TestDispatcher_Dispatch_AgentExecutorPropagatesCancellation(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)
	agentExec := &detectingAgentExecutor{}
	dispatcher.SetAgentExecutor(agentExec)
	dispatcher.timeout = 5 * time.Second

	task := &types.AgentTask{
		TaskID:     "agent-cancel",
		ContractID: "test-contract",
		Input:      map[string]any{"input": "test"},
		Metadata:   map[string]string{types.TaskMetadataKeyExecutor: types.ExecutorTypeAgent},
		Status:     types.TaskStatusPending,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, _ = dispatcher.Dispatch(ctx, task)
	// dispatch returns when timeoutCtx cancels (~50ms), but the real question is
	// whether the agent executor itself observed the cancellation.
	time.Sleep(400 * time.Millisecond)
	if !agentExec.WasCancelled() {
		t.Fatal("Agent executor should receive cancellable context, not Background")
	}
}

func TestDispatcher_Dispatch_HumanExecutorRespectsCancellation(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)
	dispatcher.timeout = 5 * time.Second

	task := &types.AgentTask{
		TaskID:     "human-cancel",
		ContractID: "any",
		Input:      map[string]any{"prompt": "hello"},
		Metadata:   map[string]string{types.TaskMetadataKeyExecutor: types.ExecutorTypeHuman},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	before := runtime.NumGoroutine()
	_, _ = dispatcher.Dispatch(ctx, task)
	// Give the internal polling loop a chance to observe ctx cancellation and exit.
	time.Sleep(300 * time.Millisecond)
	after := runtime.NumGoroutine()
	if after > before+2 {
		t.Fatalf("Human executor polling goroutine leaked: before=%d after=%d", before, after)
	}
}

func TestDispatcher_Dispatch_AgentExecutorReceivesContextSummary(t *testing.T) {
	contextpack.DefaultRegistry.Reset()
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)
	agentExec := &capturingAgentExecutor{}
	dispatcher.SetAgentExecutor(agentExec)

	task := &types.AgentTask{
		TaskID:     "agent-context-summary",
		ContractID: "default",
		Input:      map[string]any{"goal": "fix provider config"},
		Metadata:   map[string]string{types.TaskMetadataKeyExecutor: types.ExecutorTypeAgent},
		Status:     types.TaskStatusPending,
	}
	bundle, err := contextpack.NewAssembler().Assemble(task)
	if err != nil {
		t.Fatalf("assemble should succeed: %v", err)
	}
	artifact, err := contextpack.DefaultRegistry.Register(bundle)
	if err != nil {
		t.Fatalf("register should succeed: %v", err)
	}
	if err := contextpack.AttachReadinessMetadata(task, artifact); err != nil {
		t.Fatalf("attach readiness metadata should succeed: %v", err)
	}

	result, err := dispatcher.Dispatch(context.Background(), task)
	if err != nil {
		t.Fatalf("dispatch should succeed: %v", err)
	}
	if result.Status != types.TaskStatusCompleted {
		t.Fatalf("expected completed result, got %s", result.Status)
	}
	if agentExec.request == nil || agentExec.request.ContextSummary == nil {
		t.Fatal("expected context summary on agent execution request")
	}
	if agentExec.request.ContextSummary.ConsumptionMode != contextpack.ConsumptionModeSummary {
		t.Fatalf("expected summary consumption mode, got %+v", agentExec.request.ContextSummary)
	}
	if agentExec.request.ContextSummary.BundleID != artifact.BundleID {
		t.Fatalf("expected bundle id %s, got %+v", artifact.BundleID, agentExec.request.ContextSummary)
	}
}

func TestDispatcher_Dispatch_AgentExecutorMissingContextDoesNotBlock(t *testing.T) {
	contextpack.DefaultRegistry.Reset()
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)
	agentExec := &capturingAgentExecutor{}
	dispatcher.SetAgentExecutor(agentExec)

	task := &types.AgentTask{
		TaskID:     "agent-missing-context-summary",
		ContractID: "default",
		Input:      map[string]any{"goal": "fix provider config"},
		Metadata:   map[string]string{types.TaskMetadataKeyExecutor: types.ExecutorTypeAgent},
		Status:     types.TaskStatusPending,
	}

	result, err := dispatcher.Dispatch(context.Background(), task)
	if err != nil {
		t.Fatalf("dispatch should succeed without context readiness: %v", err)
	}
	if result.Status != types.TaskStatusCompleted {
		t.Fatalf("expected completed result, got %s", result.Status)
	}
	if agentExec.request == nil || agentExec.request.ContextSummary == nil {
		t.Fatal("expected context summary on agent execution request")
	}
	if agentExec.request.ContextSummary.Status != contextpack.PreflightStatusMissing {
		t.Fatalf("expected missing context status, got %+v", agentExec.request.ContextSummary)
	}
	if agentExec.request.ContextSummary.ConsumptionMode != contextpack.ConsumptionModeNone {
		t.Fatalf("expected none consumption mode, got %+v", agentExec.request.ContextSummary)
	}
}

func TestDispatcher_Dispatch_ContractPathDoesNotInjectContextSummary(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	modelProvider := &capturingModelProvider{}
	contractExec.SetProvider(modelProvider)
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)
	contract := &types.AgentContract{ContractID: "contract-context-boundary"}
	if err := contractExec.RegisterContract(contract); err != nil {
		t.Fatalf("register contract should succeed: %v", err)
	}

	task := &types.AgentTask{
		TaskID:     "contract-context-boundary-task",
		ContractID: "contract-context-boundary",
		Input:      map[string]any{"goal": "fix provider config"},
		Metadata: map[string]string{
			contextpack.MetadataBundleID: "ctx-example",
		},
		Status: types.TaskStatusPending,
	}

	result, err := dispatcher.Dispatch(context.Background(), task)
	if err != nil {
		t.Fatalf("dispatch should succeed: %v", err)
	}
	if result.Status != types.TaskStatusCompleted {
		t.Fatalf("expected completed result, got %s", result.Status)
	}
	if modelProvider.request == nil {
		t.Fatal("expected model provider request")
	}
	if _, ok := modelProvider.request.Input[contextpack.MetadataBundleID]; ok {
		t.Fatalf("contract path should not inject context metadata into provider input: %+v", modelProvider.request.Input)
	}
	if _, ok := modelProvider.request.Input["context_summary"]; ok {
		t.Fatalf("contract path should not inject context summary into provider input: %+v", modelProvider.request.Input)
	}
}

func TestDispatcher_Dispatch_AgentExecutorReceivesRequestedSources(t *testing.T) {
	contextpack.DefaultRegistry.Reset()
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)
	agentExec := &capturingAgentExecutor{}
	dispatcher.SetAgentExecutor(agentExec)

	task := &types.AgentTask{
		TaskID:     "agent-requested-sources",
		ContractID: "default",
		Input:      map[string]any{"goal": "fix provider config"},
		Metadata: map[string]string{
			types.TaskMetadataKeyExecutor:        types.ExecutorTypeAgent,
			contextpack.MetadataRequestedSources: "docs/specs/model-provider/requirements.md, docs/specs/missing.md",
		},
		Status: types.TaskStatusPending,
	}

	result, err := dispatcher.Dispatch(context.Background(), task)
	if err != nil {
		t.Fatalf("dispatch should succeed: %v", err)
	}
	if result.Status != types.TaskStatusCompleted {
		t.Fatalf("expected completed result, got %s", result.Status)
	}
	if agentExec.request == nil {
		t.Fatal("expected agent execution request")
	}
	if len(agentExec.request.RequestedSources) != 2 {
		t.Fatalf("expected 2 requested sources, got %+v", agentExec.request.RequestedSources)
	}
	if agentExec.request.RequestedSources[0] != "docs/specs/model-provider/requirements.md" {
		t.Fatalf("unexpected first requested source: %s", agentExec.request.RequestedSources[0])
	}
	if agentExec.request.RequestedSources[1] != "docs/specs/missing.md" {
		t.Fatalf("unexpected second requested source: %s", agentExec.request.RequestedSources[1])
	}
}

type capturingAgentExecutor struct {
	request *agent.AgentExecutionRequest
}

func (e *capturingAgentExecutor) Execute(ctx context.Context, req *agent.AgentExecutionRequest) (*agent.AgentExecutionResult, error) {
	e.request = req
	return &agent.AgentExecutionResult{Output: map[string]any{"status": "ok"}}, nil
}

func (e *capturingAgentExecutor) GetAutonomyLevel() agent.AutonomyLevel {
	return agent.AutonomyLevelLow
}

type capturingModelProvider struct {
	request *provider.ModelRequest
}

func (p *capturingModelProvider) Execute(ctx context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	p.request = req
	return &provider.ModelResponse{Output: map[string]any{"status": "ok"}}, nil
}


func TestDispatcher_AutonomyResolver(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatcher := NewDispatcher(contractExec, humanExec)
	agentExec := &capturingAgentExecutor{}
	dispatcher.SetAgentExecutor(agentExec)

	// Test 1: default resolver returns AutonomyLevelLow
	task := &types.AgentTask{
		TaskID:   "autonomy-default",
		Input:    map[string]any{"input": "test"},
		Metadata: map[string]string{types.TaskMetadataKeyExecutor: types.ExecutorTypeAgent},
		Status:   types.TaskStatusPending,
	}
	_, err := dispatcher.Dispatch(context.Background(), task)
	if err != nil {
		t.Fatalf("dispatch should succeed: %v", err)
	}
	if agentExec.request.Autonomy != agent.AutonomyLevelLow {
		t.Fatalf("expected AutonomyLevelLow, got %v", agentExec.request.Autonomy)
	}

	// Test 2: metadata autonomy_level=high
	task2 := &types.AgentTask{
		TaskID: "autonomy-metadata",
		Input:  map[string]any{"input": "test"},
		Metadata: map[string]string{
			types.TaskMetadataKeyExecutor: types.ExecutorTypeAgent,
			"autonomy_level":             "high",
		},
		Status: types.TaskStatusPending,
	}
	_, err = dispatcher.Dispatch(context.Background(), task2)
	if err != nil {
		t.Fatalf("dispatch should succeed: %v", err)
	}
	if agentExec.request.Autonomy != agent.AutonomyLevelHigh {
		t.Fatalf("expected AutonomyLevelHigh, got %v", agentExec.request.Autonomy)
	}

	// Test 3: custom resolver via WithAutonomyResolver
	dispatcher.WithAutonomyResolver(func(t *types.AgentTask) agent.AutonomyLevel {
		return agent.AutonomyLevelFull
	})
	task3 := &types.AgentTask{
		TaskID:   "autonomy-custom",
		Input:    map[string]any{"input": "test"},
		Metadata: map[string]string{types.TaskMetadataKeyExecutor: types.ExecutorTypeAgent},
		Status:   types.TaskStatusPending,
	}
	_, err = dispatcher.Dispatch(context.Background(), task3)
	if err != nil {
		t.Fatalf("dispatch should succeed: %v", err)
	}
	if agentExec.request.Autonomy != agent.AutonomyLevelFull {
		t.Fatalf("expected AutonomyLevelFull, got %v", agentExec.request.Autonomy)
	}
}


func TestDispatcher_AuditLog(t *testing.T) {
	contractExec := contractexec.NewContractExecutor()
	humanExec := humanexec.NewHumanExecutor()
	dispatch := NewDispatcher(contractExec, humanExec)

	contract := &types.AgentContract{
		ContractID: "audit-contract",
		InputSchema: &types.InputSchema{
			Fields: []types.FieldDef{{Name: "name", Type: types.FieldTypeString, Required: true}},
		},
	}
	contractExec.RegisterContract(contract)

	task := &types.AgentTask{
		TaskID:     "audit-task-1",
		ContractID: "audit-contract",
		Input:      map[string]any{"name": "test"},
	}

	ctx := context.Background()
	_, err := dispatch.Dispatch(ctx, task)
	if err != nil {
		t.Fatalf("Dispatch should succeed: %v", err)
	}

	entries := dispatch.AuditLog()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 audit entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.TaskID != "audit-task-1" {
		t.Errorf("Expected TaskID audit-task-1, got %s", entry.TaskID)
	}
	if entry.ExecutorType != "contract" {
		t.Errorf("Expected ExecutorType contract, got %s", entry.ExecutorType)
	}
	if entry.Status != "completed" {
		t.Errorf("Expected Status completed, got %s", entry.Status)
	}
	if entry.Duration < 0 {
		t.Errorf("Expected non-negative Duration, got %v", entry.Duration)
	}
	if entry.Timestamp.IsZero() {
		t.Error("Expected non-zero Timestamp")
	}
	if entry.Error != "" {
		t.Errorf("Expected empty Error, got %s", entry.Error)
	}
}
