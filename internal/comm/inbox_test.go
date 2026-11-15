package comm

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func inboxMsg(id, to string) Message {
	return Message{
		ID:        id,
		From:      "axis-runtime",
		To:        to,
		Type:      MsgQuery,
		Payload:   map[string]any{"call_id": id},
		Timestamp: time.Now().UTC(),
	}
}

func TestInbox_DeliverListIsolation(t *testing.T) {
	inbox := NewInbox(t.TempDir())
	if err := inbox.Deliver("human", &InboxRecord{Message: inboxMsg("m1", "human")}); err != nil {
		t.Fatal(err)
	}
	if err := inbox.Deliver("human", &InboxRecord{Message: inboxMsg("m2", "human"), Status: InboxStatusCompleted}); err != nil {
		t.Fatal(err)
	}
	if err := inbox.Deliver("agent-1", &InboxRecord{Message: inboxMsg("m3", "agent-1")}); err != nil {
		t.Fatal(err)
	}

	recs, corrupted, err := inbox.List("human")
	if err != nil || corrupted != 0 {
		t.Fatalf("list: %v corrupted=%d", err, corrupted)
	}
	if len(recs) != 2 || recs[0].Message.ID != "m1" || recs[1].Status != InboxStatusCompleted {
		t.Fatalf("unexpected records: %+v", recs)
	}

	// Per-actor isolation: agent file does not see human records.
	agentRecs, _, err := inbox.List("agent-1")
	if err != nil || len(agentRecs) != 1 || agentRecs[0].Message.ID != "m3" {
		t.Fatalf("isolation broken: %+v %v", agentRecs, err)
	}

	if _, err := os.Stat(filepath.Join(inbox.Dir(), "human.jsonl")); err != nil {
		t.Fatalf("per-actor file missing: %v", err)
	}
}

