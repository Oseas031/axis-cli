package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/axis-cli/axis/internal/model/providerconfig"
)

func TestConfigDump_JSONAndRedaction(t *testing.T) {
	root := t.TempDir()
	store := providerconfig.NewStore(root)
	secretKey := "sk-secret-do-not-leak-123456"
	if err := store.Save(&providerconfig.Config{
		ActiveProfile: "ds",
		Profiles: map[string]providerconfig.Profile{
			"ds": {
				Name:     "ds",
				Provider: "deepseek",
				Model:    "deepseek-chat",
				APIKey:   secretKey,
			},
		},
	}); err != nil {
		t.Fatalf("save provider config: %v", err)
	}

	app := &App{providerName: "mock", root: root}
	cmd := NewRootCommand(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"config", "dump", "--json", "--runtime-profile", "minimal"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config dump failed: %v\n%s", err, out.String())
	}

	raw := out.String()
	if strings.Contains(raw, secretKey) {
		t.Fatalf("config dump leaked API key in output: %s", raw)
	}

	var parsed configDump
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, raw)
	}
	if parsed.RuntimeProfile != "minimal" {
		t.Errorf("runtime_profile = %q, want minimal", parsed.RuntimeProfile)
	}
	if parsed.Provider != "deepseek" || parsed.ProfileName != "ds" {
		t.Errorf("unexpected provider/profile: %+v", parsed)
	}
	if !parsed.HasAPIKey {
		t.Error("expected has_api_key=true")
	}
	if parsed.RuntimeLive {
		t.Error("expected runtime_live=false when no runtime is running")
	}
}

func TestConfigDump_RejectsInvalidRuntimeProfile(t *testing.T) {
	app := &App{providerName: "mock", root: t.TempDir()}
	cmd := NewRootCommand(app)
	cmd.SetArgs([]string{"config", "dump", "--runtime-profile", "unknown"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for invalid --runtime-profile")
	}
}

func TestRuntimeProfile_MinimalInitializesOrchestrator(t *testing.T) {
	app := &App{
		providerName:   "mock",
		runtimeProfile: "minimal",
		root:           t.TempDir(),
	}
	app.initOrchestrator()
	if app.orch == nil {
		t.Fatal("expected minimal profile to initialize orchestrator")
	}
}

func TestAskCmd_JSONOutput(t *testing.T) {
	app := &App{providerName: "mock", root: t.TempDir()}
	cmd := NewRootCommand(app)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"ask", "check", "build", "--task-id", "ask-json-test", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("ask --json failed: %v\n%s", err, out.String())
	}

	raw := out.String()
	if strings.Contains(raw, "Not submitted.") {
		t.Fatalf("unexpected human prose in --json output: %s", raw)
	}

	var payload struct {
		Submitted bool `json:"submitted"`
		Task      struct {
			TaskID string `json:"task_id"`
		} `json:"task"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("invalid JSON from ask --json: %v\n%s", err, raw)
	}
	if payload.Submitted {
		t.Error("expected submitted=false for dry-run")
	}
	if payload.Task.TaskID != "ask-json-test" {
		t.Errorf("task_id = %q, want ask-json-test", payload.Task.TaskID)
	}
}
