package executor

import (
	"context"
	"strings"
	"testing"

	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/tool"
	"github.com/axis-cli/axis/internal/types"
)

// nagTestProvider returns tool calls for the first N turns, then a final output.
// It records the SystemPrompt on each call.
type nagTestProvider struct {
	toolTurns int
	callCount int
	prompts   []string
}

func (p *nagTestProvider) Execute(ctx context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	p.prompts = append(p.prompts, req.SystemPrompt)
	p.callCount++
	if p.callCount <= p.toolTurns {
		return &provider.ModelResponse{
			ToolCalls: []types.ToolCall{
				{ID: "call-1", Name: "bash", Input: map[string]any{"cmd": "echo hi"}},
			},
		}, nil
	}
	return &provider.ModelResponse{Output: map[string]any{"status": "done"}}, nil
}

// noopTool is a minimal tool implementation for testing.
type noopTool struct{}

func (n *noopTool) Name() string                  { return "bash" }
func (n *noopTool) Schema() types.ToolDefinition  { return types.ToolDefinition{Name: "bash"} }
func (n *noopTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	return map[string]any{"ok": true}, nil
}

func TestExecuteMultiTurn_NagAfter5Turns(t *testing.T) {
	exec := NewContractExecutor()

	mp := &nagTestProvider{toolTurns: 6}
	exec.SetProvider(mp)

	tr := tool.NewRegistry()
	tr.Register(&noopTool{}, nil)
	exec.SetToolRegistry(tr)

	contract := &types.AgentContract{
		ContractID:  "nag-test",
		InputSchema: &types.InputSchema{Fields: []types.FieldDef{{Name: "msg", Type: types.FieldTypeString}}},
	}
	exec.RegisterContract(contract)

	_, err := exec.Execute(context.Background(), "nag-test", map[string]any{"msg": "go"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// The 6th provider call (index 5) should have the reminder in SystemPrompt.
	// Turns 1-5 increment turnsSinceProgress to 5, so before turn 6 the nag fires.
	if len(mp.prompts) < 6 {
		t.Fatalf("Expected at least 6 provider calls, got %d", len(mp.prompts))
	}

	reminder := "Reminder: you have not updated task progress in the last 5 turns. Consider checkpointing or recording progress."
	if !strings.Contains(mp.prompts[5], reminder) {
		t.Errorf("Expected reminder in 6th call SystemPrompt, got: %q", mp.prompts[5])
	}

	// First 5 calls should NOT contain the reminder
	for i := 0; i < 5; i++ {
		if strings.Contains(mp.prompts[i], reminder) {
			t.Errorf("Call %d should not contain reminder, got: %q", i+1, mp.prompts[i])
		}
	}
}
