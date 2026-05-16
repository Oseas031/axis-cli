package actor

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/axis-cli/axis/internal/comm"
	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/tool"
	"github.com/axis-cli/axis/internal/types"
)

// LLMAdapter wraps an LLM provider as an Actor.
type LLMAdapter struct {
	id       string
	provider provider.ModelProvider
	tools    *tool.Registry
	scope    []string
	status   ActorStatus
	mu       sync.Mutex
	results  map[string]comm.Message
	maxTurns int
}

// LLMAdapterConfig configures an LLMAdapter.
type LLMAdapterConfig struct {
	ID       string
	Provider provider.ModelProvider
	Tools    *tool.Registry
	Scope    []string
	MaxTurns int
}

// NewLLMAdapter creates an Actor backed by an LLM provider.
func NewLLMAdapter(cfg LLMAdapterConfig) *LLMAdapter {
	return &LLMAdapter{
		id:       cfg.ID,
		provider: cfg.Provider,
		tools:    cfg.Tools,
		scope:    cfg.Scope,
		status:   ActorReady,
		results:  make(map[string]comm.Message),
		maxTurns: cfg.MaxTurns,
	}
}

func (a *LLMAdapter) ID() string { return a.id }

func (a *LLMAdapter) Status() ActorStatus {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.status
}

func (a *LLMAdapter) CommStatus() comm.ActorStatus {
	return comm.ActorStatus(a.Status())
}

func (a *LLMAdapter) Receive(ctx context.Context, msg comm.Message) error {
	switch msg.Type {
	case comm.MsgTask:
		return a.executeTask(ctx, msg)
	default:
		return nil
	}
}

func (a *LLMAdapter) GetResult(msgID string) (comm.Message, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	r, ok := a.results[msgID]
	return r, ok
}

func (a *LLMAdapter) executeTask(ctx context.Context, msg comm.Message) error {
	a.mu.Lock()
	a.status = ActorBusy
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		a.status = ActorReady
		a.mu.Unlock()
	}()

	prompt, _ := msg.Payload["prompt"].(string)
	if prompt == "" {
		prompt, _ = msg.Payload["message"].(string)
	}

	toolDefs := a.scopedTools()
	var history []types.ModelMessage
	maxTurns := a.maxTurns
	if maxTurns <= 0 {
		maxTurns = 15
	}

	for turn := 0; turn < maxTurns; turn++ {
		req := &provider.ModelRequest{
			Input:   map[string]any{"message": prompt},
			Tools:   toolDefs,
			History: history,
		}
		resp, err := a.provider.Execute(ctx, req)
		if err != nil {
			a.storeResult(msg, map[string]any{"error": err.Error()})
			return err
		}

		if len(resp.ToolCalls) > 0 {
			history = append(history, types.ModelMessage{Role: "assistant", ToolCalls: resp.ToolCalls})
			for _, tc := range resp.ToolCalls {
				t, ok := a.tools.Get(tc.Name)
				if !ok || !a.isAllowed(tc.Name) {
					history = append(history, types.ModelMessage{
						Role: "tool", ToolCallID: tc.ID,
						Content: fmt.Sprintf("error: tool %s not available", tc.Name),
					})
					continue
				}
				result, execErr := t.Execute(ctx, tc.Input)
				if execErr != nil {
					history = append(history, types.ModelMessage{
						Role: "tool", ToolCallID: tc.ID,
						Content: fmt.Sprintf("error: %v", execErr),
					})
				} else {
					data, _ := json.Marshal(result)
					history = append(history, types.ModelMessage{
						Role: "tool", ToolCallID: tc.ID, Content: string(data),
					})
				}
			}
			continue
		}

		a.storeResult(msg, resp.Output)
		return nil
	}

	a.storeResult(msg, map[string]any{"error": "max turns exceeded"})
	return nil
}

func (a *LLMAdapter) scopedTools() []types.ToolDefinition {
	all := a.tools.List()
	if a.scope == nil {
		return all
	}
	allowed := make(map[string]bool, len(a.scope))
	for _, name := range a.scope {
		allowed[name] = true
	}
	var scoped []types.ToolDefinition
	for _, td := range all {
		if allowed[td.Name] {
			scoped = append(scoped, td)
		}
	}
	return scoped
}

func (a *LLMAdapter) isAllowed(name string) bool {
	if a.scope == nil {
		return true
	}
	for _, s := range a.scope {
		if s == name {
			return true
		}
	}
	return false
}

func (a *LLMAdapter) storeResult(origMsg comm.Message, output map[string]any) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.results[origMsg.ID] = comm.Message{
		ID:      origMsg.ID + "-result",
		From:    a.id,
		To:      origMsg.From,
		Type:    comm.MsgResult,
		Payload: output,
		ReplyTo: origMsg.ID,
	}
}
