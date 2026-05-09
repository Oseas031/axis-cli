// Package agent provides self-context management for agent autonomy.
package agent

import (
	"sort"
	"strings"
)

// CompressionStrategy defines how to compress context.
type CompressionStrategy int

const (
	StrategyByPriority CompressionStrategy = iota
	StrategyByLineage
	StrategyByRecency
)

// ContextCompressor compresses SelfContext to fit within token limits.
type ContextCompressor struct {
	strategy CompressionStrategy
}

// NewContextCompressor creates a new ContextCompressor with the given strategy.
func NewContextCompressor(strategy CompressionStrategy) *ContextCompressor {
	return &ContextCompressor{
		strategy: strategy,
	}
}

// Compress reduces the SelfContext to fit within maxTokens.
// It uses priority-based truncation when token limit is exceeded.
func (cc *ContextCompressor) Compress(ctx *SelfContext, maxTokens int) *SelfContext {
	if ctx == nil {
		return nil
	}

	if maxTokens <= 0 {
		maxTokens = 1000 // default
	}

	// Estimate current tokens
	currentTokens := cc.estimateTokens(ctx)
	if currentTokens <= maxTokens {
		return ctx.Clone()
	}

	// Clone and compress based on strategy
	result := ctx.Clone()

	switch cc.strategy {
	case StrategyByPriority:
		result = cc.compressByPriority(result, maxTokens)
	case StrategyByLineage:
		result = cc.compressByLineage(result, maxTokens)
	case StrategyByRecency:
		result = cc.compressByRecency(result, maxTokens)
	default:
		result = cc.compressByPriority(result, maxTokens)
	}

	return result
}

// estimateTokens estimates the token count of a SelfContext.
// This is a rough estimation based on structure, not actual tokenization.
func (cc *ContextCompressor) estimateTokens(ctx *SelfContext) int {
	if ctx == nil {
		return 0
	}

	tokens := 50 // Base overhead for structure

	// TaskID
	tokens += len(ctx.TaskID) / 4

	// TaskLineage
	tokens += len(ctx.TaskLineage) * 10

	// CodeSnapshot
	if ctx.CodeSnapshot != nil {
		tokens += 30 // structure overhead
		tokens += len(ctx.CodeSnapshot.ModifiedFiles) * 15
		tokens += len(ctx.CodeSnapshot.SpecVersion) / 4
		tokens += ctx.CodeSnapshot.TaskCount * 2
		tokens += ctx.CodeSnapshot.ToolCount * 2
	}

	// DocSnapshot
	if ctx.DocSnapshot != nil {
		tokens += 30
		tokens += len(ctx.DocSnapshot.SpecFiles) * 10
		tokens += len(ctx.DocSnapshot.DocFiles) * 10
		tokens += len(ctx.DocSnapshot.Reports) * 10
	}

	// StateSnapshot
	if ctx.StateSnapshot != nil {
		tokens += 20
		tokens += ctx.StateSnapshot.RunningTasks * 2
		tokens += ctx.StateSnapshot.PendingTasks * 2
		tokens += ctx.StateSnapshot.CompletedTasks * 2
		tokens += ctx.StateSnapshot.FailedTasks * 2
	}

	// AutonomyLevel and CompetenceScore
	tokens += 10

	return tokens
}

// compressByPriority compresses context by keeping highest priority items.
func (cc *ContextCompressor) compressByPriority(ctx *SelfContext, maxTokens int) *SelfContext {
	// Priority order:
	// 1. StateSnapshot (most relevant for execution)
	// 2. TaskID and TaskLineage
	// 3. CodeSnapshot
	// 4. DocSnapshot

	result := &SelfContext{
		TaskID:          ctx.TaskID,
		TaskLineage:     ctx.TaskLineage,
		AutonomyLevel:   ctx.AutonomyLevel,
		CompetenceScore: ctx.CompetenceScore,
	}

	// Always keep state snapshot
	if ctx.StateSnapshot != nil {
		result.StateSnapshot = ctx.StateSnapshot
	}

	// Keep code snapshot but truncate files if needed
	if ctx.CodeSnapshot != nil {
		codeTokens := 30 + len(ctx.CodeSnapshot.ModifiedFiles)*15 + ctx.CodeSnapshot.TaskCount*2
		remaining := maxTokens - cc.estimateTokens(result) - 50

		result.CodeSnapshot = &CodeSnapshot{
			SpecVersion: ctx.CodeSnapshot.SpecVersion,
			TaskCount:   ctx.CodeSnapshot.TaskCount,
			ToolCount:   ctx.CodeSnapshot.ToolCount,
		}

		if codeTokens <= remaining {
			result.CodeSnapshot.ModifiedFiles = ctx.CodeSnapshot.ModifiedFiles
		} else {
			// Keep only most important files (shortest paths first)
			sort.Slice(ctx.CodeSnapshot.ModifiedFiles, func(i, j int) bool {
				return len(ctx.CodeSnapshot.ModifiedFiles[i]) < len(ctx.CodeSnapshot.ModifiedFiles[j])
			})
			maxFiles := remaining / 15
			if maxFiles > 0 && maxFiles < len(ctx.CodeSnapshot.ModifiedFiles) {
				result.CodeSnapshot.ModifiedFiles = ctx.CodeSnapshot.ModifiedFiles[:maxFiles]
			}
		}
	}

	// Truncate doc snapshot if needed
	if ctx.DocSnapshot != nil {
		remaining := maxTokens - cc.estimateTokens(result)
		docTokens := 30 + len(ctx.DocSnapshot.SpecFiles)*10 + len(ctx.DocSnapshot.DocFiles)*10 + len(ctx.DocSnapshot.Reports)*10

		result.DocSnapshot = &DocSnapshot{}

		if docTokens <= remaining {
			result.DocSnapshot = ctx.DocSnapshot
		} else {
			// Keep spec files first, then docs, then reports
			result.DocSnapshot.SpecFiles = ctx.DocSnapshot.SpecFiles
			remaining -= len(ctx.DocSnapshot.SpecFiles) * 10

			if remaining > 0 {
				for _, f := range ctx.DocSnapshot.DocFiles {
					if remaining < 10 {
						break
					}
					result.DocSnapshot.DocFiles = append(result.DocSnapshot.DocFiles, f)
					remaining -= 10
				}
			}

			if remaining > 0 {
				for _, r := range ctx.DocSnapshot.Reports {
					if remaining < 10 {
						break
					}
					result.DocSnapshot.Reports = append(result.DocSnapshot.Reports, r)
					remaining -= 10
				}
			}
		}
	}

	return result
}

