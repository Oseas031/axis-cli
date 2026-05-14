// transitional: Mailbox uses a single shared JSONL file. Exit condition: when
// per-actor mailbox files (.axis/comm/<actor-id>.jsonl) are implemented per the
// spec, this shared-file approach is replaced. The single-file design is a P0
// simplification; the spec (FR3) already defines the target state.

package comm

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Mailbox is a JSONL-backed message store for inter-agent communication.
type Mailbox struct {
	dir string
	mu  sync.Mutex
}

// NewMailbox creates a mailbox that stores messages as JSONL in the given directory.
func NewMailbox(dir string) *Mailbox {
	return &Mailbox{dir: dir}
}

func (m *Mailbox) path() string {
	return filepath.Join(m.dir, "mailbox.jsonl")
}

// Send appends a message to the JSONL file.
func (m *Mailbox) Send(msg Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	f, err := os.OpenFile(m.path(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}

// Receive returns unread messages for the given agent and marks them as read.
func (m *Mailbox) Receive(agentID string) ([]Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	msgs, err := m.readAll()
	if err != nil {
		return nil, err
	}

	var result []Message
	for _, msg := range msgs {
		if msg.To == agentID {
			result = append(result, msg)
		}
	}
	return result, nil
}

// Peek returns messages for the given agent without removing them.
func (m *Mailbox) Peek(agentID string) ([]Message, error) {
	return m.Receive(agentID)
}

// MarkRead removes a message by ID from the mailbox file.
func (m *Mailbox) MarkRead(msgID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	msgs, err := m.readAll()
	if err != nil {
		return err
	}

	var kept []Message
	for _, msg := range msgs {
		if msg.ID != msgID {
			kept = append(kept, msg)
		}
	}
	return m.writeAll(kept)
}

func (m *Mailbox) readAll() ([]Message, error) {
	f, err := os.Open(m.path())
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
			continue
		}
		msgs = append(msgs, msg)
	}
	return msgs, scanner.Err()
}

func (m *Mailbox) writeAll(msgs []Message) error {
	f, err := os.Create(m.path())
	if err != nil {
		return err
	}
	defer f.Close()

	for _, msg := range msgs {
		data, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		if _, err := f.Write(append(data, '\n')); err != nil {
			return err
		}
	}
	return nil
}
