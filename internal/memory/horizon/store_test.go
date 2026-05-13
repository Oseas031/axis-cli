package horizon

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/axis-cli/axis/internal/memory/longterm"
)

func TestStore_InitAndStore(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	entry := Entry{
		ID:        "test-pattern-1",
		Category:  CategoryPatterns,
		Title:     "Retry failures cluster around API timeouts",
		Tags:      []string{"retry", "timeout"},
		CreatedAt: time.Now(),
		Body:      "When API calls timeout, retrying immediately fails. Wait 2s before retry.",
	}
	if err := s.Store(entry); err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	results, err := s.Recall("timeout", CategoryPatterns)
	if err != nil {
		t.Fatalf("Recall failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Title != entry.Title {
		t.Errorf("expected title %q, got %q", entry.Title, results[0].Title)
	}
}

func TestStore_RecallNoMatch(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	s.Init()

	s.Store(Entry{ID: "x", Category: CategoryPatterns, Title: "hello", Body: "world", CreatedAt: time.Now()})

	results, _ := s.Recall("nonexistent", CategoryPatterns)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestStore_RecallAllCategories(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	s.Init()

	s.Store(Entry{ID: "p1", Category: CategoryPatterns, Title: "pattern", Body: "shared keyword", CreatedAt: time.Now()})
	s.Store(Entry{ID: "n1", Category: CategoryNarrative, Title: "narrative", Body: "shared keyword", CreatedAt: time.Now()})

	results, _ := s.Recall("shared", "")
	if len(results) != 2 {
		t.Errorf("expected 2 results across categories, got %d", len(results))
	}
}

func TestDream_ClustersAndDistills(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	store.Init()

	events := &mockEventStore{
		events: []longterm.EventRecord{
			{EventType: "task.failed", EntityID: "t1", Timestamp: time.Now(), Payload: map[string]any{"error": "connection timeout to api.example.com:8080/v2/users endpoint failed with network unreachable"}},
			{EventType: "task.failed", EntityID: "t2", Timestamp: time.Now(), Payload: map[string]any{"error": "connection timeout to api.example.com:8080/v2/users endpoint failed with connection reset"}},
			{EventType: "task.failed", EntityID: "t3", Timestamp: time.Now(), Payload: map[string]any{"error": "permission denied: /etc/shadow"}},
		},
	}

	result, err := Dream(context.Background(), events, store, DreamOptions{})
	if err != nil {
		t.Fatalf("Dream failed: %v", err)
	}
	if result.EventsRead != 3 {
		t.Errorf("expected 3 events read, got %d", result.EventsRead)
	}
	// "connection timeout" cluster has 2 events → 1 pattern
	// "permission denied" has 1 event → no pattern (needs >=2)
	if result.PatternsNew != 1 {
		t.Errorf("expected 1 new pattern, got %d", result.PatternsNew)
	}
}

func TestDream_NoEvents(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	store.Init()

	events := &mockEventStore{events: nil}
	result, err := Dream(context.Background(), events, store, DreamOptions{})
	if err != nil {
		t.Fatalf("Dream failed: %v", err)
	}
	if result.EventsRead != 0 || result.PatternsNew != 0 {
		t.Errorf("expected empty result, got %+v", result)
	}
}

func TestBuildPrinciplesPromptSection_WithContent(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	store.Init()

	// Write a principle
	store.Store(Entry{
		ID:        "test-principle",
		Category:  CategoryPrinciples,
		Title:     "Never retry without strategy change",
		Body:      "Repeating the same action expecting different results is wasteful.",
		CreatedAt: time.Now(),
	})

	section := store.BuildPrinciplesPromptSection()
	if section == "" {
		t.Fatal("expected non-empty principles section")
	}
	if !strings.Contains(section, "Repeating the same action") {
		t.Errorf("expected body content in section, got: %s", section)
	}
	if !strings.Contains(section, "Derived principles") {
		t.Errorf("expected header in section, got: %s", section)
	}
}

func TestBuildPrinciplesPromptSection_Empty(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	store.Init()

	section := store.BuildPrinciplesPromptSection()
	if section != "" {
		t.Errorf("expected empty section when no principles, got: %q", section)
	}
}

// mockEventStore implements longterm.Store for testing.
type mockEventStore struct {
	events []longterm.EventRecord
}

func (m *mockEventStore) Append(ctx context.Context, event longterm.EventRecord) error { return nil }
func (m *mockEventStore) QueryEvents(ctx context.Context, filter longterm.EventFilter) ([]longterm.EventRecord, error) {
	return m.events, nil
}
func (m *mockEventStore) MarkForgotten(ctx context.Context, entityID string, at time.Time) error {
	return nil
}
func (m *mockEventStore) Close() error { return nil }

func TestForget_ArchivesOldNarrative(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	s.Init()

	s.Store(Entry{
		ID:        "old-narrative",
		Category:  CategoryNarrative,
		Title:     "Something old",
		Body:      "Original body content",
		CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
	})

	result, err := Forget(s, false)
	if err != nil {
		t.Fatalf("Forget failed: %v", err)
	}
	if result.Archived != 1 {
		t.Errorf("expected 1 archived, got %d", result.Archived)
	}

	entries, _ := s.Recall("", CategoryNarrative)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Body != "[archived]" {
		t.Errorf("expected body [archived], got %q", entries[0].Body)
	}
}

func TestForget_DeletesVeryOld(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	s.Init()

	s.Store(Entry{
		ID:        "very-old",
		Category:  CategoryNarrative,
		Title:     "Ancient",
		Body:      "Should be deleted",
		CreatedAt: time.Now().Add(-35 * 24 * time.Hour),
	})

	result, err := Forget(s, false)
	if err != nil {
		t.Fatalf("Forget failed: %v", err)
	}
	if result.Deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", result.Deleted)
	}

	entries, _ := s.Recall("", CategoryNarrative)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after delete, got %d", len(entries))
	}
}

