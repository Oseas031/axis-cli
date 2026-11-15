// Package comm provides inter-agent communication primitives.
//
// Inbox is the persistent human-agent scheduling store defined by
// docs/specs/human-agent-scheduling/: one JSONL file per actor under
// .axis/comm/<actor-id>.jsonl, holding that actor's requests together with
// their resolution state. It supersedes the transitional shared-file Mailbox
// for the human flow; the audit trail itself lives in the event log.
package comm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/axis-cli/axis/internal/storeutil"
)

// Inbox lifecycle statuses for InboxRecord.
const (
	InboxStatusPending   = "pending"
	InboxStatusCompleted = "completed"
	InboxStatusFailed    = "failed"
	InboxStatusDegraded  = "degraded"
)

const (
	// DefaultMaxRecordBytes caps a single delivered record; larger deliveries
	// are rejected instead of poisoning the file for every future reader.
	DefaultMaxRecordBytes int64 = 256 << 10 // 256 KiB

	// maxScanLineBytes bounds memory while reading defensively against
	// externally written oversized lines.
	maxScanLineBytes = 16 << 20 // 16 MiB

	legacyMailboxName    = "mailbox.jsonl"
	legacyMigratedSuffix = ".imported"
)

var (
	// ErrNotPending is returned when resolving/degrading a request that is no
	// longer pending. Callers treat it as an idempotent no-op (resolve wins).
	ErrNotPending = errors.New("comm: request is not pending")

	// ErrCallNotFound is returned when the given call ID has no record in the
	// actor's inbox.
	ErrCallNotFound = errors.New("comm: call not found")

	// ErrInvalidActorID rejects actor IDs that could escape the comm directory.
	ErrInvalidActorID = errors.New("comm: invalid actor id")

	// ErrRecordTooLarge is returned by Deliver when the serialized record
	// exceeds Inbox.MaxRecordBytes.
	ErrRecordTooLarge = errors.New("comm: record exceeds size limit")
)

// InboxRecord is one delivered message plus its lifecycle state.
type InboxRecord struct {
	Message     Message        `json:"message"`
	Status      string         `json:"status"`
	Output      map[string]any `json:"output,omitempty"`
	Error       string         `json:"error,omitempty"`
	ResolvedAt  *time.Time     `json:"resolved_at,omitempty"`
	Nudges      int            `json:"nudges,omitempty"`
	LastNudgeAt *time.Time     `json:"last_nudge_at,omitempty"`
}

// Pending reports whether the record still awaits a human response.
func (r *InboxRecord) Pending() bool { return r.Status == InboxStatusPending }

// Inbox is a per-actor persistent request store. All methods are safe for
// concurrent use within one process; cross-process access is prevented by the
// runtime single-instance guard (see docs/specs/human-agent-scheduling/ D2).
type Inbox struct {
	dir            string
	mu             sync.Mutex
	MaxRecordBytes int64

	migrated bool
}

// NewInbox creates an Inbox rooted at dir (typically <project>/.axis/comm).
func NewInbox(dir string) *Inbox {
	return &Inbox{dir: dir, MaxRecordBytes: DefaultMaxRecordBytes}
}

// Dir returns the root directory backing this inbox.
func (i *Inbox) Dir() string { return i.dir }

// ValidateActorID rejects empty ids, path separators, traversal segments and
// characters unsafe for filenames. Security check uses segment matching, not
// prefix matching (CLAUDE.md §10).
func ValidateActorID(actorID string) error {
	if actorID == "" || len(actorID) > 64 {
		return fmt.Errorf("%w: %q", ErrInvalidActorID, actorID)
	}
	if actorID == "." || actorID == ".." {
		return fmt.Errorf("%w: %q", ErrInvalidActorID, actorID)
	}
	for _, r := range actorID {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_' || r == '.':
		default:
			return fmt.Errorf("%w: %q", ErrInvalidActorID, actorID)
		}
	}
	return nil
}

func (i *Inbox) actorPath(actorID string) (string, error) {
	if err := ValidateActorID(actorID); err != nil {
		return "", err
	}
	return filepath.Join(i.dir, actorID+".jsonl"), nil
}

