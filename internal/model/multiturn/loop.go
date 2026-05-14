// Package multiturn provides a shared multi-turn tool-calling loop.
package multiturn

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/tool"
	"github.com/axis-cli/axis/internal/types"
)

// Compactor compacts message history to manage context window.
type Compactor interface {
	Compact(ctx context.Context, history []types.ModelMessage) []types.ModelMessage
}

// LoopConfig configures the multi-turn loop.
type LoopConfig struct {
	Provider       provider.ModelProvider
	Tools          *tool.Registry
	MaxIterations  int
	MaxErrors      int // circuit breaker threshold
	Compactor      Compactor
	TurnTimeout    time.Duration
	OnToolExecuted func(toolName string, result map[string]any, err error) // optional hook
}

// LoopResult is the outcome of a multi-turn loop execution.
type LoopResult struct {
	Output  map[string]any
	History []types.ModelMessage
	Error   string
}

// Run executes the multi-turn tool-calling loop.
func Run(ctx context.Context, cfg LoopConfig, req *provider.ModelRequest) (*LoopResult, error) {
	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 20
	}
	if cfg.MaxErrors <= 0 {
		cfg.MaxErrors = 5
	}
	if cfg.TurnTimeout <= 0 {
		cfg.TurnTimeout = 45 * time.Second
	}

	var history []types.ModelMessage
	consecutiveErrors := 0
	var lastToolOutputs [3]string // ring buffer for runaway detection
	outputIdx := 0
	outputCount := 0

	for iter := 0; iter < cfg.MaxIterations; iter++ {
		select {
		case <-ctx.Done():
			return &LoopResult{History: history, Error: fmt.Sprintf("context cancelled: %v", ctx.Err())}, ctx.Err()
		default:
		}

		req.History = history

		turnCtx, turnCancel := context.WithTimeout(ctx, cfg.TurnTimeout)
		resp, err := cfg.Provider.Execute(turnCtx, req)
		turnCancel()
		if err != nil {
			return &LoopResult{History: history, Error: fmt.Sprintf("provider error: %v", err)}, err
		}

		// No tool calls — return output
		if len(resp.ToolCalls) == 0 {
			return &LoopResult{Output: resp.Output, History: history}, nil
		}

		// Append assistant message with tool calls
		history = append(history, types.ModelMessage{
			Role:      "assistant",
			ToolCalls: resp.ToolCalls,
		})

		// Execute each tool call
		for _, tc := range resp.ToolCalls {
			toolImpl, ok := cfg.Tools.Get(tc.Name)
			if !ok {
				errMsg := fmt.Sprintf("tool %s not found", tc.Name)
				history = append(history, types.ModelMessage{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    errMsg,
				})
				if cfg.OnToolExecuted != nil {
					cfg.OnToolExecuted(tc.Name, nil, fmt.Errorf("%s", errMsg))
				}
				consecutiveErrors++
				if consecutiveErrors >= cfg.MaxErrors {
					return &LoopResult{History: history, Error: fmt.Sprintf("circuit breaker: %d consecutive errors", consecutiveErrors)}, nil
				}
				continue
			}

			toolCtx, toolCancel := context.WithTimeout(ctx, cfg.TurnTimeout)
			result, execErr := toolImpl.Execute(toolCtx, tc.Input)
			toolCancel()
			if execErr != nil {
				history = append(history, types.ModelMessage{
					Role:       "tool",
					ToolCallID: tc.ID,
					Content:    fmt.Sprintf("error: %v", execErr),
				})
				if cfg.OnToolExecuted != nil {
					cfg.OnToolExecuted(tc.Name, nil, execErr)
				}
				consecutiveErrors++
				if consecutiveErrors >= cfg.MaxErrors {
					return &LoopResult{History: history, Error: fmt.Sprintf("circuit breaker: %d consecutive errors", consecutiveErrors)}, nil
				}
				continue
			}

			consecutiveErrors = 0
			content, _ := json.Marshal(result)
			contentStr := string(content)

			// Runaway detection: if last 3 tool outputs are identical, abort
			lastToolOutputs[outputIdx%3] = contentStr
			outputIdx++
			outputCount++
			if outputCount >= 3 && lastToolOutputs[0] == lastToolOutputs[1] && lastToolOutputs[1] == lastToolOutputs[2] {
				return &LoopResult{History: history, Error: "runaway detected: last 3 tool outputs identical"}, nil
			}

			history = append(history, types.ModelMessage{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    contentStr,
			})
			if cfg.OnToolExecuted != nil {
				cfg.OnToolExecuted(tc.Name, result, nil)
			}
		}

		// Compact history
		if cfg.Compactor != nil {
			history = cfg.Compactor.Compact(ctx, history)
		}
	}

	return &LoopResult{History: history, Error: fmt.Sprintf("iteration budget exhausted (%d turns)", cfg.MaxIterations)}, nil
}
