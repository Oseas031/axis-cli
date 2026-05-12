package comm

import (
	"testing"
	"time"
)

func testMsg(id, from, to string) Message {
	return Message{
		ID: id, From: from, To: to,
		Type: MsgNotify, Payload: map[string]any{"text": "hello"},
		Timestamp: time.Now(),
	}
}

func TestMailbox_SendAndPeek(t *testing.T) {
	mb := NewMailbox(t.TempDir())
	msg := testMsg("m1", "a", "b")
	if err := mb.Send(msg); err != nil {
		t.Fatal(err)
	}
	msgs, err := mb.Peek("b")
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 || msgs[0].ID != "m1" {
		t.Errorf("Peek: got %d msgs", len(msgs))
	}
}

func TestMailbox_Receive_Clears(t *testing.T) {
	mb := NewMailbox(t.TempDir())
	mb.Send(testMsg("m1", "a", "b"))
	mb.Send(testMsg("m2", "a", "b"))

	msgs, _ := mb.Receive("b")
	if len(msgs) != 2 {
		t.Fatalf("expected 2, got %d", len(msgs))
	}
	msgs, _ = mb.Peek("b")
	if len(msgs) != 0 {
		t.Errorf("expected empty after Receive, got %d", len(msgs))
	}
}

func TestMailbox_Ack(t *testing.T) {
	mb := NewMailbox(t.TempDir())
	mb.Send(testMsg("m1", "a", "b"))
	mb.Send(testMsg("m2", "a", "b"))
	mb.Send(testMsg("m3", "a", "b"))

	if err := mb.Ack("b", []string{"m2"}); err != nil {
		t.Fatal(err)
	}
	msgs, _ := mb.Peek("b")
	if len(msgs) != 2 {
		t.Fatalf("expected 2 after ack, got %d", len(msgs))
	}
	for _, m := range msgs {
		if m.ID == "m2" {
			t.Error("m2 should have been acked")
		}
	}
}

func TestMailbox_EmptyPeek(t *testing.T) {
	mb := NewMailbox(t.TempDir())
	msgs, err := mb.Peek("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if msgs != nil {
		t.Errorf("expected nil, got %v", msgs)
	}
}
