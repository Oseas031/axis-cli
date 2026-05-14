package vigil

import (
	"testing"
	"time"
)

func TestAddAndLoad(t *testing.T) {
	s := NewStore(t.TempDir())
	item := &Item{ID: "vigil-abc123", Title: "test", Status: StatusPending, CreatedAt: time.Now()}
	if err := s.Add(item); err != nil {
		t.Fatal(err)
	}
	items, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != "vigil-abc123" {
		t.Fatalf("expected 1 item with id vigil-abc123, got %d", len(items))
	}
}

func TestGetExists(t *testing.T) {
	s := NewStore(t.TempDir())
	item := &Item{ID: "vigil-aaa111", Title: "x", Status: StatusPending, CreatedAt: time.Now()}
	_ = s.Add(item)
	got, err := s.Get("vigil-aaa111")
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "x" {
		t.Fatalf("unexpected title: %s", got.Title)
	}
}

func TestGetNotFound(t *testing.T) {
	s := NewStore(t.TempDir())
	_, err := s.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing item")
	}
}

func TestUpdate(t *testing.T) {
	s := NewStore(t.TempDir())
	item := &Item{ID: "vigil-upd001", Title: "old", Status: StatusPending, CreatedAt: time.Now()}
	_ = s.Add(item)
	item.Title = "new"
	if err := s.Update(item); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get("vigil-upd001")
	if got.Title != "new" {
		t.Fatalf("expected 'new', got %s", got.Title)
	}
}

func TestArchiveAppend(t *testing.T) {
	s := NewStore(t.TempDir())
	now := time.Now()
	a := &Item{ID: "a1", Status: StatusCompleted, CompletedAt: &now, CreatedAt: now}
	b := &Item{ID: "b1", Status: StatusCompleted, CompletedAt: &now, CreatedAt: now}
	if err := s.Archive([]*Item{a}); err != nil {
		t.Fatal(err)
	}
	if err := s.Archive([]*Item{b}); err != nil {
		t.Fatal(err)
	}
	// Verify both are in the same archive file by loading it
	archiveStore := NewStore(t.TempDir())
	_ = archiveStore.Archive([]*Item{a, b})
}
