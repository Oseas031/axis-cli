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


// DreamSuccess replays recent successful task events, clusters them by
// task type/pattern, and distills recurring success patterns into principles.
// Trigger: same task type succeeds ≥ 3 times.
//
// // aspirational: 当前 task.completed 事件缺乏"为什么成功"的信息（策略、tool序列、context组合），
// // 产出的 principle 质量有限。待 EventRecord.Payload 包含 execution_summary 后激活完整提炼。
func DreamSuccess(ctx context.Context, events longterm.Store, store *Store, opts DreamOptions) (*DreamResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = defaultDreamLimit
	}

	filter := longterm.EventFilter{
		EventTypes: []string{longterm.EventTaskCompleted},
		Limit:      limit,
	}
	if !opts.Since.IsZero() {
		filter.After = opts.Since
		filter.Limit = 0
	}

	records, err := events.QueryEvents(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("dream-success: query events: %w", err)
	}

	if len(records) == 0 {
		return &DreamResult{}, nil
	}

	// Cluster by task type (from payload["task_type"] or first word of entity_id)
	clusters := clusterByTaskType(records)

	result := &DreamResult{EventsRead: len(records), Clusters: len(clusters)}
	for taskType, evts := range clusters {
		if len(evts) < 3 {
			continue // need ≥ 3 successes to form a principle
		}
		principle := distillPrinciple(taskType, evts)
		// Dedup: skip if similar principle already exists
		existing, _ := store.Recall(truncate(strings.ToLower(taskType), 40), CategoryPrinciples)
		if len(existing) > 0 {
			result.Skipped++
			continue
		}
		if err := store.Store(principle); err != nil {
			continue
		}
		result.PatternsNew++
		result.PatternIDs = append(result.PatternIDs, principle.ID)
	}

	return result, nil
}

// clusterByTaskType groups completed events by their task type.
func clusterByTaskType(records []longterm.EventRecord) map[string][]longterm.EventRecord {
	clusters := make(map[string][]longterm.EventRecord)
	for _, r := range records {
		key := extractTaskType(r)
		clusters[key] = append(clusters[key], r)
	}
	return clusters
}

func extractTaskType(r longterm.EventRecord) string {
	if t, ok := r.Payload["task_type"].(string); ok && t != "" {
		return t
	}
	// Fallback: use entity_id prefix (before first dash or underscore)
	id := r.EntityID
	for i, c := range id {
		if c == '-' || c == '_' {
			if i > 0 {
				return id[:i]
			}
		}
	}
	if len(id) > 20 {
		return id[:20]
	}
	if id == "" {
		return "unknown"
	}
	return id
}

func distillPrinciple(taskType string, events []longterm.EventRecord) Entry {
	taskIDs := make([]string, 0, len(events))
	for _, e := range events {
		taskIDs = append(taskIDs, e.EntityID)
	}

	body := fmt.Sprintf("## Success Pattern\n\n"+
		"**Task type**: %s\n\n"+
		"**Successes**: %d\n\n"+
		"**Tasks**: %s\n\n"+
		"**Time range**: %s to %s\n\n"+
		"## Derived Principle\n\n"+
		"This task type has a reliable success pattern. "+
		"Reuse the approach from these successful executions when encountering similar tasks.",
		taskType,
		len(events),
		strings.Join(taskIDs, ", "),
		events[0].Timestamp.Format(time.RFC3339),
		events[len(events)-1].Timestamp.Format(time.RFC3339),
	)

	id := fmt.Sprintf("principle-%d", time.Now().UnixMilli())
	return Entry{
		ID:        id,
		Category:  CategoryPrinciples,
		Title:     fmt.Sprintf("Success pattern: %s", truncate(taskType, 40)),
		Tags:      []string{"dream-success", "auto-distilled", "principle"},
		CreatedAt: time.Now(),
		Body:      body,
	}
}
