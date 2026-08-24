package audit

import (
	"testing"
	"time"
)

func TestHotStore_Append(t *testing.T) {
	config := &HotConfig{
		MaxAgeDays: 7,
		CapacityMB: 10,
		BucketSize: "day",
	}
	store := NewHotStore(config)

	entry := AuditEntry{
		Timestamp:    time.Now(),
		TaskID:       "task-1",
		ExecutorType: "contract",
		Status:       "completed",
	}

	store.Append(entry)

	if store.Count() != 1 {
		t.Errorf("expected 1 entry, got %d", store.Count())
	}
}

func TestHotStore_Query(t *testing.T) {
	config := &HotConfig{
		MaxAgeDays: 7,
		CapacityMB: 10,
		BucketSize: "day",
	}
	store := NewHotStore(config)

	now := time.Now()
	entries := []AuditEntry{
		{Timestamp: now, TaskID: "task-1", ExecutorType: "contract", Status: "completed"},
		{Timestamp: now.Add(-1 * time.Hour), TaskID: "task-2", ExecutorType: "agent", Status: "completed"},
		{Timestamp: now.Add(-2 * time.Hour), TaskID: "task-3", ExecutorType: "contract", Status: "failed"},
	}

	for _, entry := range entries {
		store.Append(entry)
	}

	results := store.Query(AuditFilter{ExecutorType: "contract"})
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	results = store.Query(AuditFilter{TaskID: "task-2"})
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestHotStore_GetRecent(t *testing.T) {
	config := &HotConfig{
		MaxAgeDays: 7,
		CapacityMB: 10,
		BucketSize: "day",
	}
	store := NewHotStore(config)

	now := time.Now()
	for i := 0; i < 10; i++ {
		store.Append(AuditEntry{
			Timestamp: now.Add(-time.Duration(i) * time.Minute),
			TaskID:    "task-" + string(rune('0'+i)),
		})
	}

	recent := store.GetRecent(5)
	if len(recent) != 5 {
		t.Errorf("expected 5 results, got %d", len(recent))
	}
}

func TestColdStore_WriteAndQuery(t *testing.T) {
	config := &ColdConfig{
		RetentionDays: 90,
		Compression:   "gzip",
		BucketSize:    "month",
	}

	store, err := NewColdStore(t.TempDir(), config)
	if err != nil {
		t.Fatalf("failed to create cold store: %v", err)
	}

	now := time.Now()
	entries := []AuditEntry{
		{Timestamp: now, TaskID: "task-1", ExecutorType: "contract", Status: "completed"},
		{Timestamp: now.Add(-1 * time.Hour), TaskID: "task-2", ExecutorType: "agent", Status: "completed"},
	}

	month := now.Format("2006-01")
	if err := store.WriteBucket(month, entries); err != nil {
		t.Fatalf("failed to write bucket: %v", err)
	}

	results := store.Query(AuditFilter{TaskID: "task-1"})
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestStorageManager_AppendAndQuery(t *testing.T) {
	config := DefaultConfig()
	config.Hot.MaxAgeDays = 1
	config.Cold.RetentionDays = 30

	manager, err := NewStorageManager(config, t.TempDir())
	if err != nil {
		t.Fatalf("failed to create storage manager: %v", err)
	}

	now := time.Now()
	entries := []AuditEntry{
		{Timestamp: now, TaskID: "task-1", ExecutorType: "contract", Status: "completed"},
		{Timestamp: now.Add(-1 * time.Hour), TaskID: "task-2", ExecutorType: "agent", Status: "completed"},
		{Timestamp: now.Add(-2 * time.Hour), TaskID: "task-3", ExecutorType: "contract", Status: "failed"},
	}

	for _, entry := range entries {
		manager.Append(entry)
	}

	if manager.HotCount() != 3 {
		t.Errorf("expected 3 hot entries, got %d", manager.HotCount())
	}

	results, err := manager.Query(AuditFilter{Status: "completed"})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestQueryOptimizer_Cache(t *testing.T) {
	config := DefaultConfig()
	config.Hot.MaxAgeDays = 1

	manager, err := NewStorageManager(config, t.TempDir())
	if err != nil {
		t.Fatalf("failed to create storage manager: %v", err)
	}

	optimizer := NewQueryOptimizer(manager, 10)

	now := time.Now()
	manager.Append(AuditEntry{
		Timestamp: now,
		TaskID:    "task-1",
		Status:    "completed",
	})

	filter := AuditFilter{TaskID: "task-1"}

	_, err = optimizer.Query(filter)
	if err != nil {
		t.Fatalf("first query failed: %v", err)
	}

	_, err = optimizer.Query(filter)
	if err != nil {
		t.Fatalf("second query failed: %v", err)
	}

	stats := optimizer.GetStats()
	if stats.CacheHits < 1 {
		t.Errorf("expected at least 1 cache hit, got %d", stats.CacheHits)
	}
}

func TestMigrationManager_MigrateNow(t *testing.T) {
	config := DefaultConfig()
	config.Hot.MaxAgeDays = 0

	manager, err := NewStorageManager(config, t.TempDir())
	if err != nil {
		t.Fatalf("failed to create storage manager: %v", err)
	}

	now := time.Now()
	manager.Append(AuditEntry{
		Timestamp: now.Add(-24 * time.Hour),
		TaskID:    "old-task",
		Status:    "completed",
	})

	migrationManager := NewMigrationManager(manager, config)

	err = migrationManager.MigrateNow()
	if err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if manager.HotCount() != 0 {
		t.Errorf("expected 0 hot entries after migration, got %d", manager.HotCount())
	}

	if manager.ColdCount() != 1 {
		t.Errorf("expected 1 cold entry after migration, got %d", manager.ColdCount())
	}
}
