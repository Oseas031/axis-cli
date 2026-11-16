package executor

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/tool"
	"github.com/axis-cli/axis/internal/types"
)

type dsmlLeakProvider struct {
	calls int
	fail  bool
}

func (p *dsmlLeakProvider) Name() string { return "dsml-leak-mock" }

func (p *dsmlLeakProvider) Execute(ctx context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	p.calls++
	if p.fail || p.calls == 1 {
		return &provider.ModelResponse{
			Output: map[string]any{
				"text": "<｜DSML｜ calls>\n<｜DSML｜ invoke name=\"bash\">\n</｜DSML｜ invoke>\n</｜DSML｜ calls>",
			},
		}, nil
	}
	return &provider.ModelResponse{
		Output: map[string]any{
			"text": "recovered final answer",
		},
	}, nil
}

func TestEmitLiveEvent_UTF8Truncation(t *testing.T) {
	longChinese := strings.Repeat("中文测试字符", 60) // 360 runes > 220
	var captured string
	ctx := WithLiveEventObserver(context.Background(), func(eventType, message string) {
		captured = message
	})
	emitLiveEvent(ctx, "agent_thinking", longChinese)

	if !utf8.ValidString(captured) {
		t.Fatalf("expected valid UTF-8 after truncation, got invalid bytes")
	}
	if !strings.HasSuffix(captured, "...") {
		t.Fatalf("expected ellipsis suffix, got %q", captured)
	}
}

func TestExecuteMultiTurn_DSMLRecoveryAndFailGuard(t *testing.T) {
	tr := tool.NewRegistry()
	_ = tr.Register(&fixedTool{name: "bash", result: map[string]any{"ok": true}}, nil)
	p := &dsmlLeakProvider{}
	exec := NewContractExecutorWithConfig(ExecutorConfig{
		Provider:     p,
		ToolRegistry: tr,
		MaxTurns:     5,
	})
	if err := exec.RegisterContract(&types.AgentContract{
		ContractID: "test-dsml",
	}); err != nil {
		t.Fatalf("RegisterContract: %v", err)
	}

	res, err := exec.Execute(context.Background(), "test-dsml", map[string]any{"prompt": "hi"})
	if err != nil {
		t.Fatalf("expected recovery after single DSML leak, got err: %v", err)
	}
	if got, _ := res.Output["text"].(string); got != "recovered final answer" {
		t.Fatalf("unexpected output: %#v", res.Output)
	}

	failProvider := &dsmlLeakProvider{fail: true}
	execFail := NewContractExecutorWithConfig(ExecutorConfig{
		Provider:     failProvider,
		ToolRegistry: tr,
		MaxTurns:     6,
	})
	_ = execFail.RegisterContract(&types.AgentContract{
		ContractID: "test-dsml",
	})
	_, err = execFail.Execute(context.Background(), "test-dsml", map[string]any{"prompt": "hi"})
	if err == nil || !strings.Contains(err.Error(), "DSML") {
		t.Fatalf("expected DSML error on repeated leaks, got: %v", err)
	}
}
