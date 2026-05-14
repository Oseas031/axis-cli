package vigil

import (
	"crypto/sha256"
	"fmt"
	"time"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusStale      Status = "stale"
)

type Origin struct {
	Type string `json:"type"`
	Ref  string `json:"ref"`
}

type StatusChange struct {
	From   Status    `json:"from"`
	To     Status    `json:"to"`
	At     time.Time `json:"at"`
	Reason string    `json:"reason,omitempty"`
}

type Item struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Priority    string         `json:"priority"`
	Status      Status         `json:"status"`
	Tags        []string       `json:"tags"`
	Origin      Origin         `json:"origin"`
	DependsOn   []string       `json:"depends_on"`
	Notes       string         `json:"notes,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	CommitHash  string         `json:"commit_hash,omitempty"`
	History     []StatusChange `json:"history"`
}

func GenerateID(title string, createdAt time.Time) string {
	h := sha256.Sum256([]byte(title + createdAt.String()))
	return fmt.Sprintf("vigil-%x", h[:3])
}

func (it *Item) IsActive() bool {
	return it.Status == StatusPending || it.Status == StatusInProgress
}
