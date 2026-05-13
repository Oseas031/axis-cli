package immunity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/memory/longterm"
)

// fakeEventStore is an in-memory EventStore for fast deterministic tests.
type fakeEventStore struct {
	events []longterm.EventRecord
	failOn string // event type to fail on Append (for failure-injection)
}

func (f *fakeEventStore) Append(_ context.Context, e longterm.EventRecord) error {
	if f.failOn != "" && e.EventType == f.failOn {
		return errors.New("injected append failure")
	}
	f.events = append(f.events, e)
	return nil
}

func (f *fakeEventStore) QueryEvents(_ context.Context, filter longterm.EventFilter) ([]longterm.EventRecord, error) {
	var out []longterm.EventRecord
	for _, e := range f.events {
		if filter.EntityID != "" && e.EntityID != filter.EntityID {
			continue
		}
		if len(filter.EventTypes) > 0 {
			match := false
			for _, t := range filter.EventTypes {
				if e.EventType == t {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		out = append(out, e)
	}
	return out, nil
}

func newTestStore(t *testing.T, fake *fakeEventStore) *Store {
	t.Helper()
	s := NewStore(fake)
	ts := time.Date(2026, 5, 12, 4, 0, 0, 0, time.UTC)
	s.now = func() time.Time {
		ts = ts.Add(time.Millisecond) // monotonic for unique IDs
		return ts
	}
	return s
}

func seedFailedTask(f *fakeEventStore, taskID string) {
	f.events = append(f.events, longterm.EventRecord{
		EventType: longterm.EventTaskCreated,
		EntityID:  taskID,
		Timestamp: time.Date(2026, 5, 12, 3, 0, 0, 0, time.UTC),
		Payload: map[string]any{
			"intent_kind":         "build.binary",
			"contract_tool_allow": []any{"go", "git"},
			"intent_args":         map[string]any{"target": "axis"},
		},
	})
	f.events = append(f.events, longterm.EventRecord{
		EventType: longterm.EventTaskFailed,
		EntityID:  taskID,
		Timestamp: time.Date(2026, 5, 12, 3, 5, 0, 0, time.UTC),
		Payload: map[string]any{
			"intent_kind":         "build.binary",
			"contract_tool_allow": []any{"go", "git"},
			"intent_args":         map[string]any{"target": "axis"},
			"error_class":         "failure.provider.timeout",
		},
	})
}

func seedCompletedTask(f *fakeEventStore, taskID string) {
	f.events = append(f.events, longterm.EventRecord{
		EventType: longterm.EventTaskCompleted,
		EntityID:  taskID,
		Timestamp: time.Now().UTC(),
	})
}

func TestPromote_HappyPath(t *testing.T) {
	f := &fakeEventStore{}
	seedFailedTask(f, "task-1")
	s := newTestStore(t, f)

	rec, err := s.Promote(context.Background(), PromoteInput{
		SourceTaskID: "task-1",
		Cause:        "504 timeout on every retry",
		PromotedBy:   "user:alex",
	})
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if rec.ImmunityID == "" || len(rec.SignatureHash) != 32 {
		t.Errorf("malformed record: %+v", rec)
	}
	if rec.FailureClass != "failure.provider.timeout" {
		t.Errorf("expected class auto-derived from payload, got %q", rec.FailureClass)
	}

	// One promoted event should now exist.
	count := 0
	for _, e := range f.events {
		if e.EventType == longterm.EventImmunityPromoted {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 promoted event, got %d", count)
	}
}

func TestPromote_RejectsNonTerminalTask(t *testing.T) {
	f := &fakeEventStore{}
	s := newTestStore(t, f)

	_, err := s.Promote(context.Background(), PromoteInput{
		SourceTaskID: "task-missing",
		Cause:        "x",
		PromotedBy:   "u",
	})
	if !errors.Is(err, ErrTaskNotTerminal) {
		t.Errorf("expected ErrTaskNotTerminal, got %v", err)
	}
}

func TestPromote_RejectsSuccessfulTask(t *testing.T) {
	f := &fakeEventStore{}
	seedCompletedTask(f, "task-ok")
	s := newTestStore(t, f)

	_, err := s.Promote(context.Background(), PromoteInput{
		SourceTaskID: "task-ok",
		Cause:        "x",
		PromotedBy:   "u",
	})
	if !errors.Is(err, ErrTaskNotFailed) {
		t.Errorf("expected ErrTaskNotFailed, got %v", err)
	}
}

func TestPromote_RejectsEmptyCause(t *testing.T) {
	f := &fakeEventStore{}
	seedFailedTask(f, "task-1")
	s := newTestStore(t, f)

	_, err := s.Promote(context.Background(), PromoteInput{
		SourceTaskID: "task-1",
		Cause:        "   ",
		PromotedBy:   "u",
	})
	if !errors.Is(err, ErrCauseRequired) {
		t.Errorf("expected ErrCauseRequired, got %v", err)
	}
}

func TestPromote_RejectsUnknownClassFromInput(t *testing.T) {
	f := &fakeEventStore{}
	seedFailedTask(f, "task-1")
	s := newTestStore(t, f)

	_, err := s.Promote(context.Background(), PromoteInput{
		SourceTaskID: "task-1",
		Cause:        "x",
		PromotedBy:   "u",
		FailureClass: "failure.bogus.unknown",
	})
	if !errors.Is(err, ErrUnknownFailureClass) {
		t.Errorf("expected ErrUnknownFailureClass, got %v", err)
	}
}

func TestPromote_CallerClassOverridesPayload(t *testing.T) {
	f := &fakeEventStore{}
	seedFailedTask(f, "task-1") // payload says failure.provider.timeout
	s := newTestStore(t, f)

	rec, err := s.Promote(context.Background(), PromoteInput{
		SourceTaskID: "task-1",
		Cause:        "actually a contract issue",
		PromotedBy:   "u",
		FailureClass: "failure.contract.unsatisfied",
	})
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if rec.FailureClass != "failure.contract.unsatisfied" {
		t.Errorf("caller class should win, got %q", rec.FailureClass)
	}
}

func TestPromote_AppendFailurePropagates(t *testing.T) {
	f := &fakeEventStore{failOn: longterm.EventImmunityPromoted}
	seedFailedTask(f, "task-1")
	s := newTestStore(t, f)

	_, err := s.Promote(context.Background(), PromoteInput{
		SourceTaskID: "task-1",
		Cause:        "x",
		PromotedBy:   "u",
	})
	if err == nil {
		t.Errorf("expected append failure to propagate")
	}
	// No event recorded.
	for _, e := range f.events {
		if e.EventType == longterm.EventImmunityPromoted {
			t.Errorf("partial state: promoted event recorded despite append failure")
		}
	}
}

func TestShow_ReturnsPromotedRecord(t *testing.T) {
	f := &fakeEventStore{}
	seedFailedTask(f, "task-1")
	s := newTestStore(t, f)

	rec, _ := s.Promote(context.Background(), PromoteInput{
		SourceTaskID: "task-1", Cause: "x", PromotedBy: "u",
	})

	got, err := s.Show(context.Background(), rec.ImmunityID)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if got.ImmunityID != rec.ImmunityID || got.Cause != "x" {
		t.Errorf("Show returned wrong record: %+v", got)
	}
	if got.Deprecated {
		t.Errorf("freshly promoted record should not be deprecated")
	}
}

func TestShow_UnknownIDReturnsNotFound(t *testing.T) {
	f := &fakeEventStore{}
	s := newTestStore(t, f)

	_, err := s.Show(context.Background(), "imm-nonexistent")
	if !errors.Is(err, ErrImmunityNotFound) {
		t.Errorf("expected ErrImmunityNotFound, got %v", err)
	}
}

func TestForget_MarksDeprecated(t *testing.T) {
	f := &fakeEventStore{}
	seedFailedTask(f, "task-1")
	s := newTestStore(t, f)

	rec, _ := s.Promote(context.Background(), PromoteInput{
		SourceTaskID: "task-1", Cause: "x", PromotedBy: "u",
	})

	if err := s.Forget(context.Background(), rec.ImmunityID, "fixed by upgrade", "agent:auto"); err != nil {
		t.Fatalf("Forget: %v", err)
	}

	got, err := s.Show(context.Background(), rec.ImmunityID)
	if err != nil {
		t.Fatalf("Show after Forget: %v", err)
	}
	if !got.Deprecated {
		t.Errorf("expected Deprecated=true after Forget")
	}
	if got.DeprecateReason != "fixed by upgrade" {
		t.Errorf("DeprecateReason mismatch: %q", got.DeprecateReason)
	}
	if got.DeprecatedAt == nil {
		t.Errorf("DeprecatedAt should be set")
	}

	// Verify original promoted event untouched.
	for _, e := range f.events {
		if e.EventType == longterm.EventImmunityPromoted {
			if e.Payload["cause"] != "x" {
				t.Errorf("promoted event mutated after Forget")
			}
		}
	}
}

func TestForget_UnknownIDReturnsNotFound(t *testing.T) {
	f := &fakeEventStore{}
	s := newTestStore(t, f)

	err := s.Forget(context.Background(), "imm-nope", "x", "u")
	if !errors.Is(err, ErrImmunityNotFound) {
		t.Errorf("expected ErrImmunityNotFound, got %v", err)
	}
}

func TestList_FilterByClassAndDeprecated(t *testing.T) {
	f := &fakeEventStore{}
	seedFailedTask(f, "task-1")
	seedFailedTask(f, "task-2")
	s := newTestStore(t, f)

	r1, _ := s.Promote(context.Background(), PromoteInput{
		SourceTaskID: "task-1", Cause: "a", PromotedBy: "u",
		FailureClass: "failure.provider.timeout",
	})
	r2, _ := s.Promote(context.Background(), PromoteInput{
		SourceTaskID: "task-2", Cause: "b", PromotedBy: "u",
		FailureClass: "failure.tool.permission_denied",
	})
	_ = s.Forget(context.Background(), r2.ImmunityID, "fixed", "u")

	// Default: deprecated excluded.
	got, err := s.List(context.Background(), ListFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 || got[0].ImmunityID != r1.ImmunityID {
		t.Errorf("expected only non-deprecated record, got %d records", len(got))
	}

	// With IncludeDeprecated: both.
	got, _ = s.List(context.Background(), ListFilter{IncludeDeprecated: true})
	if len(got) != 2 {
		t.Errorf("expected 2 records with IncludeDeprecated, got %d", len(got))
	}

	// Filter by class.
	got, _ = s.List(context.Background(), ListFilter{
		Class:             "failure.tool.permission_denied",
		IncludeDeprecated: true,
	})
	if len(got) != 1 || got[0].ImmunityID != r2.ImmunityID {
		t.Errorf("class filter returned wrong records: %+v", got)
	}
}

func TestList_LimitRespected(t *testing.T) {
	f := &fakeEventStore{}
	for i := 0; i < 5; i++ {
		seedFailedTask(f, fmtTaskID(i))
	}
	s := newTestStore(t, f)
	for i := 0; i < 5; i++ {
		_, _ = s.Promote(context.Background(), PromoteInput{
			SourceTaskID: fmtTaskID(i), Cause: "x", PromotedBy: "u",
		})
	}

	got, _ := s.List(context.Background(), ListFilter{Limit: 2})
	if len(got) != 2 {
		t.Errorf("Limit=2 should yield 2, got %d", len(got))
	}
}

func TestPromote_AgainstRealLongtermStore(t *testing.T) {
	// Integration smoke: ensure Promote also works against the real
	// longterm.FileStore (not just the fake). Catches API drift.
	dir := t.TempDir()
	es, err := longterm.Open(dir)
	if err != nil {
		t.Fatalf("longterm.Open: %v", err)
	}
	defer es.Close()

	ctx := context.Background()
	_ = es.Append(ctx, longterm.EventRecord{
		EventType: longterm.EventTaskFailed,
		EntityID:  "task-real",
		Timestamp: time.Now().UTC(),
		Payload: map[string]any{
			"intent_kind": "scan.repo",
			"error_class": "failure.runtime.panic",
		},
	})

	s := NewStore(es)
	rec, err := s.Promote(ctx, PromoteInput{
		SourceTaskID: "task-real",
		Cause:        "panic in scanner",
		PromotedBy:   "agent:scanner",
	})
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if rec.FailureClass != "failure.runtime.panic" {
		t.Errorf("class derivation from real store wrong: %q", rec.FailureClass)
	}
}

func fmtTaskID(i int) string {
	return "task-" + string(rune('a'+i))
}
