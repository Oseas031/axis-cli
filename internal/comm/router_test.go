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

func (a *testActor) ID() string              { return a.id }
func (a *testActor) CommStatus() ActorStatus { return a.status }
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


type mockLogger struct {
	events []CommEvent
}

func (l *mockLogger) LogCommEvent(e CommEvent) {
	l.events = append(l.events, e)
}

func TestRouter_LogsDirectDelivery(t *testing.T) {
	mb := NewMailbox(t.TempDir())
	r := NewRouter(mb)
	log := &mockLogger{}
	r.SetLogger(log)

	a := &testActor{id: "agent-1", status: StatusReady}
	r.Register(a)

	msg := testMsg("m1", "human", "agent-1")
	if err := r.Send(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	if len(log.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(log.events))
	}
	e := log.events[0]
	if e.DeliveryMode != "direct" {
		t.Errorf("expected direct, got %s", e.DeliveryMode)
	}
	if e.MessageID != "m1" || e.From != "human" || e.To != "agent-1" {
		t.Errorf("unexpected event fields: %+v", e)
	}
}

func TestRouter_LogsQueuedOffline(t *testing.T) {
	mb := NewMailbox(t.TempDir())
	r := NewRouter(mb)
	log := &mockLogger{}
	r.SetLogger(log)

	a := &testActor{id: "agent-1", status: StatusOffline}
	r.Register(a)

	msg := testMsg("m2", "human", "agent-1")
	if err := r.Send(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	if len(log.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(log.events))
	}
	e := log.events[0]
	if e.DeliveryMode != "queued" {
		t.Errorf("expected queued, got %s", e.DeliveryMode)
	}
	if e.Reason != "actor offline or busy" {
		t.Errorf("unexpected reason: %s", e.Reason)
	}
}

func TestRouter_LogsQueuedUnregistered(t *testing.T) {
	mb := NewMailbox(t.TempDir())
	r := NewRouter(mb)
	log := &mockLogger{}
	r.SetLogger(log)

	msg := testMsg("m3", "a", "nobody")
	if err := r.Send(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	if len(log.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(log.events))
	}
	if log.events[0].Reason != "actor not registered" {
		t.Errorf("unexpected reason: %s", log.events[0].Reason)
	}
}

func TestRouter_NoLoggerNoPanic(t *testing.T) {
	mb := NewMailbox(t.TempDir())
	r := NewRouter(mb)
	// No logger set — should not panic
	msg := testMsg("m4", "a", "b")
	if err := r.Send(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
}
