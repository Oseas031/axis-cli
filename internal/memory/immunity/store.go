package immunity

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/axis-cli/axis/internal/memory/longterm"
)

// EventReader is the subset of longterm.Store that immunity needs for reads.
// Defining it as an interface lets tests use lightweight in-memory fakes.
type EventReader interface {
	QueryEvents(ctx context.Context, filter longterm.EventFilter) ([]longterm.EventRecord, error)
}

// EventAppender is the subset of longterm.Store needed for writes.
type EventAppender interface {
	Append(ctx context.Context, event longterm.EventRecord) error
}

// EventStore combines reader and appender.
type EventStore interface {
	EventReader
	EventAppender
}

// Store is the immunity-memory operations layer. It is a thin derived
// view over the long-term event log; the event log remains the only
// authoritative store (per design D1).
type Store struct {
	events EventStore
	now    func() time.Time // injectable for deterministic tests
}

// NewStore constructs a Store backed by the given longterm event store.
func NewStore(events EventStore) *Store {
	return &Store{events: events, now: time.Now}
}

// Promote registers a failed task as an Immunity record. The source task
// MUST have a terminal task.failed event in the log; promotion of a
// successful task or a non-terminal task is rejected.
func (s *Store) Promote(ctx context.Context, in PromoteInput) (ImmunityRecord, error) {
	if err := in.Validate(); err != nil {
		return ImmunityRecord{}, err
	}

	terminal, err := s.lookupTerminalEvent(ctx, in.SourceTaskID)
	if err != nil {
		return ImmunityRecord{}, err
	}

	class := in.FailureClass
	if class == "" {
		class = deriveFailureClass(terminal)
	}
	if !IsKnownClass(class) {
		// Caller-provided class was validated already; this guards the
		// derive-from-payload path.
		return ImmunityRecord{}, ErrUnknownFailureClass
	}

	intentKind, toolAllow, args := extractSignatureInputs(terminal)
	sig := BuildSignature(intentKind, NormalizeArgs(args), toolAllow, class)
	hash := sig.Hash()

	now := s.now().UTC()
	rec := ImmunityRecord{
		ImmunityID:    formatImmunityID(hash, now),
		SourceTaskID:  in.SourceTaskID,
		Signature:     sig,
		SignatureHash: hash,
		Cause:         strings.TrimSpace(in.Cause),
		FailureClass:  class,
		PromotedBy:    strings.TrimSpace(in.PromotedBy),
		PromotedAt:    now,
	}

	payload := map[string]any{
		"immunity_id":    rec.ImmunityID,
		"source_task_id": rec.SourceTaskID,
		"signature":      rec.Signature,
		"signature_hash": rec.SignatureHash,
		"cause":          rec.Cause,
		"failure_class":  string(rec.FailureClass),
		"promoted_by":    rec.PromotedBy,
	}
	event := longterm.EventRecord{
		EventType: longterm.EventImmunityPromoted,
		EntityID:  rec.ImmunityID,
		Timestamp: now,
		Payload:   payload,
	}
	if err := s.events.Append(ctx, event); err != nil {
		return ImmunityRecord{}, fmt.Errorf("immunity: append promoted event: %w", err)
	}
	return rec, nil
}

// Forget soft-marks an Immunity record deprecated by appending a
// memory.immunity.forgotten event. The original promoted event is
// never mutated.
func (s *Store) Forget(ctx context.Context, immunityID, reason, actor string) error {
	if strings.TrimSpace(immunityID) == "" {
		return ErrImmunityNotFound
	}
	if strings.TrimSpace(actor) == "" {
		return ErrPromotedByRequired
	}
	if _, err := s.Show(ctx, immunityID); err != nil {
		return err
	}
	event := longterm.EventRecord{
		EventType: longterm.EventImmunityForgotten,
		EntityID:  immunityID,
		Timestamp: s.now().UTC(),
		Payload: map[string]any{
			"reason":       strings.TrimSpace(reason),
			"forgotten_by": strings.TrimSpace(actor),
		},
	}
	if err := s.events.Append(ctx, event); err != nil {
		return fmt.Errorf("immunity: append forgotten event: %w", err)
	}
	return nil
}

