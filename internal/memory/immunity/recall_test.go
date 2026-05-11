package immunity

import (
	"context"
	"testing"

	"github.com/axis-cli/axis/internal/memory/longterm"
)

// seedThreePromotions returns a Store + Recaller with three promoted
// records covering different intents and tool sets.
func seedThreePromotions(t *testing.T) (*Store, *Recaller, []ImmunityRecord) {
	t.Helper()
	f := &fakeEventStore{}
	ctx := context.Background()

	// task-1: intent=build.binary, tools=[go, git], class=provider.timeout
	f.events = append(f.events, eventFailed("task-1", "build.binary", []any{"go", "git"}, "failure.provider.timeout"))
	// task-2: intent=build.binary, tools=[go, git, make], class=provider.timeout
	f.events = append(f.events, eventFailed("task-2", "build.binary", []any{"go", "git", "make"}, "failure.provider.timeout"))
	// task-3: intent=scan.repo, tools=[grep], class=tool.permission_denied
	f.events = append(f.events, eventFailed("task-3", "scan.repo", []any{"grep"}, "failure.tool.permission_denied"))

	s := newTestStore(t, f)
	recs := make([]ImmunityRecord, 0, 3)
	for _, taskID := range []string{"task-1", "task-2", "task-3"} {
		rec, err := s.Promote(ctx, PromoteInput{
			SourceTaskID: taskID, Cause: "x", PromotedBy: "u",
		})
		if err != nil {
			t.Fatalf("seed Promote %s: %v", taskID, err)
		}
		recs = append(recs, rec)
	}
	return s, NewRecaller(s), recs
}

func eventFailed(taskID, intent string, tools []any, class string) longterm.EventRecord {
	return longterm.EventRecord{
		EventType: longterm.EventTaskFailed,
		EntityID:  taskID,
		Payload: map[string]any{
			"intent_kind":         intent,
			"contract_tool_allow": tools,
			"error_class":         class,
		},
	}
}

func TestRecall_ExactSignatureMatch(t *testing.T) {
	_, r, recs := seedThreePromotions(t)

	// task-1's signature should match itself exactly.
	got, err := r.Recall(context.Background(), recs[0].Signature, 0)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 exact match, got %d", len(got))
	}
	if got[0].ImmunityID != recs[0].ImmunityID {
		t.Errorf("matched wrong record: got %s want %s", got[0].ImmunityID, recs[0].ImmunityID)
	}
}

func TestRecall_NoMatchReturnsEmpty(t *testing.T) {
	_, r, _ := seedThreePromotions(t)

	// Build a signature unrelated to anything seeded.
	mystery := BuildSignature("totally.unrelated", nil, nil, "failure.runtime.unknown")
	got, _ := r.Recall(context.Background(), mystery, 0)
	if len(got) != 0 {
		t.Errorf("expected zero matches, got %d", len(got))
	}
}

func TestRecall_LimitRespected(t *testing.T) {
	s, r, _ := seedThreePromotions(t)
	ctx := context.Background()

	// Promote two more failures with the SAME signature as task-1.
	f := s.events.(*fakeEventStore)
	f.events = append(f.events, eventFailed("task-1-dup1", "build.binary", []any{"go", "git"}, "failure.provider.timeout"))
	f.events = append(f.events, eventFailed("task-1-dup2", "build.binary", []any{"go", "git"}, "failure.provider.timeout"))
	_, _ = s.Promote(ctx, PromoteInput{SourceTaskID: "task-1-dup1", Cause: "x", PromotedBy: "u"})
	_, _ = s.Promote(ctx, PromoteInput{SourceTaskID: "task-1-dup2", Cause: "x", PromotedBy: "u"})

	// Three records share task-1's signature; limit=2 should cap.
	target := BuildSignature("build.binary",
		map[string]string{},
		[]string{"go", "git"},
		"failure.provider.timeout",
	)
	got, _ := r.Recall(context.Background(), target, 2)
	if len(got) != 2 {
		t.Errorf("limit=2 should yield 2, got %d", len(got))
	}
}

func TestRecallSimilar_IntentOnly(t *testing.T) {
	_, r, _ := seedThreePromotions(t)
	// task-1 and task-2 share intent build.binary.
	got, err := r.RecallSimilar(context.Background(), PartialSignature{IntentKind: "build.binary"}, 0)
	if err != nil {
		t.Fatalf("RecallSimilar: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 matches for intent=build.binary, got %d", len(got))
	}
}

func TestRecallSimilar_ToolSupersetMatch(t *testing.T) {
	_, r, _ := seedThreePromotions(t)
	// partial requires [go, git]; task-1 has [go, git] (exact), task-2 has [go, git, make] (superset).
	got, _ := r.RecallSimilar(context.Background(), PartialSignature{
		ContractToolAllow: []string{"go", "git"},
	}, 0)
	if len(got) != 2 {
		t.Errorf("expected 2 superset matches, got %d", len(got))
	}
	// partial requires [go, git, ruby]; nothing matches.
	got, _ = r.RecallSimilar(context.Background(), PartialSignature{
		ContractToolAllow: []string{"go", "git", "ruby"},
	}, 0)
	if len(got) != 0 {
		t.Errorf("expected 0 matches with ruby, got %d", len(got))
	}
}

func TestRecallSimilar_DeprecatedExcluded(t *testing.T) {
	s, r, recs := seedThreePromotions(t)
	ctx := context.Background()

	// Forget task-1, then RecallSimilar by its intent.
	if err := s.Forget(ctx, recs[0].ImmunityID, "fixed", "u"); err != nil {
		t.Fatalf("Forget: %v", err)
	}
	got, _ := r.RecallSimilar(ctx, PartialSignature{IntentKind: "build.binary"}, 0)
	for _, rec := range got {
		if rec.ImmunityID == recs[0].ImmunityID {
			t.Errorf("forgotten record should be excluded from RecallSimilar")
		}
	}
}

func TestRecall_EmptyPartialMatchesAll(t *testing.T) {
	_, r, _ := seedThreePromotions(t)
	// Zero PartialSignature means "any value matches" — should return all 3.
	got, _ := r.RecallSimilar(context.Background(), PartialSignature{}, 0)
	if len(got) != 3 {
		t.Errorf("empty partial should match all 3, got %d", len(got))
	}
}