// Deliver appends a record to the actor's inbox file. The record must carry a
// non-empty message ID addressed to the same actor.
func (i *Inbox) Deliver(actorID string, rec *InboxRecord) error {
	path, err := i.actorPath(actorID)
	if err != nil {
		return err
	}
	if rec == nil || rec.Message.ID == "" {
		return fmt.Errorf("comm: delivered record requires a message id")
	}
	if actorID != rec.Message.To {
		return fmt.Errorf("comm: record addressed to %q cannot be delivered to %q", rec.Message.To, actorID)
	}
	if rec.Status == "" {
		rec.Status = InboxStatusPending
	}

	i.mu.Lock()
	defer i.mu.Unlock()
	if err := i.migrateLegacyLocked(); err != nil {
		return err
	}

	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("comm: marshal record: %w", err)
	}
	maxBytes := i.MaxRecordBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxRecordBytes
	}
	if int64(len(data))+1 > maxBytes {
		return fmt.Errorf("%w: %d > %d bytes", ErrRecordTooLarge, len(data)+1, maxBytes)
	}
	return storeutil.AppendJSONL(path, json.RawMessage(data), storeutil.DefaultJSONLOptions())
}

// List returns the actor's records (oldest first, expired included) without
// mutating anything, together with the number of corrupt lines skipped.
func (i *Inbox) List(actorID string) ([]InboxRecord, int, error) {
	path, err := i.actorPath(actorID)
	if err != nil {
		return nil, 0, err
	}

	i.mu.Lock()
	defer i.mu.Unlock()
	if err := i.migrateLegacyLocked(); err != nil {
		return nil, 0, err
	}

	var records []InboxRecord
	corrupted, err := scanRecords(path, func(line []byte) bool {
		var rec InboxRecord
		if err := json.Unmarshal(line, &rec); err != nil {
			return false
		}
		records = append(records, rec)
		return true
	})
	if err != nil {
		return nil, corrupted, err
	}
	return records, corrupted, nil
}

