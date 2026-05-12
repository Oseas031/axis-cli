package comm

import (
	"context"
	"testing"
)

type testActor struct {
	id       string
	status   ActorStatus
	received []Message
}

func (a *testActor) ID() string                                      { return a.id }
func (a *testActor) CommStatus() ActorStatus                         { return a.status }
func (a *testActor) Receive(ctx context.Context, msg Message) error {
	a.received = append(a.received, msg)
	return nil
}

func TestRouter_DirectDelivery(t *testing.T) {
	mb := NewMailbox(t.TempDir())
	r := NewRouter(mb)

	a := &testActor{id: "agent-1", status: StatusReady}
	r.Register(a)

	msg := testMsg("m1", "human", "agent-1")
	if err := r.Send(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	if len(a.received) != 1 {
		t.Errorf("expected direct delivery, got %d", len(a.received))
	}
}

func TestRouter_OfflineQueues(t *testing.T) {
	mb := NewMailbox(t.TempDir())
	r := NewRouter(mb)

	a := &testActor{id: "human", status: StatusOffline}
	r.Register(a)

	msg := testMsg("m1", "agent-1", "human")
	if err := r.Send(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	if len(a.received) != 0 {
		t.Error("offline actor should not receive directly")
	}
	msgs, _ := mb.Peek("human")
	if len(msgs) != 1 {
		t.Errorf("expected 1 in mailbox, got %d", len(msgs))
	}
}

func TestRouter_UnregisteredQueues(t *testing.T) {
	mb := NewMailbox(t.TempDir())
	r := NewRouter(mb)

	msg := testMsg("m1", "a", "unknown-actor")
	if err := r.Send(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	msgs, _ := mb.Peek("unknown-actor")
	if len(msgs) != 1 {
		t.Errorf("expected 1 in mailbox, got %d", len(msgs))
	}
}
