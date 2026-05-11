package intent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

const ParserModeDeterministic = "deterministic"

type Request struct {
	Prompt     string
	ContractID string
	TaskID     string
}

type Result struct {
	Task       *types.AgentTask
	Confidence string
	Notes      []string
}

type Parser interface {
	Parse(ctx context.Context, req Request) (*Result, error)
}

type DeterministicParser struct{}

func NewDeterministicParser() *DeterministicParser {
	return &DeterministicParser{}
}

func (p *DeterministicParser) Parse(ctx context.Context, req Request) (*Result, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	contractID := strings.TrimSpace(req.ContractID)
	if contractID == "" {
		contractID = "default"
	}
	taskID := strings.TrimSpace(req.TaskID)
	if taskID == "" {
		taskID = "ask-" + time.Now().Format("20060102-150405")
	}
	task := &types.AgentTask{
		TaskID:     taskID,
		ContractID: contractID,
		Input: map[string]any{
			"message": prompt,
			"goal":    prompt,
		},
		Status:    types.TaskStatusPending,
		CreatedAt: time.Now(),
		Metadata: map[string]string{
			// Legacy un-namespaced keys for backward compatibility
			"source":          "natural_language",
			"original_prompt": prompt,
			"parser_mode":     ParserModeDeterministic,
			// New namespaced keys per metadata-key-conventions.md
			"intent.source":          "natural_language",
			"intent.original_prompt": prompt,
			"intent.parser_mode":     ParserModeDeterministic,
		},
	}
	return &Result{
		Task:       task,
		Confidence: "deterministic",
		Notes:      []string{"Prompt wrapped as an ordinary Axis task."},
	}, nil
}