func TestInbox_MutatePendingCAS(t *testing.T) {
	inbox := NewInbox(t.TempDir())
	_ = inbox.Deliver("human", &InboxRecord{Message: inboxMsg("c1", "human")})

	err := inbox.MutatePending("human", "c1", func(r *InboxRecord) {
		r.Status = InboxStatusCompleted
		r.Output = map[string]any{"choice": "B"}
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := inbox.Get("human", "c1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != InboxStatusCompleted || got.ResolvedAt == nil {
		t.Fatalf("resolve not applied: %+v", got)
	}
	if got.Output["choice"] != "B" {
		t.Fatalf("output not persisted: %+v", got.Output)
	}

	// Loser of the race gets ErrNotPending (resolve wins over degrade).
	err = inbox.MutatePending("human", "c1", func(r *InboxRecord) { r.Status = InboxStatusDegraded })
	if !errors.Is(err, ErrNotPending) {
		t.Fatalf("expected ErrNotPending, got %v", err)
	}

	if err := inbox.MutatePending("human", "missing", func(r *InboxRecord) {}); !errors.Is(err, ErrCallNotFound) {
		t.Fatalf("expected ErrCallNotFound, got %v", err)
	}
}

func TestInbox_RejectsOversizedAndKeepsFileReadable(t *testing.T) {
	inbox := NewInbox(t.TempDir())
	inbox.MaxRecordBytes = 512
	if err := inbox.Deliver("human", &InboxRecord{Message: inboxMsg("ok", "human")}); err != nil {
		t.Fatal(err)
	}

	big := &InboxRecord{Message: inboxMsg("big", "human")}
	big.Message.Payload = map[string]any{"blob": strings.Repeat("x", 4096)}
	err := inbox.Deliver("human", big)
	if !errors.Is(err, ErrRecordTooLarge) {
		t.Fatalf("expected ErrRecordTooLarge, got %v", err)
	}

	recs, corrupted, err := inbox.List("human")
	if err != nil || corrupted != 0 || len(recs) != 1 {
		t.Fatalf("file poisoned after rejection: n=%d corrupted=%d err=%v", len(recs), corrupted, err)
	}
}

func TestInbox_SkipsCorruptLinesWithCount(t *testing.T) {
	dir := t.TempDir()
	inbox := NewInbox(dir)
	_ = inbox.Deliver("human", &InboxRecord{Message: inboxMsg("good", "human")})

	f, _ := os.OpenFile(filepath.Join(dir, "human.jsonl"), os.O_APPEND|os.O_WRONLY, 0o640)
	_, _ = f.WriteString("{\"message\":{\"id\":\"broken\"}\n") // truncated JSON
	_, _ = f.WriteString("\n")                                 // blank line
	_ = f.Close()
	_ = inbox.Deliver("human", &InboxRecord{Message: inboxMsg("good2", "human")})

	recs, corrupted, err := inbox.List("human")
	if err != nil {
		t.Fatal(err)
	}
	if corrupted != 1 {
		t.Fatalf("corrupted count = %d, want 1", corrupted)
	}
	if len(recs) != 2 || recs[0].Message.ID != "good" || recs[1].Message.ID != "good2" {
		t.Fatalf("valid records damaged by corrupt line: %+v", recs)
	}
}

func TestInbox_LegacyMigration(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, legacyMailboxName)
	content := "{\"id\":\"old-1\",\"from\":\"a\",\"to\":\"human\",\"type\":\"query\",\"payload\":{},\"timestamp\":\"2026-08-01T00:00:00Z\"}\n" +
		"{\"id\":\"old-2\",\"from\":\"b\",\"to\":\"agent-9\",\"type\":\"task\",\"payload\":{},\"timestamp\":\"2026-08-01T00:00:00Z\"}\n"
	if err := os.WriteFile(legacy, []byte(content), 0o640); err != nil {
		t.Fatal(err)
	}

	inbox := NewInbox(dir)
	humanRecs, _, err := inbox.List("human")
	if err != nil {
		t.Fatal(err)
	}
	if len(humanRecs) != 1 || humanRecs[0].Message.ID != "old-1" || humanRecs[0].Status != InboxStatusPending {
		t.Fatalf("legacy human message not migrated: %+v", humanRecs)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatalf("legacy file should be renamed, stat err=%v", err)
	}
	if _, err := os.Stat(legacy + legacyMigratedSuffix); err != nil {
		t.Fatalf("imported marker missing: %v", err)
	}

	// Idempotent: second operation must not duplicate.
	again, _, err := inbox.List("human")
	if err != nil || len(again) != 1 {
		t.Fatalf("migration re-ran: %+v %v", again, err)
	}
}

func TestInbox_ConcurrentDeliverAndMutate(t *testing.T) {
	inbox := NewInbox(t.TempDir())
	_ = inbox.Deliver("human", &InboxRecord{Message: inboxMsg("race", "human")})

	var wg sync.WaitGroup
	for k := 0; k < 8; k++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			_ = inbox.Deliver("human", &InboxRecord{Message: inboxMsg(string(rune('a'+k)), "human")})
			_ = inbox.MutatePending("human", "race", func(r *InboxRecord) { r.Status = InboxStatusCompleted })
		}(k)
	}
	wg.Wait()

	if _, _, err := inbox.List("human"); err != nil {
		t.Fatal(err)
	}
}

func TestValidateActorID(t *testing.T) {
	valid := []string{"human", "agent-1", "a.b_c"}
	for _, id := range valid {
		if err := ValidateActorID(id); err != nil {
			t.Fatalf("%q should be valid: %v", id, err)
		}
	}
	invalid := []string{"", "..", ".", "a/b", `a\b`, "a\\b", "sp ace", strings.Repeat("x", 65), "大"}
	for _, id := range invalid {
		if err := ValidateActorID(id); !errors.Is(err, ErrInvalidActorID) {
			t.Fatalf("%q should be rejected, got %v", id, err)
		}
	}
}