// Get fetches one record by message ID (== call ID envelope).
func (i *Inbox) Get(actorID, msgID string) (*InboxRecord, error) {
	records, _, err := i.List(actorID)
	if err != nil {
		return nil, err
	}
	for idx := range records {
		if records[idx].Message.ID == msgID {
			return &records[idx], nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrCallNotFound, msgID)
}

// MutatePending applies mutate to the named record iff it is still pending,
// rewriting the file atomically. A concurrent resolution wins; the loser gets
// ErrNotPending (resolve-wins-over-degrade arbitration).
func (i *Inbox) MutatePending(actorID, msgID string, mutate func(*InboxRecord)) error {
	path, err := i.actorPath(actorID)
	if err != nil {
		return err
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	records, _, err := readAllRecords(path)
	if err != nil {
		return err
	}
	idx := -1
	for j := range records {
		if records[j].Message.ID == msgID {
			idx = j
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("%w: %s", ErrCallNotFound, msgID)
	}
	if !records[idx].Pending() {
		return fmt.Errorf("%w: %s is %s", ErrNotPending, msgID, records[idx].Status)
	}

	now := time.Now().UTC()
	mutate(&records[idx])
	if records[idx].Status == InboxStatusPending {
		return fmt.Errorf("comm: mutate must transition status for %s", msgID)
	}
	records[idx].ResolvedAt = &now

	data, err := marshalRecords(records)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("comm: write tmp: %w", err)
	}
	if err := storeutil.AtomicReplace(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("comm: replace: %w", err)
	}
	return nil
}

// MutatePendingAllowPending applies mutate regardless of outcome status as
// long as the record exists and is still pending afterwards — used for
// non-terminal bookkeeping like nudge counters. CAS against concurrent
// terminal resolutions is preserved (ErrNotPending if already resolved).
func (i *Inbox) MutatePendingAllowPending(actorID, msgID string, mutate func(*InboxRecord)) error {
	path, err := i.actorPath(actorID)
	if err != nil {
		return err
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	records, _, err := readAllRecords(path)
	if err != nil {
		return err
	}
	idx := -1
	for j := range records {
		if records[j].Message.ID == msgID {
			idx = j
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("%w: %s", ErrCallNotFound, msgID)
	}
	if !records[idx].Pending() {
		return fmt.Errorf("%w: %s is %s", ErrNotPending, msgID, records[idx].Status)
	}

	mutate(&records[idx])
	if !records[idx].Pending() {
		now := time.Now().UTC()
		records[idx].ResolvedAt = &now
	}

	data, err := marshalRecords(records)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("comm: write tmp: %w", err)
	}
	if err := storeutil.AtomicReplace(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("comm: replace: %w", err)
	}
	return nil
}

func marshalRecords(records []InboxRecord) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	for j := range records {
		if err := enc.Encode(records[j]); err != nil {
			return nil, fmt.Errorf("comm: marshal records: %w", err)
		}
	}
	return buf.Bytes(), nil
}

// migrateLegacyLocked performs the one-time move from the transitional shared
// mailbox.jsonl into per-actor files, renaming the source to mailbox.jsonl.imported.
// It is a no-op when no legacy file exists. Caller holds i.mu.
func (i *Inbox) migrateLegacyLocked() error {
	if i.migrated {
		return nil
	}
	i.migrated = true

	legacyPath := filepath.Join(i.dir, legacyMailboxName)
	if _, err := os.Stat(legacyPath); err != nil {
		return nil // nothing to migrate
	}

	msgs, err := readLegacyMessages(legacyPath)
	if err != nil {
		return fmt.Errorf("comm: legacy migration read: %w", err)
	}
	byActor := make(map[string][]InboxRecord)
	for _, m := range msgs {
		if ValidateActorID(m.To) != nil || m.ID == "" {
			continue // undeliverable legacy rows are dropped, counted nowhere
		}
		byActor[m.To] = append(byActor[m.To], InboxRecord{
			Message: m,
			Status:  InboxStatusPending,
		})
	}
	for actor, recs := range byActor {
		path, err := i.actorPath(actor)
		if err != nil {
			continue
		}
		data, err := marshalRecords(recs)
		if err != nil {
			continue
		}
		f, ferr := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
		if ferr != nil {
			continue
		}
		_, _ = f.Write(data)
		_ = f.Close()
	}
	if err := os.Rename(legacyPath, legacyPath+legacyMigratedSuffix); err != nil {
		return fmt.Errorf("comm: legacy migration rename: %w", err)
	}
	return nil
}

func readLegacyMessages(path string) ([]Message, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var msgs []Message
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64<<10), maxScanLineBytes)
	for scanner.Scan() {
		var m Message
		if json.Unmarshal(scanner.Bytes(), &m) == nil && m.ID != "" {
			msgs = append(msgs, m)
		}
	}
	return msgs, scanner.Err()
}

// readAllRecords loads every well-formed record; corrupt lines are skipped.
func readAllRecords(path string) ([]InboxRecord, int, error) {
	var records []InboxRecord
	corrupted, err := scanRecords(path, func(line []byte) bool {
		var rec InboxRecord
		if json.Unmarshal(line, &rec) != nil {
			return false
		}
		records = append(records, rec)
		return true
	})
	return records, corrupted, err
}

// scanRecords feeds each non-empty line to accept; malformed or oversized
// lines increment the corrupted count instead of failing the whole read.
func scanRecords(path string, accept func(line []byte) bool) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("comm: open %s: %w", filepath.Base(path), err)
	}
	defer f.Close()

	corrupted := 0
	r := bufio.NewReader(f)
	for {
		line, rerr := r.ReadBytes('\n')
		if len(line) == 0 {
			if rerr != nil {
				break
			}
			continue
		}
		line = bytes.TrimSuffix(line, []byte("\n"))
		line = bytes.TrimSuffix(line, []byte("\r"))
		if len(bytes.TrimSpace(line)) > 0 {
			if len(line) > maxScanLineBytes || !accept(line) {
				corrupted++
			}
		}
		if rerr != nil {
			break
		}
	}
	return corrupted, nil
}
