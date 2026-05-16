package actor

import (
	"context"
	"testing"

	"github.com/axis-cli/axis/internal/comm"
	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/tool"
	"github.com/axis-cli/axis/internal/types"
)

type mockProvider struct {
	output map[string]any
}

func (m *mockProvider) Execute(ctx context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	return &provider.ModelResponse{Output: m.output}, nil
}

// mockToolProvider is reserved for multi-turn tool-use tests.
var _ provider.ModelProvider = (*mockToolProvider)(nil)

type mockToolProvider struct {
	callCount int
}

func (m *mockToolProvider) Execute(ctx context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	m.callCount++
	if m.callCount == 1 && len(req.History) == 0 {
		return &provider.ModelResponse{
			ToolCalls: []types.ToolCall{{ID: "tc1", Name: "file_read", Input: map[string]any{"path": "test.txt"}}},
		}, nil
	}
	return &provider.ModelResponse{Output: map[string]any{"text": "task completed"}}, nil
}

func TestLLMAdapter_BasicExecution(t *testing.T) {
	mock := &mockProvider{output: map[string]any{"text": "done"}}
	adapter := NewLLMAdapter(LLMAdapterConfig{
		ID: "agent-1", Provider: mock, Tools: tool.NewRegistry(),
	})

	msg := comm.Message{
		ID: "msg-1", From: "leader", To: "agent-1",
		Type: comm.MsgTask, Payload: map[string]any{"prompt": "do something"},
	}
	if err := adapter.Receive(context.Background(), msg); err != nil {
		t.Fatal(err)
	}

	result, ok := adapter.GetResult("msg-1")
	if !ok {
		t.Fatal("no result")
	}
	if result.Payload["text"] != "done" {
		t.Errorf("payload = %v", result.Payload)
	}
	if result.Type != comm.MsgResult {
		t.Errorf("type = %v", result.Type)
	}
}

func TestLLMAdapter_ScopedTools(t *testing.T) {
	reg := tool.NewRegistry()
	reg.Register(tool.NewCompactTool(), nil)

	adapter := NewLLMAdapter(LLMAdapterConfig{
		ID: "worker-1", Provider: &mockProvider{output: map[string]any{"text": "ok"}},
		Tools: reg, Scope: []string{"compact"},
	})

	defs := adapter.scopedTools()
	if len(defs) != 1 || defs[0].Name != "compact" {
		t.Errorf("scoped tools = %v", defs)
	}
	if adapter.isAllowed("bash") {
		t.Error("bash should not be allowed")
	}
	if !adapter.isAllowed("compact") {
		t.Error("compact should be allowed")
	}
}

func TestSpawnExecutor_LeaderWorkerFlow(t *testing.T) {
	mock := &mockProvider{output: map[string]any{"text": "subtask result"}}
	reg := tool.NewRegistry()
	mb := comm.NewMailbox(t.TempDir())
	router := comm.NewRouter(mb)
	se := NewSpawnExecutor(SpawnExecutorConfig{Provider: mock, Tools: reg, Router: router})
	result, err := se.Execute(context.Background(), SpawnRequest{TaskID: "sub-1", Prompt: "analyze this file", Isolation: "full", ParentID: "leader", MessageID: "orig-msg-1"})
	if err != nil { t.Fatal(err) }
	if result["text"] != "subtask result" { t.Errorf("payload = %v", result) }
}
func TestSpawnExecutor_WorkerNoSpawnPermission(t *testing.T) {
	mock := &mockProvider{output: map[string]any{"text": "ok"}}
	reg := tool.NewRegistry()
	reg.Register(tool.NewSpawnTool(), nil)
	reg.Register(tool.NewCompactTool(), nil)

	mb := comm.NewMailbox(t.TempDir())
	router := comm.NewRouter(mb)

	se := NewSpawnExecutor(SpawnExecutorConfig{
		Provider: mock, Tools: reg, Router: router,
	})

	scope := se.defaultWorkerScope()
	for _, name := range scope {
		if name == "spawn" {
			t.Error("worker scope should not include spawn")
		}
	}
	found := false
	for _, name := range scope {
		if name == "compact" {
			found = true
		}
	}
	if !found {
		t.Error("worker scope should include compact")
	}
}
