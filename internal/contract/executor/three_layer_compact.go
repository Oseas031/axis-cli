package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

// TranscriptStore persists full conversation transcripts before compaction.
type TranscriptStore struct {
	Dir string // Directory to store transcripts
}

// Save writes the full message history to disk before compaction.
func (s *TranscriptStore) Save(taskID string, history []types.ModelMessage) (string, error) {
	if s.Dir == "" {
		return "", nil
	}
	if err := os.MkdirAll(s.Dir, 0755); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("%s-%d.json", taskID, time.Now().UnixMilli())
	path := filepath.Join(s.Dir, filename)
	data, err := json.Marshal(history)
	if err != nil {
		return "", err
	}
	return path, os.WriteFile(path, data, 0644)
}

// ThreeLayerCompaction implements the three-layer context compaction model:
//   - Layer 1 (micro): ToolResultCompaction — runs every turn, replaces old tool results
//   - Layer 2 (auto): Saves transcript + SummarizationCompaction — threshold-triggered
//   - Layer 3 (manual): Same as Layer 2, triggered by Agent via compact tool
//
// This wraps CompactionPipeline with transcript persistence.
type ThreeLayerCompaction struct {
	// Micro layer: always runs (cheap, no LLM call)
	Micro *ToolResultCompaction
	// Auto layer: runs when tokens exceed threshold
	Auto *SummarizationCompaction
	// Transcript store for persistence before summarization
	Transcript *TranscriptStore
	// Budget is the token threshold that triggers auto compaction
	Budget int
	// TaskID for transcript naming
	TaskID string
}

// Compact runs the three-layer compaction pipeline.
func (c *ThreeLayerCompaction) Compact(ctx context.Context, history []types.ModelMessage) []types.ModelMessage {
	// Layer 1: micro_compact (always runs)
	if c.Micro != nil {
		history, _ = c.Micro.Compact(ctx, history, c.Budget)
	}

	// Layer 2: auto_compact (threshold-triggered)
	if EstimateTokens(history) > c.Budget && c.Auto != nil {
		// Save transcript before summarization
		if c.Transcript != nil {
			c.Transcript.Save(c.TaskID, history)
		}
		history, _ = c.Auto.Compact(ctx, history, c.Budget)
	}

	return history
}
