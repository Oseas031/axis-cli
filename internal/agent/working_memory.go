package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/axis-cli/axis/internal/memory/working"
)

// WorkingMemoryRecaller queries WorkingMemory BM25 for relevant context.
type WorkingMemoryRecaller struct {
	engine *working.Engine
	limit  int
}

func NewWorkingMemoryRecaller(engine *working.Engine, limit int) *WorkingMemoryRecaller {
	if limit <= 0 {
		limit = 5
	}
	return &WorkingMemoryRecaller{engine: engine, limit: limit}
}

// Recall returns a formatted context block from BM25 search.
func (r *WorkingMemoryRecaller) Recall(ctx context.Context, query string) string {
	if query == "" || r.engine == nil {
		return ""
	}
	hits, err := r.engine.Recall(ctx, query, r.limit)
	if err != nil || len(hits) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("[WORKING MEMORY]\n")
	for _, h := range hits {
		sb.WriteString(fmt.Sprintf("- [%s] %s: %s (score=%.2f)\n", h.Type, h.Source, h.Summary, h.Relevance))
	}
	return sb.String()
}
