// Package comm provides the communication layer for Actor message delivery.
package comm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Mailbox provides persistent async message storage per Actor.
type Mailbox struct {
	dir string
	mu  sync.Mutex
}

// NewMailbox creates a mailbox backed by the given directory.
// Each actor gets a file: <dir>/<actor-id>.jsonl
func NewMailbox(dir string) *Mailbox {
	return &Mailbox{dir: dir}
}

func (m *Mailbox) path(actorID string) string {
	return filepath.Join(m.dir, actorID+".jsonl")
}

// Send appends a message to the recipient's mailbox file.
func (m *Mailbox) Send(msg Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(m.dir, 0o755); err != nil {
		return fmt.Errorf("mailbox: mkdir: %w", err)
	}
	f, err := os.OpenFile(m.path(msg.To), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("mailbox: open: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("mailbox: marshal: %w", err)
	}
	_, err = f.Write(append(data, '\n'))
	return err
}

// Peek returns all messages for an actor without removing them.
func (m *Mailbox) Peek(actorID string) ([]Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.readAll(actorID)
}

// Receive returns all messages and clears the mailbox.
func (m *Mailbox) Receive(actorID string) ([]Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	msgs, err := m.readAll(actorID)
	if err != nil {
		return nil, err
	}
	// Clear the file
	os.Remove(m.path(actorID))
	return msgs, nil
}

// Ack removes specific messages by ID from the mailbox.
func (m *Mailbox) Ack(actorID string, msgIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	msgs, err := m.readAll(actorID)
	if err != nil {
		return err
	}
	idSet := make(map[string]bool, len(msgIDs))
	for _, id := range msgIDs {
		idSet[id] = true
	}
	// Rewrite without acked messages
	var remaining []Message
	for _, msg := range msgs {
		if !idSet[msg.ID] {
			remaining = append(remaining, msg)
		}
	}
	return m.writeAll(actorID, remaining)
}

func (m *Mailbox) readAll(actorID string) ([]Message, error) {
	f, err := os.Open(m.path(actorID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var msgs []Message
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var msg Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue // skip malformed lines
		}
		msgs = append(msgs, msg)
	}
	return msgs, scanner.Err()
}

func (m *Mailbox) writeAll(actorID string, msgs []Message) error {
	if len(msgs) == 0 {
		os.Remove(m.path(actorID))
		return nil
	}
	f, err := os.Create(m.path(actorID))
	if err != nil {
		return err
	}
	defer f.Close()
	for _, msg := range msgs {
		data, _ := json.Marshal(msg)
		f.Write(append(data, '\n'))
	}
	return nil
}
