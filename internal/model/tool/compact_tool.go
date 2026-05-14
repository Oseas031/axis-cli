package tool

import (
	"context"

	"github.com/axis-cli/axis/internal/types"
)

// CompactTool allows the Agent to signal that context compaction is desired.
type CompactTool struct{}

func NewCompactTool() *CompactTool { return &CompactTool{} }

func (t *CompactTool) Name() string { return "compact" }

func (t *CompactTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "compact",
		Description: "Request context compaction to free up conversation space. Use when the conversation is getting long.",
		Parameters:  []types.FieldDef{},
	}
}

func (t *CompactTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	return map[string]any{
		"status":  "compaction_requested",
		"message": "Context compaction will be applied on next turn.",
	}, nil
}