// compressByLineage compresses context by keeping most recent lineage entries.
func (cc *ContextCompressor) compressByLineage(ctx *SelfContext, maxTokens int) *SelfContext {
	result := ctx.Clone()

	// Truncate lineage to keep only most recent entries
	lineageTokens := len(ctx.TaskLineage) * 10
	remaining := maxTokens - cc.estimateTokens(ctx) + lineageTokens

	if remaining < 0 && len(ctx.TaskLineage) > 0 {
		keepCount := remaining / 10
		if keepCount < 1 {
			keepCount = 1
		}
		// Keep most recent (last in list)
		result.TaskLineage = ctx.TaskLineage[len(ctx.TaskLineage)-keepCount:]
	}

	return result
}

// compressByRecency compresses context by prioritizing recent items.
func (cc *ContextCompressor) compressByRecency(ctx *SelfContext, maxTokens int) *SelfContext {
	result := &SelfContext{
		TaskID:          ctx.TaskID,
		AutonomyLevel:   ctx.AutonomyLevel,
		CompetenceScore: ctx.CompetenceScore,
	}

	// Keep last 3 lineage entries (most recent)
	if len(ctx.TaskLineage) > 0 {
		n := len(ctx.TaskLineage)
		if n > 3 {
			n = 3
		}
		result.TaskLineage = ctx.TaskLineage[len(ctx.TaskLineage)-n:]
	}

	// Keep state snapshot
	result.StateSnapshot = ctx.StateSnapshot

	// Keep code snapshot with truncated file list
	if ctx.CodeSnapshot != nil {
		result.CodeSnapshot = &CodeSnapshot{
			SpecVersion: ctx.CodeSnapshot.SpecVersion,
			TaskCount:   ctx.CodeSnapshot.TaskCount,
			ToolCount:   ctx.CodeSnapshot.ToolCount,
		}
		// Keep only first 10 files
		if len(ctx.CodeSnapshot.ModifiedFiles) > 10 {
			result.CodeSnapshot.ModifiedFiles = ctx.CodeSnapshot.ModifiedFiles[:10]
		} else {
			result.CodeSnapshot.ModifiedFiles = ctx.CodeSnapshot.ModifiedFiles
		}
	}

	// Keep doc snapshot with truncated lists
	if ctx.DocSnapshot != nil {
		result.DocSnapshot = &DocSnapshot{}
		// Keep first 5 items from each list
		if len(ctx.DocSnapshot.SpecFiles) > 5 {
			result.DocSnapshot.SpecFiles = ctx.DocSnapshot.SpecFiles[:5]
		} else {
			result.DocSnapshot.SpecFiles = ctx.DocSnapshot.SpecFiles
		}
		if len(ctx.DocSnapshot.DocFiles) > 5 {
			result.DocSnapshot.DocFiles = ctx.DocSnapshot.DocFiles[:5]
		} else {
			result.DocSnapshot.DocFiles = ctx.DocSnapshot.DocFiles
		}
		if len(ctx.DocSnapshot.Reports) > 5 {
			result.DocSnapshot.Reports = ctx.DocSnapshot.Reports[:5]
		} else {
			result.DocSnapshot.Reports = ctx.DocSnapshot.Reports
		}
	}

	return result
}

// CompressFiles compresses a list of file paths to fit within maxTokens.
func CompressFiles(files []string, maxTokens int) []string {
	if len(files) == 0 || maxTokens <= 0 {
		return files
	}

	// Sort by path length (shorter = higher priority)
	sorted := make([]string, len(files))
	copy(sorted, files)
	sort.Slice(sorted, func(i, j int) bool {
		return len(sorted[i]) < len(sorted[j])
	})

	result := make([]string, 0)
	tokens := 0
	for _, f := range sorted {
		fileTokens := len(strings.Split(f, "/")) * 2
		if tokens+fileTokens > maxTokens {
			break
		}
		result = append(result, f)
		tokens += fileTokens
	}

	return result
}
