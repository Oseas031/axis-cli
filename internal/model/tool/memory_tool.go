package tool

import (
	"context"
	"fmt"
	"time"

	"github.com/axis-cli/axis/internal/memory/horizon"
	"github.com/axis-cli/axis/internal/types"
)

// RecallMemoryTool allows the Agent to search long-horizon memory.
type RecallMemoryTool struct {
	store *horizon.Store
}

func NewRecallMemoryTool(store *horizon.Store) *RecallMemoryTool {
	return &RecallMemoryTool{store: store}
}

func (t *RecallMemoryTool) Name() string { return "recall_memory" }

func (t *RecallMemoryTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "recall_memory",
		Description: "Search long-horizon memory (patterns, principles, narrative) by keyword. Use only when facing unknown scenarios, high-risk decisions, or when current patterns seem insufficient.",
		Parameters: []types.FieldDef{
			{Name: "keyword", Type: types.FieldTypeString, Required: true, Description: "Search keyword"},
			{Name: "category", Type: types.FieldTypeString, Required: false, Description: "Filter by category: patterns, principles, or narrative. Empty searches all."},
		},
	}
}

func (t *RecallMemoryTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	keyword, _ := input["keyword"].(string)
	category, _ := input["category"].(string)

	entries, err := t.store.Recall(keyword, horizon.Category(category))
	if err != nil {
		return map[string]any{"error": fmt.Sprintf("recall failed: %v", err)}, nil
	}

	results := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		results = append(results, map[string]any{
			"id":       e.ID,
			"title":    e.Title,
			"category": string(e.Category),
			"tags":     e.Tags,
			"body":     e.Body,
		})
	}
	return map[string]any{"count": len(results), "entries": results}, nil
}

// StoreMemoryTool allows the Agent to write to long-horizon memory.
type StoreMemoryTool struct {
	store *horizon.Store
}

func NewStoreMemoryTool(store *horizon.Store) *StoreMemoryTool {
	return &StoreMemoryTool{store: store}
}

func (t *StoreMemoryTool) Name() string { return "store_memory" }

func (t *StoreMemoryTool) Schema() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        "store_memory",
		Description: "Store a new entry in long-horizon memory. Use after distilling a pattern, principle, or narrative from task execution.",
		Parameters: []types.FieldDef{
			{Name: "category", Type: types.FieldTypeString, Required: true, Description: "Category: patterns, principles, or narrative"},
			{Name: "title", Type: types.FieldTypeString, Required: true, Description: "Short title for the memory entry"},
			{Name: "body", Type: types.FieldTypeString, Required: true, Description: "Content of the memory entry (markdown)"},
			{Name: "tags", Type: types.FieldTypeString, Required: false, Description: "Comma-separated tags"},
		},
	}
}

func (t *StoreMemoryTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	category, _ := input["category"].(string)
	title, _ := input["title"].(string)
	body, _ := input["body"].(string)
	tagsStr, _ := input["tags"].(string)

	if category == "" || title == "" || body == "" {
		return map[string]any{"error": "category, title, and body are required"}, nil
	}

	cat := horizon.Category(category)
	if cat != horizon.CategoryPatterns && cat != horizon.CategoryPrinciples && cat != horizon.CategoryNarrative {
		return map[string]any{"error": "category must be one of: patterns, principles, narrative"}, nil
	}

	var tags []string
	if tagsStr != "" {
		tags = append(tags, splitTags(tagsStr)...)
	}

	id := fmt.Sprintf("%s-%d", sanitizeForID(title), time.Now().UnixMilli())
	entry := horizon.Entry{
		ID:        id,
		Category:  cat,
		Title:     title,
		Tags:      tags,
		CreatedAt: time.Now(),
		Body:      body,
	}

	if err := t.store.Store(entry); err != nil {
		return map[string]any{"error": fmt.Sprintf("store failed: %v", err)}, nil
	}
	return map[string]any{"status": "stored", "id": id, "category": category}, nil
}

func splitTags(s string) []string {
	var tags []string
	for _, t := range splitComma(s) {
		t = trimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

func splitComma(s string) []string {
	result := []string{}
	for _, part := range []byte(s) {
		if part == ',' {
			result = append(result, "")
		} else {
			if len(result) == 0 {
				result = append(result, "")
			}
			result[len(result)-1] += string(part)
		}
	}
	return result
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

func sanitizeForID(title string) string {
	r := []byte{}
	for _, c := range []byte(title) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			r = append(r, c)
		} else if c >= 'A' && c <= 'Z' {
			r = append(r, c+32)
		} else if c == ' ' || c == '_' {
			r = append(r, '-')
		}
	}
	if len(r) > 40 {
		r = r[:40]
	}
	return string(r)
}