func TestForget_SkipsRecent(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	s.Init()

	s.Store(Entry{
		ID:        "recent",
		Category:  CategoryNarrative,
		Title:     "Fresh",
		Body:      "Should stay",
		CreatedAt: time.Now().Add(-1 * 24 * time.Hour),
	})

	result, err := Forget(s, false)
	if err != nil {
		t.Fatalf("Forget failed: %v", err)
	}
	if result.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", result.Skipped)
	}

	entries, _ := s.Recall("Should stay", CategoryNarrative)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry untouched, got %d", len(entries))
	}
}

func TestForget_NeverTouchesPatterns(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	s.Init()

	s.Store(Entry{
		ID:        "old-pattern",
		Category:  CategoryPatterns,
		Title:     "Old pattern",
		Body:      "Pattern body",
		CreatedAt: time.Now().Add(-35 * 24 * time.Hour),
	})

	result, err := Forget(s, false)
	if err != nil {
		t.Fatalf("Forget failed: %v", err)
	}
	if result.Archived != 0 || result.Deleted != 0 {
		t.Errorf("expected no changes to patterns, got archived=%d deleted=%d", result.Archived, result.Deleted)
	}

	entries, _ := s.Recall("Pattern body", CategoryPatterns)
	if len(entries) != 1 {
		t.Errorf("expected pattern untouched, got %d entries", len(entries))
	}
}

func TestForget_DryRun(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	s.Init()

	s.Store(Entry{
		ID:        "dry-run-target",
		Category:  CategoryNarrative,
		Title:     "Will be archived?",
		Body:      "Original content",
		CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
	})

	result, err := Forget(s, true)
	if err != nil {
		t.Fatalf("Forget failed: %v", err)
	}
	if result.Archived != 1 {
		t.Errorf("expected 1 archived in dry-run, got %d", result.Archived)
	}

	// Verify file was NOT modified
	entries, _ := s.Recall("Original content", CategoryNarrative)
	if len(entries) != 1 {
		t.Errorf("expected file untouched in dry-run, got %d entries", len(entries))
	}
	if entries[0].Body != "Original content" {
		t.Errorf("expected original body preserved in dry-run, got %q", entries[0].Body)
	}
}
