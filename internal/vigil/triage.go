package vigil

import "time"

type TriageResult struct {
	Staled   []string
	Upgraded []string
	Archived []string
}

func Triage(items []*Item, now time.Time) (*TriageResult, []*Item, []*Item) {
	if len(items) == 0 {
		return &TriageResult{}, []*Item{}, []*Item{}
	}

	// Count how many items depend on each ID
	depCount := map[string]int{}
	for _, it := range items {
		for _, d := range it.DependsOn {
			depCount[d]++
		}
	}

	result := &TriageResult{}
	var active, toArchive []*Item

	for _, it := range items {
		// Rule: pending > 7 days → stale
		if it.Status == StatusPending && now.Sub(it.CreatedAt) > 7*24*time.Hour {
			it.Status = StatusStale
			it.History = append(it.History, StatusChange{
				From: StatusPending, To: StatusStale, At: now, Reason: "stale after 7 days",
			})
			result.Staled = append(result.Staled, it.ID)
		}

		// Rule: referenced by ≥3 others → P0
		if depCount[it.ID] >= 3 && it.Priority != "P0" {
			it.Priority = "P0"
			result.Upgraded = append(result.Upgraded, it.ID)
		}

		// Rule: completed > 48h → archive
		if it.Status == StatusCompleted && it.CompletedAt != nil && now.Sub(*it.CompletedAt) > 48*time.Hour {
			toArchive = append(toArchive, it)
			result.Archived = append(result.Archived, it.ID)
		} else {
			active = append(active, it)
		}
	}

	return result, active, toArchive
}
