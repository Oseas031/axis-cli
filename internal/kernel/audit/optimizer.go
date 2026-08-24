package audit

import (
	"crypto/sha256"
	"encoding/json"
	"sync"
	"time"
)

// QueryOptimizer optimizes audit queries with caching.
type QueryOptimizer struct {
	mu      sync.RWMutex
	manager *StorageManager
	cache   *LRUCache
	stats   QueryStats
}

// QueryStats tracks query statistics.
type QueryStats struct {
	TotalQueries   int64
	CacheHits      int64
	CacheMisses    int64
	AvgQueryTimeMs float64
}

// LRUCache implements a simple LRU cache.
type LRUCache struct {
	mu       sync.RWMutex
	capacity int
	items    map[string]*cacheEntry
	order    []string
}

type cacheEntry struct {
	value     []AuditEntry
	timestamp time.Time
	ttl       time.Duration
}

// NewQueryOptimizer creates a new query optimizer.
func NewQueryOptimizer(manager *StorageManager, cacheSizeMB int) *QueryOptimizer {
	capacity := cacheSizeMB * 1024 / 1024 // Approximate entry count
	if capacity <= 0 {
		capacity = 1000
	}

	return &QueryOptimizer{
		manager: manager,
		cache: &LRUCache{
			capacity: capacity,
			items:    make(map[string]*cacheEntry),
			order:    make([]string, 0),
		},
	}
}

// Query executes an optimized audit query.
func (qo *QueryOptimizer) Query(filter AuditFilter) ([]AuditEntry, error) {
	start := time.Now()
	qo.stats.TotalQueries++

	key := qo.cacheKey(filter)

	if cached := qo.cache.Get(key); cached != nil {
		qo.stats.CacheHits++
		qo.updateStats(time.Since(start))
		return cached, nil
	}

	qo.stats.CacheMisses++

	results, err := qo.manager.Query(filter)
	if err != nil {
		return nil, err
	}

	qo.cache.Set(key, results, 5*time.Minute)
	qo.updateStats(time.Since(start))

	return results, nil
}

// GetByTaskID returns entries for a specific task ID.
func (qo *QueryOptimizer) GetByTaskID(taskID string) ([]AuditEntry, error) {
	return qo.Query(AuditFilter{TaskID: taskID})
}

// GetRecent returns the N most recent entries.
func (qo *QueryOptimizer) GetRecent(n int) []AuditEntry {
	return qo.manager.GetRecent(n)
}

// GetStats returns query statistics.
func (qo *QueryOptimizer) GetStats() QueryStats {
	qo.mu.RLock()
	defer qo.mu.RUnlock()
	return qo.stats
}

// InvalidateCache clears the cache.
func (qo *QueryOptimizer) InvalidateCache() {
	qo.cache.Clear()
}

// cacheKey generates a cache key for a filter.
func (qo *QueryOptimizer) cacheKey(filter AuditFilter) string {
	data, _ := json.Marshal(filter)
	hash := sha256.Sum256(data)
	return string(hash[:])
}

// updateStats updates average query time.
func (qo *QueryOptimizer) updateStats(d time.Duration) {
	qo.mu.Lock()
	defer qo.mu.Unlock()

	ms := float64(d.Milliseconds())
	qo.stats.AvgQueryTimeMs = (qo.stats.AvgQueryTimeMs*float64(qo.stats.TotalQueries-1) + ms) / float64(qo.stats.TotalQueries)
}

// Get returns a cached value if present and not expired.
func (c *LRUCache) Get(key string) []AuditEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.items[key]
	if !ok {
		return nil
	}

	if time.Since(entry.timestamp) > entry.ttl {
		delete(c.items, key)
		c.removeFromOrder(key)
		return nil
	}

	return entry.value
}

// Set adds or updates a cache entry.
func (c *LRUCache) Set(key string, value []AuditEntry, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.items[key]; ok {
		c.items[key] = &cacheEntry{
			value:     value,
			timestamp: time.Now(),
			ttl:       ttl,
		}
		c.moveToEnd(key)
		return
	}

	if len(c.items) >= c.capacity {
		c.evict()
	}

	c.items[key] = &cacheEntry{
		value:     value,
		timestamp: time.Now(),
		ttl:       ttl,
	}
	c.order = append(c.order, key)
}

// Clear removes all cache entries.
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheEntry)
	c.order = c.order[:0]
}

// evict removes the least recently used entry.
func (c *LRUCache) evict() {
	if len(c.order) == 0 {
		return
	}

	oldest := c.order[0]
	delete(c.items, oldest)
	c.order = c.order[1:]
}

// removeFromOrder removes a key from the order list.
func (c *LRUCache) removeFromOrder(key string) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			return
		}
	}
}

// moveToEnd moves a key to the end of the order list.
func (c *LRUCache) moveToEnd(key string) {
	c.removeFromOrder(key)
	c.order = append(c.order, key)
}
