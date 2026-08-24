// Package audit provides a tiered storage system for audit logs.
// Hot data stays in memory for fast access, cold data is compressed on disk.
package audit

import (
	"time"
)

// AuditEntry records a single dispatch event.
type AuditEntry struct {
	Timestamp    time.Time         `json:"timestamp"`
	TaskID       string            `json:"task_id"`
	ExecutorType string            `json:"executor_type"`
	Duration     time.Duration     `json:"duration"`
	Status       string            `json:"status"`
	Error        string            `json:"error,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// AuditFilter defines query constraints for audit entries.
type AuditFilter struct {
	TaskID       string
	ExecutorType string
	Status       string
	After        time.Time
	Before       time.Time
	Limit        int
	Offset       int
}

// Tier represents a storage tier.
type Tier int

const (
	TierHot  Tier = iota // In-memory, fast access
	TierCold             // Compressed on disk
)

// Config holds configuration for the tiered storage system.
type Config struct {
	Hot   HotConfig   `yaml:"hot"`
	Cold  ColdConfig  `yaml:"cold"`
	Index IndexConfig `yaml:"index"`
	Cache CacheConfig `yaml:"cache"`
}

// HotConfig configures the hot storage tier.
type HotConfig struct {
	MaxAgeDays int    `yaml:"max_age_days"`
	CapacityMB int    `yaml:"capacity_mb"`
	BucketSize string `yaml:"bucket_size"` // "hour", "day"
}

// ColdConfig configures the cold storage tier.
type ColdConfig struct {
	RetentionDays int    `yaml:"retention_days"`
	Compression   string `yaml:"compression"` // "gzip", "snappy"
	BucketSize    string `yaml:"bucket_size"` // "day", "month"
}

// IndexConfig configures indexing.
type IndexConfig struct {
	Enabled bool   `yaml:"enabled"`
	Type    string `yaml:"type"` // "btree", "hash"
	Fanout  int    `yaml:"fanout"`
}

// CacheConfig configures the query cache.
type CacheConfig struct {
	Enabled    bool `yaml:"enabled"`
	SizeMB     int  `yaml:"size_mb"`
	TTLSeconds int  `yaml:"ttl_seconds"`
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		Hot: HotConfig{
			MaxAgeDays: 7,
			CapacityMB: 100,
			BucketSize: "day",
		},
		Cold: ColdConfig{
			RetentionDays: 90,
			Compression:   "gzip",
			BucketSize:    "month",
		},
		Index: IndexConfig{
			Enabled: true,
			Type:    "btree",
			Fanout:  64,
		},
		Cache: CacheConfig{
			Enabled:    true,
			SizeMB:     10,
			TTLSeconds: 300,
		},
	}
}
