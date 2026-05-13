package main

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/memory/longterm"
)

// seedFailedTaskInLongterm writes a task.failed event so the immunity
// CLI has something to promote. Returns the task ID.
func seedFailedTaskInLongterm(t *testing.T, taskID, errorClass string) {
	t.Helper()
	dir := filepath.Join(".axis", "memory", "longterm")
	es, err := longterm.Open(dir)
	if err != nil {
		t.Fatalf("longterm.Open: %v", err)
	}
	defer es.Close()
	if err := es.Append(context.Background(), longterm.EventRecord{
		EventType: longterm.EventTaskFailed,
		EntityID:  taskID,
		Timestamp: time.Now().UTC(),
		Payload: map[string]any{
			"intent_kind":         "build.binary",
			"contract_tool_allow": []any{"go", "git"},
			"error_class":         errorClass,
		},
	}); err != nil {
		t.Fatalf("seed Append: %v", err)
	}
}

func TestImmunity_PromoteAndList(t *testing.T) {
	defer withTempCwd(t)()
	seedFailedTaskInLongterm(t, "task-cli-1", "failure.provider.timeout")

	out, err := execCLI(t, "memory", "immunity", "promote", "task-cli-1",
		"--cause", "504 timeout",
		"--by", "user:test",
	)
	if err != nil {
		t.Fatalf("promote: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Promoted task-cli-1 to immunity record imm-") {
		t.Errorf("promote output unexpected:\n%s", out)
	}
	if !strings.Contains(out, "failure.provider.timeout") {
		t.Errorf("promote should show derived class:\n%s", out)
	}

	out, err = execCLI(t, "memory", "immunity", "list")
	if err != nil {
		t.Fatalf("list: %v\n%s", err, out)
	}
	if !strings.Contains(out, "task-cli-1") && !strings.Contains(out, "imm-") {
		t.Errorf("list missing the new record:\n%s", out)
	}
	if !strings.Contains(out, "failure.provider.timeout") {
		t.Errorf("list missing class column:\n%s", out)
	}
}

func TestImmunity_PromoteRejectsMissingTask(t *testing.T) {
	defer withTempCwd(t)()
	out, err := execCLI(t, "memory", "immunity", "promote", "task-nope",
		"--cause", "x",
		"--by", "u",
	)
	if err == nil {
		t.Fatalf("expected error for missing task, got success\n%s", out)
	}
	if !strings.Contains(err.Error(), "not in terminal state") {
		t.Errorf("error message unexpected: %v", err)
	}
}

func TestImmunity_PromoteRequiresFlags(t *testing.T) {
	defer withTempCwd(t)()
	// Missing --cause and --by.
	_, err := execCLI(t, "memory", "immunity", "promote", "task-x")
	if err == nil {
		t.Errorf("expected error when required flags missing")
	}
}

func TestImmunity_ShowAndForget(t *testing.T) {
	defer withTempCwd(t)()
	seedFailedTaskInLongterm(t, "task-cli-2", "failure.tool.permission_denied")

	out, err := execCLI(t, "memory", "immunity", "promote", "task-cli-2",
		"--cause", "tool denied", "--by", "u", "--json",
	)
	if err != nil {
		t.Fatalf("promote: %v\n%s", err, out)
	}
	var rec struct {
		ImmunityID string `json:"immunity_id"`
	}
	if err := json.Unmarshal([]byte(out), &rec); err != nil {
		t.Fatalf("promote --json parse: %v\n%s", err, out)
	}
	if !strings.HasPrefix(rec.ImmunityID, "imm-") {
		t.Errorf("immunity_id format: %q", rec.ImmunityID)
	}

	out, err = execCLI(t, "memory", "immunity", "show", rec.ImmunityID)
	if err != nil {
		t.Fatalf("show: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Source task:") || !strings.Contains(out, "task-cli-2") {
		t.Errorf("show output incomplete:\n%s", out)
	}

	out, err = execCLI(t, "memory", "immunity", "forget", rec.ImmunityID,
		"--reason", "fixed in v2",
		"--by", "agent:janitor",
	)
	if err != nil {
		t.Fatalf("forget: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Forgot immunity record") {
		t.Errorf("forget output unexpected:\n%s", out)
	}

	// List default excludes deprecated.
	out, err = execCLI(t, "memory", "immunity", "list")
	if err != nil {
		t.Fatalf("list: %v\n%s", err, out)
	}
	if strings.Contains(out, rec.ImmunityID) {
		t.Errorf("default list should hide deprecated; got:\n%s", out)
	}

	// List --deprecated includes it.
	out, err = execCLI(t, "memory", "immunity", "list", "--deprecated")
	if err != nil {
		t.Fatalf("list --deprecated: %v\n%s", err, out)
	}
	if !strings.Contains(out, rec.ImmunityID) {
		t.Errorf("list --deprecated should include record; got:\n%s", out)
	}
}

func TestImmunity_ListEmpty(t *testing.T) {
	defer withTempCwd(t)()
	out, err := execCLI(t, "memory", "immunity", "list")
	if err != nil {
		t.Fatalf("list: %v\n%s", err, out)
	}
	if !strings.Contains(out, "No immunity records.") {
		t.Errorf("expected empty-state message:\n%s", out)
	}
}

func TestImmunity_ListJSONStable(t *testing.T) {
	defer withTempCwd(t)()
	out, err := execCLI(t, "memory", "immunity", "list", "--json")
	if err != nil {
		t.Fatalf("list --json: %v\n%s", err, out)
	}
	var parsed struct {
		Count   int `json:"count"`
		Records any `json:"records"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("list --json parse: %v\n%s", err, out)
	}
	if parsed.Count != 0 {
		t.Errorf("expected count=0, got %d", parsed.Count)
	}
}
