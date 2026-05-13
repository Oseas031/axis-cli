package horizon

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/axis-cli/axis/internal/memory/longterm"
)

const defaultDreamLimit = 50

// DreamResult holds the output of a dream replay session.
type DreamResult struct {
	EventsRead  int      `json:"events_read"`
	Clusters    int      `json:"clusters"`
	PatternsNew int      `json:"patterns_new"`
	Skipped     int      `json:"skipped"`
	PatternIDs  []string `json:"pattern_ids,omitempty"`
}

// DreamOptions configures a dream replay.
type DreamOptions struct {
	Since time.Time // Only replay events after this time. Zero = use limit.
	Limit int       // Max events to read (default 50).
}

// Dream replays recent failed events, clusters them, distills patterns,
// and writes results to the horizon store.
func Dream(ctx context.Context, events longterm.Store, store *Store, opts DreamOptions) (*DreamResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = defaultDreamLimit
	}

	filter := longterm.EventFilter{
		EventTypes: []string{longterm.EventTaskFailed},
		Limit:      limit,
	}
	if !opts.Since.IsZero() {
		filter.After = opts.Since
		filter.Limit = 0 // time-based, no count limit
	}

	records, err := events.QueryEvents(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("dream: query events: %w", err)
	}

	if len(records) == 0 {
		return &DreamResult{}, nil
	}

	// Cluster by error message prefix similarity
	clusters := clusterByError(records)

	// Distill each cluster into a pattern
	result := &DreamResult{EventsRead: len(records), Clusters: len(clusters)}
	for prefix, events := range clusters {
		if len(events) < 2 {
			continue // single occurrence, not a pattern
		}
		pattern := distillPattern(prefix, events)
		// Dedup check: skip if similar pattern already exists
		existing, _ := store.Recall(truncate(strings.ToLower(prefix), 40), CategoryPatterns)
		if len(existing) > 0 {
			result.Skipped++
			continue
		}
		if err := store.Store(pattern); err != nil {
			continue
		}
		result.PatternsNew++
		result.PatternIDs = append(result.PatternIDs, pattern.ID)
	}

	return result, nil
}

// clusterByError groups events by the first 60 chars of their error message.
func clusterByError(records []longterm.EventRecord) map[string][]longterm.EventRecord {
	clusters := make(map[string][]longterm.EventRecord)
	for _, r := range records {
		key := extractErrorPrefix(r)
		clusters[key] = append(clusters[key], r)
	}
	return clusters
}

func extractErrorPrefix(r longterm.EventRecord) string {
	msg := ""
	if e, ok := r.Payload["error"].(string); ok {
		msg = e
	} else if e, ok := r.Payload["message"].(string); ok {
		msg = e
	}
	// Normalize: take first 60 chars as cluster key
	msg = strings.TrimSpace(msg)
	if len(msg) > 60 {
		msg = msg[:60]
	}
	if msg == "" {
		msg = "unknown-error"
	}
	return msg
}

func distillPattern(prefix string, events []longterm.EventRecord) Entry {
	// Collect unique task IDs
	taskIDs := make([]string, 0, len(events))
	for _, e := range events {
		taskIDs = append(taskIDs, e.EntityID)
	}

	body := fmt.Sprintf("## Failure Pattern\n\n"+
		"**Error prefix**: %s\n\n"+
		"**Occurrences**: %d\n\n"+
		"**Affected tasks**: %s\n\n"+
		"**Time range**: %s to %s\n\n"+
		"## Suggested Action\n\n"+
		"Investigate root cause of repeated failure with this error pattern.",
		prefix,
		len(events),
		strings.Join(taskIDs, ", "),
		events[0].Timestamp.Format(time.RFC3339),
		events[len(events)-1].Timestamp.Format(time.RFC3339),
	)

	id := fmt.Sprintf("dream-%d", time.Now().UnixMilli())
	return Entry{
		ID:        id,
		Category:  CategoryPatterns,
		Title:     fmt.Sprintf("Repeated failure: %s", truncate(prefix, 40)),
		Tags:      []string{"dream", "failure-pattern", "auto-distilled"},
		CreatedAt: time.Now(),
		Body:      body,
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
