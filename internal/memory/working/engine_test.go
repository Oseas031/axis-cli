package working

import (
	"context"
	"testing"
	"time"
)

func TestRetainRelease_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	eng, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer eng.Close()

	ctx := context.Background()
	if err := eng.Retain(ctx, "ctx-001", "fix provider config"); err != nil {
		t.Fatalf("Retain: %v", err)
	}

	items, err := eng.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].BundleID != "ctx-001" {
		t.Fatalf("unexpected bundle_id: %q", items[0].BundleID)
	}
	if items[0].Reason != "fix provider config" {
		t.Fatalf("unexpected reason: %q", items[0].Reason)
	}

	if err := eng.Release(ctx, "ctx-001"); err != nil {
		t.Fatalf("Release: %v", err)
	}

	items, err = eng.List(ctx)
	if err != nil {
		t.Fatalf("List after release: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items after release, got %d", len(items))
	}
}

func TestRetain_EmptyBundleID(t *testing.T) {
	dir := t.TempDir()
	eng, _ := Open(dir)
	defer eng.Close()

	if err := eng.Retain(context.Background(), "", "reason"); err != ErrBundleIDEmpty {
		t.Fatalf("expected ErrBundleIDEmpty, got %v", err)
	}
}

func TestRetain_EmptyReason(t *testing.T) {
	dir := t.TempDir()
	eng, _ := Open(dir)
	defer eng.Close()

	if err := eng.Retain(context.Background(), "id", ""); err != ErrReasonEmpty {
		t.Fatalf("expected ErrReasonEmpty, got %v", err)
	}
}

func TestRecall_BasicKeyword(t *testing.T) {
	dir := t.TempDir()
	eng, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer eng.Close()
	ctx := context.Background()

	// Create a bundle with packets.
	bundle := &WorkingBundle{
		BundleID:   "ctx-search",
		Goal:       "implement scheduler",
		ContractID: "default",
		Packets: []ContextPacket{
			{ID: "p1", Type: "spec", Source: "docs/specs/scheduler/design.md", Summary: "Scheduler design doc", Relevance: 0.9},
			{ID: "p2", Type: "code", Source: "internal/kernel/scheduler.go", Summary: "Scheduler implementation", Relevance: 0.85},
		},
		RetainedAt:  timeNow(),
		AccessCount: 1,
	}
	eng.UpdateBundle(ctx, "ctx-search", bundle) //nolint:errcheck

	hits, err := eng.Recall(ctx, "scheduler", 10)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("expected hits for 'scheduler', got none")
	}
	found := false
	for _, h := range hits {
		if h.BundleID == "ctx-search" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected bundle ctx-search in hits")
	}
}

func TestRecall_NoMatch(t *testing.T) {
	dir := t.TempDir()
	eng, _ := Open(dir)
	defer eng.Close()
	ctx := context.Background()

	eng.Retain(ctx, "ctx-nomatch", "unrelated task") //nolint:errcheck
	hits, err := eng.Recall(ctx, "nonexistent", 10)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if len(hits) != 0 {
		t.Fatalf("expected 0 hits, got %d", len(hits))
	}
}

func TestClear(t *testing.T) {
	dir := t.TempDir()
	eng, _ := Open(dir)
	defer eng.Close()
	ctx := context.Background()

	eng.Retain(ctx, "a", "ra") //nolint:errcheck
	eng.Retain(ctx, "b", "rb") //nolint:errcheck
	if err := eng.Clear(ctx); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	items, _ := eng.List(ctx)
	if len(items) != 0 {
		t.Fatalf("expected 0 after clear, got %d", len(items))
	}
}

func TestCompact(t *testing.T) {
	dir := t.TempDir()
	eng, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer eng.Close()
	ctx := context.Background()

	eng.Retain(ctx, "compact-1", "r1") //nolint:errcheck
	eng.Retain(ctx, "compact-2", "r2") //nolint:errcheck
	if err := eng.Compact(); err != nil {
		t.Fatalf("Compact: %v", err)
	}

	items, err := eng.List(ctx)
	if err != nil {
		t.Fatalf("List after compact: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items after compact, got %d", len(items))
	}
}

func TestGetBundle_NotFound(t *testing.T) {
	dir := t.TempDir()
	eng, _ := Open(dir)
	defer eng.Close()

	_, err := eng.GetBundle(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for missing bundle")
	}
}

func timeNow() time.Time {
	return time.Now().UTC()
}
