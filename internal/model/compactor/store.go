package compactor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
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
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("compactor: open jsonl: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("compactor: marshal entry: %w", err)
	}
	data = append(data, '\n')
	_, err = f.Write(data)
	return err
}

// estimateTokens provides a rough token count: CJK chars / 1.5 + other / 4.
func estimateTokens(s string) int {
	cjk := 0
	other := 0
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF || r >= 0x3000 && r <= 0x303F {
			cjk++
		} else {
			other++
		}
	}
	return int(float64(cjk)/1.5) + other/4
}
