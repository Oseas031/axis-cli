package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/comm"
	humangate "github.com/axis-cli/axis/internal/human/gate"
)

func seedInboxRequest(t *testing.T, root, callID string, deadline time.Time) {
	t.Helper()
	inbox := comm.NewInbox(filepath.Join(root, ".axis", "comm"))
	req := map[string]any{
		"call_id": callID, "task_id": "task-" + callID, "kind": "decision",
		"prompt": "A or B?", "options": []string{"A", "B"}, "deadline": deadline,
	}
	err := inbox.Deliver(humangate.ActorID, &comm.InboxRecord{
		Message: comm.Message{
			ID: callID, From: "axis-runtime", To: humangate.ActorID,
			Type: comm.MsgQuery, Payload: map[string]any{"request": req},
			Timestamp: time.Now().UTC(),
		},
		Status: comm.InboxStatusPending,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func execInboxCmd(t *testing.T, args ...string) string {
	t.Helper()
	cmd := NewRootCommand(defaultApp)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("inbox %v failed: %v", args, err)
	}
	return out.String()
}

func TestInbox_ListEmptyOffline(t *testing.T) {
	root := t.TempDir()
	defaultApp.root = root
	defer func() { defaultApp.root = "" }()

	out := execInboxCmd(t, "inbox", "list")
	if !strings.Contains(out, "0 pending") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestInbox_ListShowsPendingWithHint(t *testing.T) {
	root := t.TempDir()
	defaultApp.root = root
	defer func() { defaultApp.root = "" }()
	seedInboxRequest(t, root, "call-1", time.Now().Add(time.Hour))

	out := execInboxCmd(t, "inbox", "list")
	if !strings.Contains(out, "1 pending") || !strings.Contains(out, "call-1") || !strings.Contains(out, "decision") {
		t.Fatalf("unexpected output: %q", out)
	}
	if !strings.Contains(out, "Next:") {
		t.Fatal("output contract requires next-step hint")
	}
}

func TestInbox_ListJSONShape(t *testing.T) {
	root := t.TempDir()
	defaultApp.root = root
	defer func() { defaultApp.root = "" }()
	seedInboxRequest(t, root, "call-j", time.Now().Add(time.Hour))

	out := execInboxCmd(t, "inbox", "list", "--json")
	var payload struct {
		Records []struct {
			Message struct {
				ID string `json:"id"`
			} `json:"message"`
			Status string `json:"status"`
		} `json:"records"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("invalid json: %v (%s)", err, out)
	}
	if len(payload.Records) != 1 || payload.Records[0].Message.ID != "call-j" || payload.Records[0].Status != "pending" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestInbox_RespondSinglePendingOffline(t *testing.T) {
	root := t.TempDir()
	defaultApp.root = root
	defer func() { defaultApp.root = "" }()
	seedInboxRequest(t, root, "call-solo", time.Now().Add(time.Hour))

	out := execInboxCmd(t, "inbox", "respond", "--choice", "B")
	if !strings.Contains(out, "Responded to call-solo") {
		t.Fatalf("unexpected output: %q", out)
	}

	inbox := comm.NewInbox(filepath.Join(root, ".axis", "comm"))
	rec, err := inbox.Get(humangate.ActorID, "call-solo")
	if err != nil {
		t.Fatal(err)
	}
	if rec.Status != comm.InboxStatusCompleted || rec.Output["choice"] != "B" {
		t.Fatalf("record not resolved correctly: %+v", rec)
	}

	// Double respond must fail (not pending).
	cmd := NewRootCommand(defaultApp)
	cmd.SetArgs([]string{"inbox", "respond", "--choice", "A"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("second respond must fail")
	}
}

func TestInbox_RespondAmbiguousRequiresCallID(t *testing.T) {
	root := t.TempDir()
	defaultApp.root = root
	defer func() { defaultApp.root = "" }()
	seedInboxRequest(t, root, "call-a", time.Now().Add(time.Hour))
	seedInboxRequest(t, root, "call-b", time.Now().Add(time.Hour))

	cmd := NewRootCommand(defaultApp)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"inbox", "respond", "--choice", "A"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("ambiguous respond must fail")
	} else if !strings.Contains(err.Error(), "--call-id") {
		t.Fatalf("error must guide to --call-id: %v", err)
	}
}

func TestInbox_ShowRendersContext(t *testing.T) {
	root := t.TempDir()
	defaultApp.root = root
	defer func() { defaultApp.root = "" }()
	seedInboxRequest(t, root, "call-show", time.Now().Add(time.Hour))

	out := execInboxCmd(t, "inbox", "show", "call-show")
	for _, want := range []string{"call-show", "pending", "decision", "A or B?", "[A B]"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}
