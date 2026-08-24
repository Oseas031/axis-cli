package audit

import (
	"fmt"
	"time"
)

// StorageManager coordinates hot and cold storage tiers.
type StorageManager struct {
	hot    *HotStore
	cold   *ColdStore
	config *Config
	stopCh chan struct{}
}

// NewStorageManager creates a new tiered storage manager.
func NewStorageManager(config *Config, rootDir string) (*StorageManager, error) {
	hot := NewHotStore(&config.Hot)

	cold, err := NewColdStore(rootDir, &config.Cold)
	if err != nil {
		return nil, fmt.Errorf("storage manager: create cold store: %w", err)
	}

	sm := &StorageManager{
		hot:    hot,
		cold:   cold,
		config: config,
		stopCh: make(chan struct{}),
	}

	return sm, nil
}

// Start begins background migration tasks.
func (sm *StorageManager) Start() {
	go sm.migrationLoop()
}

// Stop halts background tasks.
func (sm *StorageManager) Stop() {
	close(sm.stopCh)
}

// Append adds an audit entry to the hot store.
func (sm *StorageManager) Append(entry AuditEntry) {
	sm.hot.Append(entry)
}

// Query searches both hot and cold stores.
func (sm *StorageManager) Query(filter AuditFilter) ([]AuditEntry, error) {
	hotResults := sm.hot.Query(filter)
	coldResults := sm.cold.Query(filter)

	merged := mergeAuditResults(hotResults, coldResults)

	if filter.Limit > 0 && len(merged) > filter.Limit {
		merged = merged[:filter.Limit]
	}

	return merged, nil
}

// GetByTaskID returns entries matching the given task ID.
func (sm *StorageManager) GetByTaskID(taskID string) ([]AuditEntry, error) {
	return sm.Query(AuditFilter{TaskID: taskID})
}

// GetRecent returns the N most recent entries from hot store.
func (sm *StorageManager) GetRecent(n int) []AuditEntry {
	return sm.hot.GetRecent(n)
}

// Count returns the total number of entries across all tiers.
func (sm *StorageManager) Count() int {
	return sm.hot.Count() + sm.cold.Count()
}

// HotCount returns the number of entries in hot storage.
func (sm *StorageManager) HotCount() int {
	return sm.hot.Count()
}

// ColdCount returns the number of entries in cold storage.
func (sm *StorageManager) ColdCount() int {
	return sm.cold.Count()
}

// MigrateNow triggers an immediate migration from hot to cold.
func (sm *StorageManager) MigrateNow() error {
	return sm.migrateHotToCold()
}

// migrateHotToCold moves old entries from hot to cold storage.
func (sm *StorageManager) migrateHotToCold() error {
	cutoff := time.Now().AddDate(0, 0, -sm.config.Hot.MaxAgeDays)
	oldBuckets := sm.hot.GetBucketsBefore(cutoff)

	for _, bucket := range oldBuckets {
		bucketTime, err := time.Parse("2006-01-02", bucket.date)
		if err != nil {
			continue
		}

		month := bucketTime.Format("2006-01")

		if err := sm.cold.WriteBucket(month, bucket.entries); err != nil {
			return fmt.Errorf("migrate: write cold: %w", err)
		}

		sm.hot.RemoveBucket(bucket.date)
	}

	return nil
}

// migrationLoop runs periodic migrations.
func (sm *StorageManager) migrationLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-sm.stopCh:
			return
		case <-ticker.C:
			if err := sm.migrateHotToCold(); err != nil {
				// Log error but continue
				fmt.Printf("migration error: %v\n", err)
			}
		}
	}
}

// mergeAuditResults combines and sorts results from hot and cold stores.
func mergeAuditResults(hot, cold []AuditEntry) []AuditEntry {
	total := len(hot) + len(cold)
	merged := make([]AuditEntry, 0, total)
	merged = append(merged, hot...)
	merged = append(merged, cold...)

	// Sort by timestamp descending
	sortAuditEntries(merged)

	return merged
}

// sortAuditEntries sorts entries by timestamp.
func sortAuditEntries(entries []AuditEntry) {
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].Timestamp.Before(entries[j].Timestamp) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}

// GetBucketStats returns statistics about cold storage buckets.
func (sm *StorageManager) GetBucketStats() map[string]int {
	stats := make(map[string]int)
	for _, month := range sm.cold.ListBuckets() {
		bucket := sm.cold.buckets[month]
		stats[month] = bucket.entryCount
	}
	return stats
}
