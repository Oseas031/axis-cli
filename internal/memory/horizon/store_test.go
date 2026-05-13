package horizon

import (
	"testing"
	"time"
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
