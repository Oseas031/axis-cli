package actor

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// stubActor is a minimal Actor implementation for testing.
type stubActor struct {
	id       string
	status   ActorStatus
	received []Message
}

func (s *stubActor) ID() string                                  { return s.id }
func (s *stubActor) Status() ActorStatus                         { return s.status }
func (s *stubActor) Receive(ctx context.Context, msg Message) error {
	s.received = append(s.received, msg)
	return nil
}

func TestActorInterface(t *testing.T) {
	var a Actor = &stubActor{id: "test-1", status: ActorReady}
	if a.ID() != "test-1" {
		t.Errorf("ID() = %q", a.ID())
	}
	if a.Status() != ActorReady {
		t.Error("expected ActorReady")
	}
}

func TestMessageJSONRoundTrip(t *testing.T) {
	msg := Message{
		ID:        "msg-1",
		From:      "agent-a",
		To:        "human",
		Type:      MsgTask,
		Payload:   map[string]any{"task_id": "t1", "input": "hello"},
		Timestamp: time.Now().Truncate(time.Second),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID != msg.ID || decoded.From != msg.From || decoded.Type != msg.Type {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

func TestActorReceive(t *testing.T) {
	a := &stubActor{id: "a1", status: ActorReady}
	msg := Message{ID: "m1", From: "b1", To: "a1", Type: MsgNotify}
	if err := a.Receive(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	if len(a.received) != 1 {
		t.Errorf("expected 1 message, got %d", len(a.received))
	}
}
