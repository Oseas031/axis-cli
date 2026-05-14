package comm

import (
	"testing"
	"time"
)

func testMsg(id, from, to string) Message {
	return Message{ID: id, From: from, To: to, Type: MsgNotify, Payload: map[string]any{"text": "hello"}, Timestamp: time.Now()}
}

func TestSendAndReceive(t *testing.T) {
	mb := NewMailbox(t.TempDir())

	msg := testMsg("1", "agent-a", "agent-b")
	if err := mb.Send(msg); err != nil {
		t.Fatal(err)
	}

	msgs, err := mb.Receive("agent-b")
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].ID != "1" {
		t.Fatalf("unexpected message: %+v", msgs[0])
	}
}

func TestMarkRead(t *testing.T) {
	mb := NewMailbox(t.TempDir())

	mb.Send(testMsg("1", "a", "b"))

	if err := mb.MarkRead("1"); err != nil {
		t.Fatal(err)
	}

	msgs, err := mb.Receive("b")
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages after mark read, got %d", len(msgs))
	}
}

func TestFilterByRecipient(t *testing.T) {
	mb := NewMailbox(t.TempDir())

	mb.Send(testMsg("1", "a", "agent-x"))
	mb.Send(testMsg("2", "a", "agent-y"))
	mb.Send(testMsg("3", "a", "agent-x"))

	msgs, err := mb.Receive("agent-x")
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages for agent-x, got %d", len(msgs))
	}

	msgs, err = mb.Receive("agent-y")
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message for agent-y, got %d", len(msgs))
	}
}
