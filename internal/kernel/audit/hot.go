package audit

import (
	"sort"
	"sync"
	"time"
)

// HotStore stores recent audit entries in memory for fast access.
type HotStore struct {
	mu          sync.RWMutex
	entries     []AuditEntry
	timeBuckets map[string]*Bucket
	maxAge      time.Duration
	maxCapacity int
}

// Bucket groups audit entries by time period.
type Bucket struct {
	date    string
	entries []AuditEntry
	index   map[string]int // taskID -> index in entries
}

// NewHotStore creates a new hot storage tier.
func NewHotStore(config *HotConfig) *HotStore {
	maxAge := time.Duration(config.MaxAgeDays) * 24 * time.Hour
	capacity := config.CapacityMB * 1024 / 1024 // Convert to entries (approx)
	if capacity <= 0 {
		capacity = 10000
	}

	return &HotStore{
		entries:     make([]AuditEntry, 0, capacity),
		timeBuckets: make(map[string]*Bucket),
		maxAge:      maxAge,
		maxCapacity: capacity,
	}
}

// Append adds an audit entry to the hot store.
func (h *HotStore) Append(entry AuditEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = append(h.entries, entry)

	bucket := h.getOrCreateBucket(entry.Timestamp)
	bucket.entries = append(bucket.entries, entry)
	if bucket.index == nil {
		bucket.index = make(map[string]int)
	}
	bucket.index[entry.TaskID] = len(bucket.entries) - 1

	h.evictOldEntries()
}

// Query searches the hot store with the given filter.
func (h *HotStore) Query(filter AuditFilter) []AuditEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var results []AuditEntry

	for _, entry := range h.entries {
		if h.matchesFilter(entry, filter) {
			results = append(results, entry)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	if filter.Limit > 0 && len(results) > filter.Limit {
		results = results[:filter.Limit]
	}

	return results
}

// GetByTaskID returns entries matching the given task ID.
func (h *HotStore) GetByTaskID(taskID string) []AuditEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var results []AuditEntry
	for _, entry := range h.entries {
		if entry.TaskID == taskID {
			results = append(results, entry)
		}
	}
	return results
}

// GetRecent returns the N most recent entries.
func (h *HotStore) GetRecent(n int) []AuditEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if n <= 0 || n > len(h.entries) {
		n = len(h.entries)
	}

	start := len(h.entries) - n
	result := make([]AuditEntry, n)
	copy(result, h.entries[start:])

	return result
}

// Count returns the total number of entries.
func (h *HotStore) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.entries)
}

// Clear removes all entries from the hot store.
func (h *HotStore) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = h.entries[:0]
	h.timeBuckets = make(map[string]*Bucket)
}

// GetBucketsBefore returns all buckets with date before the given cutoff.
func (h *HotStore) GetBucketsBefore(cutoff time.Time) []*Bucket {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []*Bucket
	for _, bucket := range h.timeBuckets {
		bucketTime, err := time.Parse("2006-01-02", bucket.date)
		if err != nil {
			continue
		}
		if bucketTime.Before(cutoff) {
			result = append(result, bucket)
		}
	}
	return result
}

// RemoveBucket removes a bucket by date.
func (h *HotStore) RemoveBucket(date string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	bucket, ok := h.timeBuckets[date]
	if !ok {
		return
	}

	// Remove entries from main list
	newEntries := make([]AuditEntry, 0, len(h.entries)-len(bucket.entries))
	for _, entry := range h.entries {
		if entry.Timestamp.Format("2006-01-02") != date {
			newEntries = append(newEntries, entry)
		}
	}
	h.entries = newEntries

	delete(h.timeBuckets, date)
}

// getOrCreateBucket returns or creates a bucket for the given time.
func (h *HotStore) getOrCreateBucket(t time.Time) *Bucket {
	date := t.Format("2006-01-02")
	if bucket, ok := h.timeBuckets[date]; ok {
		return bucket
	}
	bucket := &Bucket{
		date:    date,
		entries: make([]AuditEntry, 0),
		index:   make(map[string]int),
	}
	h.timeBuckets[date] = bucket
	return bucket
}

// evictOldEntries removes entries older than maxAge.
func (h *HotStore) evictOldEntries() {
	if len(h.entries) <= h.maxCapacity {
		return
	}

	cutoff := time.Now().Add(-h.maxAge)
	newEntries := make([]AuditEntry, 0, len(h.entries))
	for _, entry := range h.entries {
		if entry.Timestamp.After(cutoff) {
			newEntries = append(newEntries, entry)
		}
	}
	h.entries = newEntries

	// Rebuild buckets
	h.rebuildBuckets()
}

// rebuildBuckets reconstructs time buckets from entries.
func (h *HotStore) rebuildBuckets() {
	h.timeBuckets = make(map[string]*Bucket)
	for _, entry := range h.entries {
		bucket := h.getOrCreateBucket(entry.Timestamp)
		bucket.entries = append(bucket.entries, entry)
		if bucket.index == nil {
			bucket.index = make(map[string]int)
		}
		bucket.index[entry.TaskID] = len(bucket.entries) - 1
	}
}

// matchesFilter checks if an entry matches the given filter.
func (h *HotStore) matchesFilter(entry AuditEntry, filter AuditFilter) bool {
	if filter.TaskID != "" && entry.TaskID != filter.TaskID {
		return false
	}
	if filter.ExecutorType != "" && entry.ExecutorType != filter.ExecutorType {
		return false
	}
	if filter.Status != "" && entry.Status != filter.Status {
		return false
	}
	if !filter.After.IsZero() && entry.Timestamp.Before(filter.After) {
		return false
	}
	if !filter.Before.IsZero() && entry.Timestamp.After(filter.Before) {
		return false
	}
	return true
}
