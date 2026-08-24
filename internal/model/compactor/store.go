package compactor

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/axis-cli/axis/internal/storeutil"
	"github.com/axis-cli/axis/internal/types"
)

// Store manages offloaded tool results on disk.
// Layout:
//
//	{dataDir}/refs/{timestamp}-{hash}.md   — full tool result text
//	{dataDir}/offload.jsonl                — append-only index of OffloadEntry
type Store struct {
	dataDir string
}

// NewStore creates a Store at the given directory, creating subdirs as needed.
func NewStore(dataDir string) (*Store, error) {
	refsDir := filepath.Join(dataDir, "refs")
	if err := os.MkdirAll(refsDir, 0755); err != nil {
		return nil, fmt.Errorf("compactor: init refs dir: %w", err)
	}
	return &Store{dataDir: dataDir}, nil
}

// Offload writes the full tool result to refs/ and appends an entry to offload.jsonl.
// Returns the OffloadEntry written.
func (s *Store) Offload(toolCallID, toolName, content string, summary string, score int) (*OffloadEntry, error) {
	if toolCallID == "" {
		return nil, fmt.Errorf("compactor: empty tool_call_id")
	}

	now := time.Now().UTC()
	filename := refFilename(toolCallID, now)
	refPath := filepath.Join("refs", filename)

	// Write full content to refs/
	fullPath := filepath.Join(s.dataDir, refPath)
	if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
		return nil, fmt.Errorf("compactor: write ref: %w", err)
	}

	entry := &OffloadEntry{
		Version:    1,
		Timestamp:  now,
		ToolCallID: toolCallID,
		ToolName:   toolName,
		Summary:    summary,
		ResultRef:  refPath,
		Score:      score,
		TokensOrig: estimateTokens(content),
	}

	// Append to offload.jsonl
	if err := s.appendEntry(entry); err != nil {
		return nil, err
	}

	return entry, nil
}

// ReadRef reads the full content of an offloaded tool result.
func (s *Store) ReadRef(refPath string) (string, error) {
	data, err := os.ReadFile(filepath.Join(s.dataDir, refPath))
	if err != nil {
		return "", fmt.Errorf("compactor: read ref: %w", err)
	}
	return string(data), nil
}

func (s *Store) appendEntry(entry *OffloadEntry) error {
	path := filepath.Join(s.dataDir, "offload.jsonl")
	return storeutil.AppendJSONL(path, entry, storeutil.DefaultJSONLOptions())
}

// estimateTokens returns a rough token count using the unified types estimator.
func estimateTokens(s string) int {
	return types.EstimateTokens(s)
}
