package actor

import (
	"context"
	"fmt"
	"time"

	"github.com/axis-cli/axis/internal/comm"
	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/tool"
)

// SpawnExecutor handles spawn requests by creating isolated Worker Actors.
type SpawnExecutor struct {
	provider    provider.ModelProvider
	tools       *tool.Registry
	router      *comm.Router
	WorkerScope []string
}

// SpawnExecutorConfig configures a SpawnExecutor.
type SpawnExecutorConfig struct {
	Provider    provider.ModelProvider
	Tools       *tool.Registry
	Router      *comm.Router
	WorkerScope []string
}

// NewSpawnExecutor creates a SpawnExecutor.
func NewSpawnExecutor(cfg SpawnExecutorConfig) *SpawnExecutor {
	return &SpawnExecutor{
		provider:    cfg.Provider,
		tools:       cfg.Tools,
		router:      cfg.Router,
		WorkerScope: cfg.WorkerScope,
	}
}

// SpawnRequest represents a parsed spawn tool result.
type SpawnRequest struct {
	TaskID    string
	Prompt    string
	Isolation string
	ParentID  string
	MessageID string
}

// Execute creates a Worker Actor, runs the task, and sends result back to parent.
func (se *SpawnExecutor) Execute(ctx context.Context, req SpawnRequest) error {
	scope := se.WorkerScope
	if scope == nil {
		scope = se.defaultWorkerScope()
	}

	worker := NewLLMAdapter(LLMAdapterConfig{
		ID:       fmt.Sprintf("worker-%s", req.TaskID),
		Provider: se.provider,
		Tools:    se.tools,
		Scope:    scope,
	})

	se.router.Register(worker)
	defer se.router.Unregister(worker.ID())

	taskMsg := comm.Message{
		ID:        fmt.Sprintf("spawn-%s-%d", req.TaskID, time.Now().UnixNano()),
		From:      req.ParentID,
		To:        worker.ID(),
		Type:      comm.MsgTask,
		Payload:   map[string]any{"prompt": req.Prompt},
		Timestamp: time.Now(),
	}

	if err := worker.Receive(ctx, taskMsg); err != nil {
		return se.sendResult(ctx, req, map[string]any{"error": err.Error()})
	}

	result, ok := worker.GetResult(taskMsg.ID)
	if !ok {
		return se.sendResult(ctx, req, map[string]any{"error": "worker produced no result"})
	}

	return se.sendResult(ctx, req, result.Payload)
}

func (se *SpawnExecutor) sendResult(ctx context.Context, req SpawnRequest, payload map[string]any) error {
	resultMsg := comm.Message{
		ID:        fmt.Sprintf("result-%s-%d", req.TaskID, time.Now().UnixNano()),
		From:      fmt.Sprintf("worker-%s", req.TaskID),
		To:        req.ParentID,
		Type:      comm.MsgResult,
		Payload:   payload,
		Timestamp: time.Now(),
		ReplyTo:   req.MessageID,
	}
	return se.router.Send(ctx, resultMsg)
}

func (se *SpawnExecutor) defaultWorkerScope() []string {
	all := se.tools.List()
	var scope []string
	for _, td := range all {
		if td.Name != "spawn" {
			scope = append(scope, td.Name)
		}
	}
	return scope
}
