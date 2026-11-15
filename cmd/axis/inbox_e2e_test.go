package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/comm"
	"github.com/axis-cli/axis/internal/control"
	humangate "github.com/axis-cli/axis/internal/human/gate"
	"github.com/axis-cli/axis/internal/types"
)

// T3 acceptance (AC1): a human task parks as waiting_human, the inbox shows
// it, an HTTP respond resumes it to completed, and events are audited.
func TestHumanInboxEndToEnd(t *testing.T) {
	resetCLIState()
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- runLocalRuntime(ctx, root, io.Discard, 0)
	}()
	locator := control.NewRuntimeLocator(root)
	for i := 0; i < 50; i++ {
		if _, err := locator.Load(); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	client := control.NewClient(locator, control.LocalHTTPClient())

	task := &types.AgentTask{
		TaskID:     "human-e2e",
		ContractID: "default",
		Input:      map[string]any{"message": "hello", "prompt": "confirm something"},
		Metadata:   map[string]string{types.TaskMetadataKeyExecutor: types.ExecutorTypeHuman},
	}
	if _, err := client.SubmitTask(context.Background(), task); err != nil {
		t.Fatalf("submit human task: %v", err)
	}

	// Wait for waiting_human status (task admitted → running → parked).
	waitForStatus := func(want types.TaskStatus) types.TaskStatus {
		for i := 0; i < 100; i++ {
			st, err := client.Status(context.Background(), "human-e2e")
			if err == nil && st.Status == want {
				return st.Status
			}
			time.Sleep(20 * time.Millisecond)
		}
		st, _ := client.Status(context.Background(), "human-e2e")
		return st.Status
	}
	if got := waitForStatus(types.TaskStatusWaitingHuman); got != types.TaskStatusWaitingHuman {
		t.Fatalf("status = %s, want waiting_human", got)
	}

	// Inbox lists exactly one pending decision request.
	listResp, err := client.InboxList(context.Background())
	if err != nil {
		t.Fatalf("inbox list: %v", err)
	}
	if len(listResp.Records) != 1 || listResp.Records[0].Status != comm.InboxStatusPending {
		t.Fatalf("unexpected inbox: %+v", listResp.Records)
	}
	callID := listResp.Records[0].Message.ID

	// Respond over HTTP; the parked task must resume to completed.
	if _, err := client.InboxRespond(context.Background(), callID, "", "because reasons"); err != nil {
		t.Fatalf("respond: %v", err)
	}
	final := waitForStatus(types.TaskStatusCompleted)
	if final != types.TaskStatusCompleted {
		st, _ := client.Status(context.Background(), "human-e2e")
		t.Fatalf("status after respond = %s (%+v)", final, st)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("runtime returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("runtime did not stop")
	}

	data, err := os.ReadFile(control.NewTaskEventLog(root).Path())
	if err != nil {
		t.Fatal(err)
	}
	logText := string(data)
	for _, want := range []string{"\"event_type\":\"human.request.created\"", "\"event_type\":\"human.request.resolved\""} {
		if !strings.Contains(logText, want) {
			t.Fatalf("missing %s in event log", want)
		}
	}
}

// T3 acceptance (AC2): a pending request that expired while the runtime was
// down is degraded fail-closed by the startup scan.
func TestStartupScanDegradesExpiredRequests(t *testing.T) {
	root := t.TempDir()

	// Seed an expired pending request directly into the per-actor file.
	inbox := comm.NewInbox(filepath.Join(root, ".axis", "comm"))
	expiredReq := map[string]any{
		"call_id": "call-expired", "task_id": "task-gone", "kind": "info",
		"deadline": time.Now().Add(-time.Hour),
	}
	if err := inbox.Deliver(humangate.ActorID, &comm.InboxRecord{
		Message: comm.Message{
			ID: "call-expired", From: "axis-runtime", To: humangate.ActorID,
			Type: comm.MsgQuery, Payload: map[string]any{"request": expiredReq},
			Timestamp: time.Now().UTC(),
		},
		Status: comm.InboxStatusPending,
	}); err != nil {
		t.Fatal(err)
	}

	resetCLIState()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- runLocalRuntime(ctx, root, io.Discard, 0)
	}()
	locator := control.NewRuntimeLocator(root)
	for i := 0; i < 50; i++ {
		if _, err := locator.Load(); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	client := control.NewClient(locator, control.LocalHTTPClient())

	var degraded bool
	for i := 0; i < 50; i++ {
		resp, err := client.InboxList(context.Background())
		if err == nil && len(resp.Records) > 0 && resp.Records[0].Status == comm.InboxStatusDegraded {
			degraded = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !degraded {
		t.Fatal("startup scan did not degrade the expired request")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("runtime did not stop")
	}
}

// T2/D2 single-writer discipline: starting a second runtime on the same root
// fails while the first is healthy.
func TestRunLocalRuntimeRefusesSecondInstance(t *testing.T) {
	resetCLIState()
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- runLocalRuntime(ctx, root, io.Discard, 0)
	}()
	locator := control.NewRuntimeLocator(root)
	for i := 0; i < 50; i++ {
		if _, err := locator.Load(); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	err := runLocalRuntime(ctx, root, io.Discard, 0)
	if err == nil || !strings.Contains(err.Error(), "already running") {
		t.Fatalf("second runtime must be refused, got %v", err)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("first runtime did not stop")
	}
}

// compile-time guard: keep the http import used when handlers evolve.
var _ = http.StatusOK
var _ = json.Marshal
