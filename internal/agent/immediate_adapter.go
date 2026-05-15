package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/axis-cli/axis/internal/memory/immediate"
	"github.com/axis-cli/axis/internal/memory/working"
)

// ImmediateMemoryAdapter provides situational context from ImmediateMemory.
type ImmediateMemoryAdapter struct {
	builder *immediate.ContextBuilder
	wm      *working.Engine
	budget  int
}

func NewImmediateMemoryAdapter(projectRoot string, wm *working.Engine, budget int) *ImmediateMemoryAdapter {
	if budget <= 0 {
		budget = 4000
	}
	return &ImmediateMemoryAdapter{
		builder: immediate.NewContextBuilder(projectRoot),
		wm:      wm,
		budget:  budget,
	}
}

// BuildSituationalContext returns a compact summary highlighting changed files.
func (a *ImmediateMemoryAdapter) BuildSituationalContext(ctx context.Context, taskID, intent string) string {
	if a.wm == nil {
		return ""
	}
	immCtx, err := a.builder.BuildFromWorkingSet(
		taskID, intent, nil, a.wm, immediate.NewTokenBudget(a.budget),
	)
	if err != nil || immCtx == nil {
		return ""
	}
	if immCtx.WorkingSet == nil || len(immCtx.WorkingSet.Bundles) == 0 {
		return ""
	}

	var changed []string
	for _, b := range immCtx.WorkingSet.Bundles {
		if b.FileChanged {
			changed = append(changed, b.Source)
		}
	}
	if len(changed) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("[CHANGED FILES since last observation]\n")
	for _, f := range changed {
		sb.WriteString(fmt.Sprintf("- %s\n", f))
	}
	return sb.String()
}
