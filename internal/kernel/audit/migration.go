package audit

import (
	"fmt"
	"sync"
	"time"
)

// MigrationManager handles data migration between storage tiers.
type MigrationManager struct {
	mu           sync.RWMutex
	manager      *StorageManager
	config       *Config
	stopCh       chan struct{}
	migrationLog []MigrationEvent
}

// MigrationEvent records a migration operation.
type MigrationEvent struct {
	Timestamp  time.Time     `json:"timestamp"`
	FromTier   string        `json:"from_tier"`
	ToTier     string        `json:"to_tier"`
	EntryCount int           `json:"entry_count"`
	Duration   time.Duration `json:"duration"`
	Error      string        `json:"error,omitempty"`
}

// NewMigrationManager creates a new migration manager.
func NewMigrationManager(manager *StorageManager, config *Config) *MigrationManager {
	return &MigrationManager{
		manager:      manager,
		config:       config,
		stopCh:       make(chan struct{}),
		migrationLog: make([]MigrationEvent, 0),
	}
}

// Start begins periodic migration checks.
func (mm *MigrationManager) Start() {
	go mm.migrationLoop()
}

// Stop halts migration background tasks.
func (mm *MigrationManager) Stop() {
	close(mm.stopCh)
}

// MigrateNow triggers an immediate migration.
func (mm *MigrationManager) MigrateNow() error {
	return mm.performMigration()
}

// GetMigrationLog returns the migration history.
func (mm *MigrationManager) GetMigrationLog() []MigrationEvent {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	result := make([]MigrationEvent, len(mm.migrationLog))
	copy(result, mm.migrationLog)
	return result
}

// performMigration executes the migration from hot to cold storage.
func (mm *MigrationManager) performMigration() error {
	start := time.Now()

	cutoff := time.Now().AddDate(0, 0, -mm.config.Hot.MaxAgeDays)
	oldBuckets := mm.manager.hot.GetBucketsBefore(cutoff)

	if len(oldBuckets) == 0 {
		return nil
	}

	totalEntries := 0
	for _, bucket := range oldBuckets {
		totalEntries += len(bucket.entries)
	}

	if err := mm.manager.MigrateNow(); err != nil {
		event := MigrationEvent{
			Timestamp:  start,
			FromTier:   "hot",
			ToTier:     "cold",
			EntryCount: totalEntries,
			Duration:   time.Since(start),
			Error:      err.Error(),
		}
		mm.appendEvent(event)
		return err
	}

	event := MigrationEvent{
		Timestamp:  start,
		FromTier:   "hot",
		ToTier:     "cold",
		EntryCount: totalEntries,
		Duration:   time.Since(start),
	}
	mm.appendEvent(event)

	return nil
}

// migrationLoop runs periodic migration checks.
func (mm *MigrationManager) migrationLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-mm.stopCh:
			return
		case <-ticker.C:
			if err := mm.performMigration(); err != nil {
				fmt.Printf("migration error: %v\n", err)
			}
		}
	}
}

// appendEvent adds an event to the migration log.
func (mm *MigrationManager) appendEvent(event MigrationEvent) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.migrationLog = append(mm.migrationLog, event)

	if len(mm.migrationLog) > 1000 {
		mm.migrationLog = mm.migrationLog[len(mm.migrationLog)-1000:]
	}
}

// GetStats returns migration statistics.
func (mm *MigrationManager) GetStats() MigrationStats {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	stats := MigrationStats{
		TotalMigrations: len(mm.migrationLog),
	}

	for _, event := range mm.migrationLog {
		stats.TotalEntriesMigrated += event.EntryCount
		stats.TotalDuration += event.Duration
		if event.Error != "" {
			stats.FailedMigrations++
		}
	}

	if stats.TotalMigrations > 0 {
		stats.AvgDuration = stats.TotalDuration / time.Duration(stats.TotalMigrations)
	}

	return stats
}

// MigrationStats holds migration statistics.
type MigrationStats struct {
	TotalMigrations      int           `json:"total_migrations"`
	FailedMigrations     int           `json:"failed_migrations"`
	TotalEntriesMigrated int           `json:"total_entries_migrated"`
	TotalDuration        time.Duration `json:"total_duration"`
	AvgDuration          time.Duration `json:"avg_duration"`
}
