// Package longterm provides the Long-term Memory layer for Axis.
// It stores immutable event records and derives queryable views.
package longterm

import (
	"errors"
	"time"
)

// Event types (initial set).
const (
	EventTaskCreated        = "task.created"
	EventTaskCompleted      = "task.completed"
	EventTaskFailed         = "task.failed"
	EventJudgementSubmitted = "judgement.submitted"
	EventAutonomyTransition = "autonomy.transitioned"
	EventMemoryRetained     = "memory.retained"
	EventMemoryReleased     = "memory.released"
	EventMemoryForgotten    = "memory.forgotten"
	EventToolExecuted       = "tool.executed"

	// EventImmunityPromoted is appended by internal/memory/immunity.Store.Promote
	// when a failed task is explicitly promoted to an Immunity record. See
	// docs/specs/immunity-memory/.
	EventImmunityPromoted = "memory.immunity.promoted"

	// EventImmunityForgotten is appended by internal/memory/immunity.Store.Forget
	// when an Immunity record is soft-marked deprecated. The original
	// EventImmunityPromoted record is never mutated.
	EventImmunityForgotten = "memory.immunity.forgotten"
)

var (
	// ErrEventTypeEmpty is returned when event_type is empty.
	ErrEventTypeEmpty = errors.New("longterm: event_type is empty")
	// ErrEntityIDEmpty is returned when entity_id is empty.
	ErrEntityIDEmpty = errors.New("longterm: entity_id is empty")
)

// EventRecord is the immutable unit of Long-term Memory.
type EventRecord struct {
	EventType    string         `json:"event_type"`
	EntityID     string         `json:"entity_id"`
	Timestamp    time.Time      `json:"timestamp"`
	Payload      map[string]any `json:"payload,omitempty"`
	SourceDigest string         `json:"source_digest,omitempty"`
	DeprecatedAt *time.Time     `json:"deprecated_at,omitempty"`
}

// EventFilter defines query constraints for events.
type EventFilter struct {
	EventTypes        []string
	EntityID          string
	After             time.Time
	Before            time.Time
	IncludeDeprecated bool
	Limit             int
}

// CompetenceProfile is a derived view from judgement and transition events.
type CompetenceProfile struct {
	AgentID       string             `json:"agent_id"`
	ProjectID     string             `json:"project_id"`
	LatestScore   float64            `json:"latest_score"`
	AutonomyLevel string             `json:"autonomy_level"`
	History       []CompetenceSample `json:"history"`
	SourceDigest  string             `json:"source_digest"`
}

// CompetenceSample is a single competence data point.
type CompetenceSample struct {
	Timestamp time.Time          `json:"timestamp"`
	Score     float64            `json:"score"`
	Criteria  map[string]float64 `json:"criteria,omitempty"`
}
