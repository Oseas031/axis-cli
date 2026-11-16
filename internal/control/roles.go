package control

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/axis-cli/axis/internal/types"
)

// RoleSpec defines a specialized agent role with scoped tools, prompt, and isolation defaults.
type RoleSpec struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	DefaultHarness string   `json:"default_harness"`
	Isolate        bool     `json:"isolate"`
	AllowedTools   []string `json:"allowed_tools"`
	SystemPrompt   string   `json:"system_prompt"`
}

// HandoffRecord is a typed, structured contract passed between pipeline/DAG stages.
type HandoffRecord struct {
	FromRole         string   `json:"from_role"`
	ToRole           string   `json:"to_role,omitempty"`
	TaskID           string   `json:"task_id"`
	Summary          string   `json:"summary"`
	Evidence         string   `json:"evidence,omitempty"`
	WorktreeRef      string   `json:"worktree_ref,omitempty"`
	NextDependencies []string `json:"next_dependencies,omitempty"`
}

var builtinRoles = []RoleSpec{
	{
		ID:             "coordinator",
		Name:           "Coordinator",
		Description:    "Decomposes goals into DAG steps, assigns specialists, and reviews outcomes without mutating files.",
		DefaultHarness: defaultHarnessID,
		Isolate:        false,
		AllowedTools:   []string{"read_file", "list_dir", "grep_search"},
		SystemPrompt:   "[Role: Coordinator] Analyze the goal, decompose into verifiable steps, and specify acceptance criteria. Do not mutate repository files directly.",
	},
	{
		ID:             "implementer",
		Name:           "Implementer",
		Description:    "Writes and edits code inside an isolated git worktree and reports exact changed files.",
		DefaultHarness: defaultHarnessID,
		Isolate:        true,
		AllowedTools:   []string{"read_file", "write_file", "edit_file", "bash"},
		SystemPrompt:   "[Role: Implementer] Implement the requested changes cleanly in the isolated worktree. Summarize modified files and why.",
	},
	{
		ID:             "test-runner",
		Name:           "Test Runner",
		Description:    "Executes tests, linters, and build verification, producing deterministic pass/fail evidence.",
		DefaultHarness: defaultHarnessID,
		Isolate:        false,
		AllowedTools:   []string{"read_file", "bash"},
		SystemPrompt:   "[Role: Test Runner] Run the relevant tests and build checks, and report exact command outputs and pass/fail counts.",
	},
	{
		ID:             "reviewer",
		Name:           "Reviewer",
		Description:    "Performs read-only code and diff review against requirements and safety boundaries.",
		DefaultHarness: defaultHarnessID,
		Isolate:        false,
		AllowedTools:   []string{"read_file", "grep_search"},
		SystemPrompt:   "[Role: Reviewer] Review the diff and test evidence for correctness, regressions, and security boundaries. Output verdict: APPROVE or REQUEST_CHANGES.",
	},
}

func getRoleSpec(id string) (RoleSpec, bool) {
	id = strings.TrimSpace(strings.ToLower(id))
	for _, r := range builtinRoles {
		if r.ID == id {
			return r, true
		}
	}
	return RoleSpec{}, false
}

func (s *Server) handleRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", "")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"roles": builtinRoles})
}

func buildTypedHandoff(fromRole, toRole, taskID, outText, diffText string) (HandoffRecord, error) {
	summary := strings.TrimSpace(outText)
	if summary == "" && strings.TrimSpace(diffText) == "" {
		return HandoffRecord{}, fmt.Errorf("typed handoff rejected: step %s produced empty output and diff", taskID)
	}
	if len(summary) > 2000 {
		summary = summary[:2000] + "\n...(truncated)"
	}
	rec := HandoffRecord{
		FromRole:    fromRole,
		ToRole:      toRole,
		TaskID:      taskID,
		Summary:     summary,
		WorktreeRef: strings.TrimSpace(diffText),
	}
	if rec.FromRole == "" {
		rec.FromRole = "executor"
	}
	// If output contains a JSON handoff block, parse evidence/next_dependencies when available.
	if idx := strings.Index(outText, "{"); idx >= 0 {
		if end := strings.LastIndex(outText, "}"); end > idx {
			var parsed struct {
				Summary          string   `json:"summary"`
				Evidence         string   `json:"evidence"`
				NextDependencies []string `json:"next_dependencies"`
			}
			if json.Unmarshal([]byte(outText[idx:end+1]), &parsed) == nil {
				if parsed.Summary != "" {
					rec.Summary = parsed.Summary
				}
				if parsed.Evidence != "" {
					rec.Evidence = parsed.Evidence
				}
				if len(parsed.NextDependencies) > 0 {
					rec.NextDependencies = parsed.NextDependencies
				}
			}
		}
	}
	return rec, nil
}

// --- CostBudget pre-call hard gate ---

func parseTaskBudgetUSD(task *types.AgentTask) float64 {
	if task == nil {
		return 0
	}
	if task.CostBudget > 0 {
		return task.CostBudget
	}
	if task.Metadata != nil {
		if raw := strings.TrimSpace(task.Metadata["cost_budget_usd"]); raw != "" {
			if v, err := strconv.ParseFloat(raw, 64); err == nil && v > 0 {
				return v
			}
		}
	}
	return 0
}

func parseTaskSpentUSD(task *types.AgentTask) float64 {
	if task == nil || task.Metadata == nil {
		return 0
	}
	if raw := strings.TrimSpace(task.Metadata["spent_usd"]); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil && v >= 0 {
			return v
		}
	}
	return 0
}

func checkBudgetBeforeDispatch(task *types.AgentTask) error {
	budget := parseTaskBudgetUSD(task)
	if budget <= 0 {
		return nil
	}
	spent := parseTaskSpentUSD(task)
	if spent >= budget {
		return fmt.Errorf("budget_exceeded: spent $%.4f >= limit $%.4f", spent, budget)
	}
	return nil
}

func estimateCharCostUSD(prompt, output string) float64 {
	totalChars := len(prompt) + len(output)
	if totalChars == 0 {
		return 0.0001
	}
	// Conservative local estimate (~4 chars/token, ~$0.002 per 1k tokens)
	estTokens := float64(totalChars) / 4.0
	cost := (estTokens / 1000.0) * 0.002
	if cost < 0.0001 {
		cost = 0.0001
	}
	return cost
}
