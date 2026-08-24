package audit

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// ColdStore stores historical audit entries compressed on disk.
type ColdStore struct {
	mu      sync.RWMutex
	rootDir string
	buckets map[string]*ColdBucket
	config  *ColdConfig
}

// ColdBucket represents a compressed time bucket.
type ColdBucket struct {
	month      string // "2026-06"
	filePath   string
	compressed bool
	entryCount int
}

// NewColdStore creates a new cold storage tier.
func NewColdStore(rootDir string, config *ColdConfig) (*ColdStore, error) {
	if err := os.MkdirAll(rootDir, 0750); err != nil {
		return nil, fmt.Errorf("cold store: create dir: %w", err)
	}

	cs := &ColdStore{
		rootDir: rootDir,
		buckets: make(map[string]*ColdBucket),
		config:  config,
	}

	if err := cs.scanExistingBuckets(); err != nil {
		return nil, fmt.Errorf("cold store: scan: %w", err)
	}

	return cs, nil
}

// WriteBucket writes a bucket of entries to cold storage.
func (cs *ColdStore) WriteBucket(month string, entries []AuditEntry) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	filePath := cs.bucketPath(month)

	if cs.config.Compression == "gzip" {
		if err := cs.writeGzipBucket(filePath, entries); err != nil {
			return err
		}
	} else {
		if err := cs.writeRawBucket(filePath, entries); err != nil {
			return err
		}
	}

	cs.buckets[month] = &ColdBucket{
		month:      month,
		filePath:   filePath,
		compressed: cs.config.Compression == "gzip",
		entryCount: len(entries),
	}

	return nil
}

// Query searches cold storage with the given filter.
func (cs *ColdStore) Query(filter AuditFilter) []AuditEntry {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	var results []AuditEntry

	for _, bucket := range cs.buckets {
		if !cs.bucketMatchesFilter(bucket, filter) {
			continue
		}

		entries, err := cs.readBucket(bucket)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if cs.matchesFilter(entry, filter) {
				results = append(results, entry)
			}
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
func (cs *ColdStore) GetByTaskID(taskID string) []AuditEntry {
	return cs.Query(AuditFilter{TaskID: taskID})
}

// Count returns the total number of entries across all buckets.
func (cs *ColdStore) Count() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	total := 0
	for _, bucket := range cs.buckets {
		total += bucket.entryCount
	}
	return total
}

// ListBuckets returns a list of all bucket months.
func (cs *ColdStore) ListBuckets() []string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	months := make([]string, 0, len(cs.buckets))
	for month := range cs.buckets {
		months = append(months, month)
	}
	sort.Strings(months)
	return months
}

// DeleteBucket removes a bucket.
func (cs *ColdStore) DeleteBucket(month string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	bucket, ok := cs.buckets[month]
	if !ok {
		return nil
	}

	if err := os.Remove(bucket.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cold store: delete: %w", err)
	}

	delete(cs.buckets, month)
	return nil
}

// bucketPath returns the file path for a bucket.
func (cs *ColdStore) bucketPath(month string) string {
	return filepath.Join(cs.rootDir, fmt.Sprintf("audit_%s.jsonl.gz", month))
}

// scanExistingBuckets scans the directory for existing bucket files.
func (cs *ColdStore) scanExistingBuckets() error {
	pattern := filepath.Join(cs.rootDir, "audit_*.jsonl*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, path := range matches {
		base := filepath.Base(path)
		month := base[6:13] // Extract "2026-06" from "audit_2026-06.jsonl.gz"

		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		compressed := filepath.Ext(path) == ".gz"
		cs.buckets[month] = &ColdBucket{
			month:      month,
			filePath:   path,
			compressed: compressed,
			entryCount: int(info.Size() / 100), // Approximate
		}
	}

	return nil
}

// writeGzipBucket writes entries to a gzip compressed file.
func (cs *ColdStore) writeGzipBucket(path string, entries []AuditEntry) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cold store: create: %w", err)
	}
	defer file.Close()

	gz := gzip.NewWriter(file)
	defer gz.Close()

	enc := json.NewEncoder(gz)
	for _, entry := range entries {
		if err := enc.Encode(entry); err != nil {
			return fmt.Errorf("cold store: encode: %w", err)
		}
	}

	return nil
}

// writeRawBucket writes entries to an uncompressed file.
func (cs *ColdStore) writeRawBucket(path string, entries []AuditEntry) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cold store: create: %w", err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	for _, entry := range entries {
		if err := enc.Encode(entry); err != nil {
			return fmt.Errorf("cold store: encode: %w", err)
		}
	}

	return nil
}

// readBucket reads entries from a bucket file.
func (cs *ColdStore) readBucket(bucket *ColdBucket) ([]AuditEntry, error) {
	file, err := os.Open(bucket.filePath)
	if err != nil {
		return nil, fmt.Errorf("cold store: open: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file
	if bucket.compressed {
		gz, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("cold store: gzip reader: %w", err)
		}
		defer gz.Close()
		reader = gz
	}

	var entries []AuditEntry
	r := bufio.NewReader(reader)
	dec := json.NewDecoder(r)

	for {
		var entry AuditEntry
		if err := dec.Decode(&entry); err == io.EOF {
			break
		} else if err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// bucketMatchesFilter checks if a bucket might contain matching entries.
func (cs *ColdStore) bucketMatchesFilter(bucket *ColdBucket, filter AuditFilter) bool {
	bucketTime, err := time.Parse("2006-01", bucket.month)
	if err != nil {
		return false
	}

	if !filter.After.IsZero() {
		bucketEnd := bucketTime.AddDate(0, 1, 0)
		if filter.After.After(bucketEnd) {
			return false
		}
	}

	if !filter.Before.IsZero() {
		if filter.Before.Before(bucketTime) {
			return false
		}
	}

	return true
}

// matchesFilter checks if an entry matches the given filter.
func (cs *ColdStore) matchesFilter(entry AuditEntry, filter AuditFilter) bool {
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
