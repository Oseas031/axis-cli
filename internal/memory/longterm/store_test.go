package longterm

import (
	"context"
	"testing"
	"time"
)

func TestAppendQuery_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	e := EventRecord{
		EventType: EventTaskCreated,
		EntityID:  "task-001",
		Payload:   map[string]any{"goal": "test"},
	}
	if err := s.Append(ctx, e); err != nil {
		t.Fatalf("Append: %v", err)
	}

	results, err := s.QueryEvents(ctx, EventFilter{})
	if err != nil {
		t.Fatalf("QueryEvents: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].EntityID != "task-001" {
		t.Fatalf("unexpected entity_id: %q", results[0].EntityID)
	}
}

func TestAppend_EmptyEventType(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir)
	defer s.Close()
	err := s.Append(context.Background(), EventRecord{EntityID: "x"})
	if err != ErrEventTypeEmpty {
		t.Fatalf("expected ErrEventTypeEmpty, got %v", err)
	}
}

func TestAppend_EmptyEntityID(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir)
	defer s.Close()
	err := s.Append(context.Background(), EventRecord{EventType: EventTaskCreated})
	if err != ErrEntityIDEmpty {
		t.Fatalf("expected ErrEntityIDEmpty, got %v", err)
	}
}

func TestQueryEvents_Filtering(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir)
	defer s.Close()
	ctx := context.Background()

	s.Append(ctx, EventRecord{EventType: EventTaskCreated, EntityID: "a"})
	s.Append(ctx, EventRecord{EventType: EventTaskCompleted, EntityID: "a"})
	s.Append(ctx, EventRecord{EventType: EventTaskCreated, EntityID: "b"})

	// Filter by event type.
	results, _ := s.QueryEvents(ctx, EventFilter{EventTypes: []string{EventTaskCreated}})
	if len(results) != 2 {
		t.Fatalf("expected 2 created events, got %d", len(results))
	}

	// Filter by entity.
	results, _ = s.QueryEvents(ctx, EventFilter{EntityID: "a"})
	if len(results) != 2 {
		t.Fatalf("expected 2 events for entity a, got %d", len(results))
	}

	// Combined.
	results, _ = s.QueryEvents(ctx, EventFilter{EventTypes: []string{EventTaskCompleted}, EntityID: "a"})
	if len(results) != 1 {
		t.Fatalf("expected 1 completed event for a, got %d", len(results))
	}
}

func TestForget_SoftMark(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir)
	defer s.Close()
	ctx := context.Background()

	s.Append(ctx, EventRecord{EventType: EventMemoryRetained, EntityID: "bundle-001"})
	if err := s.MarkForgotten(ctx, "bundle-001", time.Now().UTC()); err != nil {
		t.Fatalf("MarkForgotten: %v", err)
	}

	// Default query: forget event excluded.
	results, _ := s.QueryEvents(ctx, EventFilter{EventTypes: []string{EventMemoryForgotten}})
	if len(results) != 0 {
		t.Fatalf("expected 0 forgotten events by default, got %d", len(results))
	}

	// Include deprecated.
	results, _ = s.QueryEvents(ctx, EventFilter{
		EventTypes:        []string{EventMemoryForgotten},
		IncludeDeprecated: true,
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 forgotten event with IncludeDeprecated, got %d", len(results))
	}

	// Original retained event still accessible.
	results, _ = s.QueryEvents(ctx, EventFilter{EventTypes: []string{EventMemoryRetained}})
	if len(results) != 1 {
		t.Fatalf("original retained event should survive: got %d", len(results))
	}
}

func TestQueryEvents_Limit(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir)
	defer s.Close()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		s.Append(ctx, EventRecord{EventType: EventTaskCreated, EntityID: "x"})
	}

	results, _ := s.QueryEvents(ctx, EventFilter{Limit: 3})
	if len(results) != 3 {
		t.Fatalf("expected 3 results with limit, got %d", len(results))
	}
}
