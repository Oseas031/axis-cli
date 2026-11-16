package control

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

func TestWorkbenchExt_SkillsAndMemory(t *testing.T) {
	dir := t.TempDir()
	srv := NewServer(stubRuntime{}, RuntimeRecord{Protocol: "http", Address: "http://127.0.0.1:4321", ProjectRoot: dir})
	srv.SetUIEnabled(true)
	h := srv.Handler()

	// N7: Save Skill
	skBody, _ := json.Marshal(SaveSkillRequest{
		Name:        "Check-Repo",
		Description: "Verify repository status",
		Prompt:      "run git status",
		Harness:     "axis-native",
		Output:      "clean",
	})
	skReq := httptest.NewRequest(http.MethodPost, "/v1/skills", bytes.NewReader(skBody))
	skRec := httptest.NewRecorder()
	h.ServeHTTP(skRec, skReq)
	if skRec.Code != http.StatusCreated {
		t.Fatalf("POST /v1/skills status = %d, body = %s", skRec.Code, skRec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".axis", "skills", "check-repo", "SKILL.md")); err != nil {
		t.Fatalf("expected SKILL.md on disk: %v", err)
	}

	skListReq := httptest.NewRequest(http.MethodGet, "/v1/skills", nil)
	skListRec := httptest.NewRecorder()
	h.ServeHTTP(skListRec, skListReq)
	if skListRec.Code != http.StatusOK || !strings.Contains(skListRec.Body.String(), "Verify repository status") {
		t.Fatalf("unexpected GET /v1/skills response: %d %s", skListRec.Code, skListRec.Body.String())
	}

	// N9: Memory CRUD
	memBody, _ := json.Marshal(MemoryNote{Title: "Preference", Content: "Use Go stdlib only", Source: "ui"})
	memReq := httptest.NewRequest(http.MethodPost, "/v1/memory", bytes.NewReader(memBody))
	memRec := httptest.NewRecorder()
	h.ServeHTTP(memRec, memReq)
	if memRec.Code != http.StatusCreated {
		t.Fatalf("POST /v1/memory status = %d", memRec.Code)
	}
	var createdNote MemoryNote
	if err := json.Unmarshal(memRec.Body.Bytes(), &createdNote); err != nil {
		t.Fatalf("decode note: %v", err)
	}

	memListReq := httptest.NewRequest(http.MethodGet, "/v1/memory", nil)
	memListRec := httptest.NewRecorder()
	h.ServeHTTP(memListRec, memListReq)
	if !strings.Contains(memListRec.Body.String(), "Use Go stdlib only") {
		t.Fatalf("expected memory note in list: %s", memListRec.Body.String())
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/v1/memory?id="+createdNote.ID, nil)
	delRec := httptest.NewRecorder()
	h.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusOK {
		t.Fatalf("DELETE /v1/memory status = %d", delRec.Code)
	}
}

func TestWorkbenchExt_RoutinesAndMultiHarnessPipeline(t *testing.T) {
	dir := t.TempDir()
	evLog := NewTaskEventLog(dir)
	srv := NewServerWithEventLog(stubRuntime{}, RuntimeRecord{
		Protocol:    "http",
		Address:     "http://127.0.0.1:4321",
		ProjectRoot: dir,
	}, evLog)
	srv.SetUIEnabled(true)
	h := srv.Handler()

	t.Setenv("GO_WANT_HARNESS_HELPER", "1")
	regBody, _ := json.Marshal(HarnessSpec{
		ID:      "pipe-cli",
		Name:    "Pipe CLI",
		Kind:    "cli",
		Command: os.Args[0],
		Args:    []string{"-test.run=^TestHelperCLIHarness$", "--", "--prompt", "{prompt}"},
		EnvKeys: []string{"GO_WANT_HARNESS_HELPER"},
	})
	regReq := httptest.NewRequest(http.MethodPost, "/v1/harnesses", bytes.NewReader(regBody))
	regRec := httptest.NewRecorder()
	h.ServeHTTP(regRec, regReq)
	if regRec.Code != http.StatusCreated {
		t.Fatalf("register pipe-cli status = %d", regRec.Code)
	}

	// N9: Routine create + manual trigger + toggle
	rtnBody, _ := json.Marshal(Routine{
		ID:       "r-check",
		Name:     "Check",
		Harness:  "pipe-cli",
		Prompt:   "routine-ping",
		Interval: "10m",
		Enabled:  true,
	})
	rtnReq := httptest.NewRequest(http.MethodPost, "/v1/routines", bytes.NewReader(rtnBody))
	rtnRec := httptest.NewRecorder()
	h.ServeHTTP(rtnRec, rtnReq)
	if rtnRec.Code != http.StatusCreated {
		t.Fatalf("POST /v1/routines status = %d, body = %s", rtnRec.Code, rtnRec.Body.String())
	}

	runReq := httptest.NewRequest(http.MethodPost, "/v1/routines/r-check/run", nil)
	runRec := httptest.NewRecorder()
	h.ServeHTTP(runRec, runReq)
	if runRec.Code != http.StatusAccepted {
		t.Fatalf("POST /v1/routines/r-check/run status = %d", runRec.Code)
	}

	togReq := httptest.NewRequest(http.MethodPost, "/v1/routines/r-check/toggle", nil)
	togRec := httptest.NewRecorder()
	h.ServeHTTP(togRec, togReq)
	if togRec.Code != http.StatusOK || !strings.Contains(togRec.Body.String(), `"enabled":false`) {
		t.Fatalf("unexpected toggle response: %d %s", togRec.Code, togRec.Body.String())
	}

	// N10: 2-step pipeline chaining {{prev_output}}
	pipeBody, _ := json.Marshal(map[string]any{
		"id":   "p-test-1",
		"name": "Two Step",
		"steps": []PipelineStep{
			{Harness: "pipe-cli", Prompt: "step1-seed"},
			{Harness: "pipe-cli", Prompt: "step2-got={{prev_output}}"},
		},
	})
	pipeReq := httptest.NewRequest(http.MethodPost, "/v1/pipelines", bytes.NewReader(pipeBody))
	pipeRec := httptest.NewRecorder()
	h.ServeHTTP(pipeRec, pipeReq)
	if pipeRec.Code != http.StatusAccepted {
		t.Fatalf("POST /v1/pipelines status = %d, body = %s", pipeRec.Code, pipeRec.Body.String())
	}

	deadline := time.Now().Add(5 * time.Second)
	var finalRun PipelineRun
	for time.Now().Before(deadline) {
		listReq := httptest.NewRequest(http.MethodGet, "/v1/pipelines", nil)
		listRec := httptest.NewRecorder()
		h.ServeHTTP(listRec, listReq)
		var payload struct {
			Pipelines []PipelineRun `json:"pipelines"`
		}
		if json.Unmarshal(listRec.Body.Bytes(), &payload) == nil && len(payload.Pipelines) > 0 {
			finalRun = payload.Pipelines[0]
			if finalRun.Status == types.TaskStatusCompleted || finalRun.Status == types.TaskStatusFailed {
				break
			}
		}
		time.Sleep(30 * time.Millisecond)
	}
	if finalRun.Status != types.TaskStatusCompleted || len(finalRun.StepTaskIDs) != 2 {
		t.Fatalf("unexpected pipeline result: %+v", finalRun)
	}

	// Verify step 2 received step 1's output
	stReq := httptest.NewRequest(http.MethodGet, "/v1/tasks/"+finalRun.StepTaskIDs[1]+"/status", nil)
	stRec := httptest.NewRecorder()
	h.ServeHTTP(stRec, stReq)
	var stResp StatusResponse
	if err := json.Unmarshal(stRec.Body.Bytes(), &stResp); err != nil {
		t.Fatalf("decode step2 status: %v", err)
	}
	step2Text, _ := stResp.Output["text"].(string)
	if !strings.Contains(step2Text, "echo:step1-seed") {
		t.Fatalf("expected step2 output to carry step1 output, got %q", step2Text)
	}
}