// Show returns the ImmunityRecord with the given ID. Returns
// ErrImmunityNotFound when the ID has no promoted event.
func (s *Store) Show(ctx context.Context, immunityID string) (ImmunityRecord, error) {
	events, err := s.events.QueryEvents(ctx, longterm.EventFilter{
		EventTypes:        []string{longterm.EventImmunityPromoted, longterm.EventImmunityForgotten},
		EntityID:          immunityID,
		IncludeDeprecated: true,
	})
	if err != nil {
		return ImmunityRecord{}, fmt.Errorf("immunity: query events: %w", err)
	}
	var rec ImmunityRecord
	found := false
	for _, e := range events {
		switch e.EventType {
		case longterm.EventImmunityPromoted:
			r, err := recordFromPromotedEvent(e)
			if err != nil {
				return ImmunityRecord{}, err
			}
			rec = r
			found = true
		case longterm.EventImmunityForgotten:
			if found {
				ts := e.Timestamp
				rec.Deprecated = true
				rec.DeprecatedAt = &ts
				if reason, ok := e.Payload["reason"].(string); ok {
					rec.DeprecateReason = reason
				}
			}
		}
	}
	if !found {
		return ImmunityRecord{}, ErrImmunityNotFound
	}
	return rec, nil
}

// List returns Immunity records matching filter, newest first.
func (s *Store) List(ctx context.Context, filter ListFilter) ([]ImmunityRecord, error) {
	events, err := s.events.QueryEvents(ctx, longterm.EventFilter{
		EventTypes:        []string{longterm.EventImmunityPromoted, longterm.EventImmunityForgotten},
		IncludeDeprecated: true,
	})
	if err != nil {
		return nil, fmt.Errorf("immunity: query events: %w", err)
	}
	byID := make(map[string]*ImmunityRecord)
	order := []string{}
	for _, e := range events {
		switch e.EventType {
		case longterm.EventImmunityPromoted:
			r, err := recordFromPromotedEvent(e)
			if err != nil {
				return nil, err
			}
			rec := r
			byID[r.ImmunityID] = &rec
			order = append(order, r.ImmunityID)
		case longterm.EventImmunityForgotten:
			if rec, ok := byID[e.EntityID]; ok {
				ts := e.Timestamp
				rec.Deprecated = true
				rec.DeprecatedAt = &ts
				if reason, ok := e.Payload["reason"].(string); ok {
					rec.DeprecateReason = reason
				}
			}
		}
	}

	out := make([]ImmunityRecord, 0, len(order))
	for i := len(order) - 1; i >= 0; i-- { // newest first
		rec := byID[order[i]]
		if !filter.IncludeDeprecated && rec.Deprecated {
			continue
		}
		if filter.Class != "" && rec.FailureClass != filter.Class {
			continue
		}
		if filter.Since != nil && rec.PromotedAt.Before(*filter.Since) {
			continue
		}
		out = append(out, *rec)
		if filter.Limit > 0 && len(out) >= filter.Limit {
			break
		}
	}
	return out, nil
}

// lookupTerminalEvent finds the terminal event of a task and validates
// it is a failure. Returns the terminal event or a typed error.
func (s *Store) lookupTerminalEvent(ctx context.Context, taskID string) (longterm.EventRecord, error) {
	events, err := s.events.QueryEvents(ctx, longterm.EventFilter{
		EventTypes:        []string{longterm.EventTaskCompleted, longterm.EventTaskFailed},
		EntityID:          taskID,
		IncludeDeprecated: false,
	})
	if err != nil {
		return longterm.EventRecord{}, fmt.Errorf("immunity: query terminal events: %w", err)
	}
	if len(events) == 0 {
		return longterm.EventRecord{}, ErrTaskNotTerminal
	}
	// Last terminal event wins (multi-terminal is unusual but treat last as authoritative).
	last := events[len(events)-1]
	if last.EventType != longterm.EventTaskFailed {
		return longterm.EventRecord{}, ErrTaskNotFailed
	}
	return last, nil
}

