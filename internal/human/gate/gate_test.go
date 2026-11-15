package gate

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/comm"
	"github.com/axis-cli/axis/internal/types"
)

type recordingEmitter struct{ events []string }

func (r *recordingEmitter) emit(eventType, taskID, message string) {
	r.events = append(r.events, eventType+"|"+taskID+"|"+message)
}

func newTestGate(t *testing.T, cfg Config) (*Gate, *comm.Inbox, *recordingEmitter) {
	t.Helper()
	inbox := comm.NewInbox(t.TempDir())
	em := &recordingEmitter{}
	g := New(inbox, cfg, em.emit)
	t.Cleanup(g.Close)
	return g, inbox, em
}

func decisionTask(id string) *types.AgentTask {
	return &types.AgentTask{TaskID: id}
}

// T3/T4 acceptance: request → human responds → parked caller resumes approved.
func TestGate_RequestRespondRoundTrip(t *testing.T) {
	cfg := DefaultConfig()
	cfg.NudgeInterval = time.Hour
	g, inbox, _ := newTestGate(t, cfg)

	resCh := make(chan error, 1)
	go func() {
		out, err := g.Request(decisionTask("task-1"), &types.HumanCallRequest{
			Kind:    types.HumanKindDecision,
			Prompt:  "A or B?",
			Options: []string{"A", "B"},
		})
		if err != nil {
			resCh <- err
			return
		}
		if !out.Approved || out.Choice != "B" {
			t.Errorf("unexpected outcome: %+v", out)
		}
		resCh <- nil
	}()

	deadline := time.Now().Add(5 * time.Second)
	var callID string
	for time.Now().Before(deadline) {
		recs, _, _ := g.List()
		if len(recs) == 1 && recs[0].Pending() {
			callID = recs[0].Message.ID
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	if callID == "" {
		t.Fatal("request never appeared in inbox")
	}

	if _, err := inbox.Get(ActorID, callID); err != nil {
		t.Fatal(err)
	}
	if _, err := g.RespondCall(callID, "B", ""); err != nil {
		t.Fatal(err)
	}
	if err := <-resCh; err != nil {
		t.Fatalf("parked caller: %v", err)
	}

	rec, _ := inbox.Get(ActorID, callID)
	if rec.Status != comm.InboxStatusCompleted || rec.ResolvedAt == nil {
		t.Fatalf("record not completed: %+v", rec)
	}
}

// T6 acceptance: deadline reached → degrade fail-closed, never approval.
func TestGate_DegradeOnDeadline(t *testing.T) {
	cfg := Config{WaitTimeout: 30 * time.Millisecond, NudgeInterval: 500 * time.Hour, MaxNudges: 99, MaxWaiting: 2}
	g, inbox, em := newTestGate(t, cfg)

	out, err := g.Request(decisionTask("task-2"), &types.HumanCallRequest{
		Kind:    types.HumanKindDecision,
		Options: []string{"safe", "risky"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Degraded {
		t.Fatalf("expected degraded outcome, got %+v", out)
	}
	if out.Approved {
		t.Fatal("degrade must never auto-approve")
	}
	if out.Output["conservative_default"] != "safe" {
		t.Fatalf("conservative default not recorded: %+v", out.Output)
	}
	recs, _, _ := inbox.List(ActorID)
	if recs[0].Status != comm.InboxStatusDegraded {
		t.Fatalf("record status = %s", recs[0].Status)
	}
	found := false
	for _, e := range em.events {
		if strings.HasPrefix(e, EventRequestDegraded+"|task-2") {
			found = true
		}
	}
	if !found {
		t.Fatalf("degraded event missing: %v", em.events)
	}
}

// T6 acceptance AC3: nudge allowance exhausted → degrade even before deadline.
func TestGate_MaxNudgesEarlyDegrade(t *testing.T) {
	cfg := Config{WaitTimeout: time.Hour, NudgeInterval: 10 * time.Millisecond, MaxNudges: 3, MaxWaiting: 2}
	g, inbox, _ := newTestGate(t, cfg)

	start := time.Now()
	out, err := g.Request(decisionTask("task-3"), &types.HumanCallRequest{Kind: types.HumanKindInfo})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Degraded {
		t.Fatalf("expected early degrade, got %+v", out)
	}
	if elapsed := time.Since(start); elapsed > 20*time.Second {
		t.Fatalf("early degrade took too long: %v", elapsed)
	}
	recs, _, _ := inbox.List(ActorID)
	if recs[0].Nudges != 3 {
		t.Fatalf("nudges = %d, want 3", recs[0].Nudges)
	}
}

// FR3: pending requests above MaxWaiting are rejected to protect the pool.
func TestGate_TooManyWaiting(t *testing.T) {
	cfg := Config{WaitTimeout: time.Hour, NudgeInterval: time.Hour, MaxNudges: 99, MaxWaiting: 1}
	g, _, _ := newTestGate(t, cfg)

	errCh := make(chan error, 1)
	go func() {
		_, err := g.Request(decisionTask("task-a"), &types.HumanCallRequest{Kind: types.HumanKindInfo})
		errCh <- err
	}()
	time.Sleep(10 * time.Millisecond)

	_, err := g.Request(decisionTask("task-b"), &types.HumanCallRequest{Kind: types.HumanKindInfo})
	if !errors.Is(err, ErrTooManyWaiting) {
		t.Fatalf("expected ErrTooManyWaiting, got %v", err)
	}
	g.Close()
	if err := <-errCh; err != nil && !errors.Is(err, ErrGateClosed) {
		t.Fatalf("first request: %v", err)
	}
}

// Shutdown drains parked waiters with ErrGateClosed instead of hanging.
func TestGate_CloseWakesParkedCaller(t *testing.T) {
	cfg := Config{WaitTimeout: time.Hour, NudgeInterval: time.Hour, MaxNudges: 99, MaxWaiting: 2}
	g, _, _ := newTestGate(t, cfg)

	errCh := make(chan error, 1)
	go func() {
		_, err := g.Request(decisionTask("task-x"), &types.HumanCallRequest{Kind: types.HumanKindInfo})
		errCh <- err
	}()
	time.Sleep(10 * time.Millisecond)
	g.Close()

	select {
	case err := <-errCh:
		if !errors.Is(err, ErrGateClosed) {
			t.Fatalf("expected ErrGateClosed, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("close did not wake parked caller")
	}
}

// Startup scan: only expired pending requests degrade; resolved ones stay.
func TestGate_DegradeExpired(t *testing.T) {
	cfg := Config{WaitTimeout: 20 * time.Millisecond, NudgeInterval: time.Hour, MaxNudges: 99, MaxWaiting: 4}
	g, inbox, _ := newTestGate(t, cfg)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = g.Request(decisionTask("task-e"), &types.HumanCallRequest{Kind: types.HumanKindInfo})
	}()
	time.Sleep(60 * time.Millisecond) // let it degrade naturally
	<-done

	// Seed a fresh expired-but-still-pending record directly.
	expired := &types.HumanCallRequest{
		CallID: "call-expired", TaskID: "task-f", Kind: types.HumanKindInfo,
		Deadline: time.Now().Add(-time.Minute),
	}
	_ = inbox.Deliver(ActorID, &comm.InboxRecord{
		Message: comm.Message{ID: expired.CallID, From: "axis-runtime", To: ActorID,
			Type: comm.MsgQuery, Payload: map[string]any{"request": expired}, Timestamp: time.Now().UTC()},
		Status: comm.InboxStatusPending,
	})

	n, err := g.DegradeExpired()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("degraded %d, want 1", n)
	}
	rec, _ := inbox.Get(ActorID, "call-expired")
	if rec.Status != comm.InboxStatusDegraded {
		t.Fatalf("expired record untouched: %+v", rec)
	}
}

// Resolve wins over concurrent degrade exactly once.
func TestGate_ResolveWinsOverDegrade(t *testing.T) {
	cfg := Config{WaitTimeout: 25 * time.Millisecond, NudgeInterval: 500 * time.Hour, MaxNudges: 99, MaxWaiting: 2}
	g, inbox, _ := newTestGate(t, cfg)

	outCh := make(chan *Outcome, 1)
	go func() {
		out, err := g.Request(decisionTask("task-r"), &types.HumanCallRequest{Kind: types.HumanKindInfo})
		if err != nil {
			t.Error(err)
			outCh <- nil
			return
		}
		outCh <- out
	}()

	time.Sleep(5 * time.Millisecond)
	recs, _, _ := g.List()
	if len(recs) == 0 {
		t.Fatal("no record")
	}
	callID := recs[0].Message.ID
	_, respondErr := g.RespondCall(callID, "", "human answered")

	out := <-outCh
	if respondErr == nil {
		if out == nil || !out.Approved {
			t.Fatalf("respond should win: out=%+v err=%v", out, respondErr)
		}
	} else if out == nil || !out.Degraded {
		t.Fatalf("if respond lost CAS it must be idempotent: out=%+v err=%v", out, respondErr)
	}
	rec, _ := inbox.Get(ActorID, callID)
	if rec.Status == comm.InboxStatusPending {
		t.Fatal("record stuck pending after race")
	}
}

// T5: sensitive keys never reach the persisted payload.
func TestGate_ScrubbingBeforePersist(t *testing.T) {
	cfg := Config{WaitTimeout: 40 * time.Millisecond, NudgeInterval: time.Hour, MaxNudges: 99, MaxWaiting: 2}
	g, _, _ := newTestGate(t, cfg)

	req := &types.HumanCallRequest{
		Kind:     types.HumanKindInfo,
		Context:  map[string]any{"path": "x.go", "api_key": "sk-leak", "Authorization": "Bearer x"},
		Metadata: map[string]string{"env": "dev", "secret_token": "leak"},
	}
	errCh := make(chan struct{})
	go func() { _, _ = g.Request(decisionTask("task-s"), req); close(errCh) }()
	time.Sleep(80 * time.Millisecond)
	<-errCh

	recs, _, err := g.List()
	if err != nil || len(recs) == 0 {
		t.Fatalf("no records: %v", err)
	}
	blob, mErr := json.Marshal(recs[0])
	if mErr != nil {
		t.Fatal(mErr)
	}
	s := string(blob)
	for _, leak := range []string{"sk-leak", "Bearer x", "\"leak\"", "api_key", "secret_token"} {
		if strings.Contains(s, leak) {
			t.Fatalf("sensitive key/value leaked into persisted record: %s", s)
		}
	}
}
