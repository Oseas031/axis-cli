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
	eng.UpdateBundle(ctx, "ctx-search", bundle)

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

	eng.Retain(ctx, "ctx-nomatch", "unrelated task")
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

	eng.Retain(ctx, "a", "ra")
	eng.Retain(ctx, "b", "rb")
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

	eng.Retain(ctx, "compact-1", "r1")
	eng.Retain(ctx, "compact-2", "r2")
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


func TestRecall_BM25Ranking(t *testing.T) {
	dir := t.TempDir()
	eng, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer eng.Close()
	ctx := context.Background()

	// Bundle 1: highly relevant to "provider config"
	b1 := &WorkingBundle{
		BundleID: "ctx-provider",
		Goal:     "fix provider configuration loading",
		Packets: []ContextPacket{
			{ID: "p1", Type: "spec", Source: "docs/specs/provider/config.md", Summary: "Provider config spec with validation rules"},
			{ID: "p2", Type: "code", Source: "internal/model/provider/registry.go", Summary: "Provider registry manages config profiles"},
		},
		RetainedAt:  timeNow(),
		AccessCount: 1,
	}
	eng.UpdateBundle(ctx, "ctx-provider", b1)

	// Bundle 2: somewhat relevant (mentions "config" but not "provider")
	b2 := &WorkingBundle{
		BundleID: "ctx-scheduler",
		Goal:     "scheduler config timeout tuning",
		Packets: []ContextPacket{
			{ID: "p3", Type: "code", Source: "internal/kernel/scheduler.go", Summary: "Scheduler with configurable timeout"},
		},
		RetainedAt:  timeNow(),
		AccessCount: 1,
	}
	eng.UpdateBundle(ctx, "ctx-scheduler", b2)

	// Bundle 3: irrelevant
	b3 := &WorkingBundle{
		BundleID: "ctx-tool",
		Goal:     "implement bash tool execution",
		Packets: []ContextPacket{
			{ID: "p4", Type: "code", Source: "internal/model/tool/bash.go", Summary: "Bash tool runs shell commands"},
		},
		RetainedAt:  timeNow(),
		AccessCount: 1,
	}
	eng.UpdateBundle(ctx, "ctx-tool", b3)

	// Query: "provider config" should rank ctx-provider first
	hits, err := eng.Recall(ctx, "provider config", 10)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}

	if len(hits) == 0 {
		t.Fatal("expected hits for 'provider config'")
	}

	// First hit should be from ctx-provider (most relevant)
	if hits[0].BundleID != "ctx-provider" {
		t.Fatalf("expected first hit from ctx-provider, got %s", hits[0].BundleID)
	}

	// Relevance scores should be descending
	for i := 1; i < len(hits); i++ {
		if hits[i].Relevance > hits[i-1].Relevance {
			t.Fatalf("hits not sorted by relevance: [%d]=%f > [%d]=%f",
				i, hits[i].Relevance, i-1, hits[i-1].Relevance)
		}
	}

	// ctx-tool (irrelevant) should not appear
	for _, h := range hits {
		if h.BundleID == "ctx-tool" {
			t.Fatal("irrelevant bundle ctx-tool should not appear in results")
		}
	}
}

func TestRecall_EmptyQuery(t *testing.T) {
	dir := t.TempDir()
	eng, _ := Open(dir)
	defer eng.Close()
	ctx := context.Background()

	eng.UpdateBundle(ctx, "b1", &WorkingBundle{
		BundleID: "b1", Goal: "something",
		Packets:    []ContextPacket{{ID: "p1", Summary: "test"}},
		RetainedAt: timeNow(),
	})

	hits, err := eng.Recall(ctx, "", 10)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if hits != nil {
		t.Fatalf("expected nil for empty query, got %d hits", len(hits))
	}
}

func TestRecall_CJKQuery(t *testing.T) {
	dir := t.TempDir()
	eng, _ := Open(dir)
	defer eng.Close()
	ctx := context.Background()

	eng.UpdateBundle(ctx, "b-cn", &WorkingBundle{
		BundleID: "b-cn", Goal: "修复调度器配置",
		Packets: []ContextPacket{
			{ID: "p1", Type: "spec", Source: "docs/scheduler.md", Summary: "调度器超时配置文档"},
		},
		RetainedAt: timeNow(),
	})

	hits, err := eng.Recall(ctx, "调度器", 10)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("expected hits for CJK query '调度器'")
	}
	if hits[0].BundleID != "b-cn" {
		t.Fatalf("expected b-cn, got %s", hits[0].BundleID)
	}
}


func TestRecall_BM25BigramPrecision(t *testing.T) {
	// Verify bigram matching is more precise than unigram-only:
	// "记忆" should match "记忆管理" better than "记录" (which only shares unigram "记")
	dir := t.TempDir()
	eng, _ := Open(dir)
	defer eng.Close()
	ctx := context.Background()

	eng.UpdateBundle(ctx, "b-memory", &WorkingBundle{
		BundleID: "b-memory", Goal: "记忆管理系统设计",
		Packets: []ContextPacket{
			{ID: "p1", Type: "spec", Source: "memory.md", Summary: "长期记忆存储方案"},
		},
		RetainedAt: timeNow(),
	})
	eng.UpdateBundle(ctx, "b-record", &WorkingBundle{
		BundleID: "b-record", Goal: "记录日志系统",
		Packets: []ContextPacket{
			{ID: "p2", Type: "code", Source: "logger.go", Summary: "日志记录模块"},
		},
		RetainedAt: timeNow(),
	})

	hits, err := eng.Recall(ctx, "记忆", 10)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("expected hits for '记忆'")
	}
	// "记忆管理" bundle should rank first (bigram "记忆" exact match)
	if hits[0].BundleID != "b-memory" {
		t.Fatalf("expected b-memory first (bigram precision), got %s (score=%f)",
			hits[0].BundleID, hits[0].Relevance)
	}
}