// extractSignatureInputs pulls best-effort intent_kind, tool_allow, args
// from a terminal event's payload. Missing fields yield zero values.
func extractSignatureInputs(e longterm.EventRecord) (intentKind string, toolAllow []string, args map[string]any) {
	if e.Payload == nil {
		return "", nil, nil
	}
	if v, ok := e.Payload["intent_kind"].(string); ok {
		intentKind = v
	}
	if v, ok := e.Payload["contract_tool_allow"].([]any); ok {
		toolAllow = make([]string, 0, len(v))
		for _, t := range v {
			if s, ok := t.(string); ok {
				toolAllow = append(toolAllow, s)
			}
		}
	} else if v, ok := e.Payload["contract_tool_allow"].([]string); ok {
		toolAllow = append(toolAllow, v...)
	}
	if v, ok := e.Payload["intent_args"].(map[string]any); ok {
		args = v
	}
	return intentKind, toolAllow, args
}

// deriveFailureClass tries to read error_class from the failure payload.
// Returns failure.runtime.unknown if the payload has nothing usable.
func deriveFailureClass(e longterm.EventRecord) FailureClass {
	if e.Payload != nil {
		if v, ok := e.Payload["error_class"].(string); ok && v != "" {
			fc := FailureClass(v)
			if IsKnownClass(fc) {
				return fc
			}
		}
	}
	return "failure.runtime.unknown"
}

// recordFromPromotedEvent reconstructs an ImmunityRecord from its
// promoted event payload. JSON round-trips through map[string]any are
// imperfect so we hand-extract critical fields.
func recordFromPromotedEvent(e longterm.EventRecord) (ImmunityRecord, error) {
	if e.Payload == nil {
		return ImmunityRecord{}, fmt.Errorf("immunity: promoted event has nil payload: id=%s", e.EntityID)
	}
	rec := ImmunityRecord{
		ImmunityID: e.EntityID,
		PromotedAt: e.Timestamp,
	}
	if v, ok := e.Payload["source_task_id"].(string); ok {
		rec.SourceTaskID = v
	}
	if v, ok := e.Payload["signature_hash"].(string); ok {
		rec.SignatureHash = v
	}
	if v, ok := e.Payload["cause"].(string); ok {
		rec.Cause = v
	}
	if v, ok := e.Payload["failure_class"].(string); ok {
		rec.FailureClass = FailureClass(v)
	}
	if v, ok := e.Payload["promoted_by"].(string); ok {
		rec.PromotedBy = v
	}
	// signature field may be present as a generic map; we restore the
	// minimum shape needed for List/Show callers. Hash is the source of
	// truth for matching; this is mostly cosmetic for display.
	if sig, ok := e.Payload["signature"].(map[string]any); ok {
		if v, ok := sig["intent_kind"].(string); ok {
			rec.Signature.IntentKind = v
		}
		if v, ok := sig["error_class"].(string); ok {
			rec.Signature.ErrorClass = FailureClass(v)
		}
	} else if sig, ok := e.Payload["signature"].(Signature); ok {
		rec.Signature = sig
	}
	return rec, nil
}

// formatImmunityID produces a stable ID: imm-<first-12-hex>-<unix-ms>.
func formatImmunityID(hash string, t time.Time) string {
	prefix := hash
	if len(prefix) > 12 {
		prefix = prefix[:12]
	}
	return fmt.Sprintf("imm-%s-%d", prefix, t.UnixNano()/int64(time.Millisecond))
}
