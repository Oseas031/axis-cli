// Package evolution provides data models and storage for the Staged Evolution Protocol.
package evolution

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EvolutionRoot is the default project-local directory for evolution runs.
const EvolutionRoot = ".axis/evolution"

// Store manages project-local evolution run records.
type Store struct {
	root string
}

// NewStore creates a new evolution store at the given root path.
// If root is empty, it defaults to EvolutionRoot in the current working directory.
func NewStore(root string) (*Store, error) {
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}
		root = filepath.Join(wd, EvolutionRoot)
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, fmt.Errorf("create evolution root: %w", err)
	}
	return &Store{root: root}, nil
}

// RunDir returns the directory path for a given run ID.
func (s *Store) RunDir(runID string) string {
	return filepath.Join(s.root, runID)
}

// CreateRun initializes a new run directory with an intent and run record.
// It returns an error if the run already exists.
func (s *Store) CreateRun(intent EvolutionIntent, run EvolutionRun) error {
	runDir := s.RunDir(run.RunID)
	if _, err := os.Stat(runDir); err == nil {
		return fmt.Errorf("run %s already exists", run.RunID)
	}
	if err := os.MkdirAll(runDir, 0755); err != nil {
		return fmt.Errorf("create run directory: %w", err)
	}

	if err := s.writeJSON(filepath.Join(runDir, "intent.json"), intent); err != nil {
		return fmt.Errorf("write intent: %w", err)
	}
	if err := s.writeJSON(filepath.Join(runDir, "run.json"), run); err != nil {
		return fmt.Errorf("write run: %w", err)
	}
	return nil
}

// ReadRun reads the run record for the given run ID.
func (s *Store) ReadRun(runID string) (*EvolutionRun, error) {
	path := filepath.Join(s.RunDir(runID), "run.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("run %s not found", runID)
		}
		return nil, fmt.Errorf("read run: %w", err)
	}
	var run EvolutionRun
	if err := json.Unmarshal(data, &run); err != nil {
		return nil, fmt.Errorf("parse run: %w", err)
	}
	return &run, nil
}

// ReadIntent reads the intent record for the given run ID.
func (s *Store) ReadIntent(runID string) (*EvolutionIntent, error) {
	path := filepath.Join(s.RunDir(runID), "intent.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("intent for run %s not found", runID)
		}
		return nil, fmt.Errorf("read intent: %w", err)
	}
	var intent EvolutionIntent
	if err := json.Unmarshal(data, &intent); err != nil {
		return nil, fmt.Errorf("parse intent: %w", err)
	}
	return &intent, nil
}

// ListRuns returns all run IDs stored in the evolution root.
func (s *Store) ListRuns() ([]string, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		return nil, fmt.Errorf("list runs: %w", err)
	}
	var runs []string
	for _, entry := range entries {
		if entry.IsDir() {
			runs = append(runs, entry.Name())
		}
	}
	return runs, nil
}

// AppendDecision appends a decision record to the run directory as an append-only JSONL file.
func (s *Store) AppendDecision(runID string, decision EvolutionDecision) error {
	runDir := s.RunDir(runID)
	if _, err := os.Stat(runDir); os.IsNotExist(err) {
		return fmt.Errorf("run %s not found", runID)
	}
	path := filepath.Join(runDir, "decisions.jsonl")
	data, err := json.Marshal(decision)
	if err != nil {
		return fmt.Errorf("marshal decision: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open decisions ledger: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write decision: %w", err)
	}
	if _, err := f.WriteString("\n"); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}
	return nil
}

// ReadDecision reads the latest decision record for a run.
func (s *Store) ReadDecision(runID string) (*EvolutionDecision, error) {
	path := filepath.Join(s.RunDir(runID), "decisions.jsonl")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("decision for run %s not found", runID)
		}
		return nil, fmt.Errorf("read decision: %w", err)
	}
	defer f.Close()

	var lastLine string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lastLine = scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan decisions: %w", err)
	}
	if lastLine == "" {
		return nil, fmt.Errorf("decision for run %s not found", runID)
	}
	var decision EvolutionDecision
	if err := json.Unmarshal([]byte(lastLine), &decision); err != nil {
		return nil, fmt.Errorf("parse decision: %w", err)
	}
	return &decision, nil
}

// writeJSON atomically writes a JSON file by writing to a temp file and renaming.
func (s *Store) writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}

// GenerateRunID creates a unique run ID based on timestamp.
func GenerateRunID() string {
	return fmt.Sprintf("run-%d", time.Now().UnixNano())
}
