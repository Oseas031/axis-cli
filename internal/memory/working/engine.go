package working

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/axis-cli/axis/internal/memory/kv"
)

// Engine is the filesystem-backed implementation of Working Memory.
type Engine struct {
	kv      *kv.Engine
	rootDir string
}

// Open creates or opens a Working Memory engine at rootDir.
func Open(rootDir string) (*Engine, error) {
	kvEng, err := kv.Open(rootDir)
	if err != nil {
		return nil, fmt.Errorf("working: open kv: %w", err)
	}
	return &Engine{kv: kvEng, rootDir: rootDir}, nil
}

// Close closes the engine.
func (e *Engine) Close() error {
	return e.kv.Close()
}

// Retain adds a bundle to the working set.
func (e *Engine) Retain(_ context.Context, bundleID string, reason string) error {
	if bundleID == "" {
		return ErrBundleIDEmpty
	}
	if reason == "" {
		return ErrReasonEmpty
	}

	key := makeBundleKey(bundleID)

	// Check if already retained; if so, update access count and reason.
	existing, err := e.loadBundle(key)
	if err == nil && existing != nil {
		existing.AccessCount++
		existing.Reason = reason
		return e.storeBundle(key, existing)
	}

	// Create new bundle entry (metadata only; value will be filled by caller
	// via UpdateBundle or set during initial creation).
	bundle := &WorkingBundle{
		BundleID:    bundleID,
		RetainedAt:  time.Now().UTC(),
		Reason:      reason,
		AccessCount: 1,
		Packets:     make([]ContextPacket, 0),
	}
	return e.storeBundle(key, bundle)
}

// Release removes a bundle from the working set.
func (e *Engine) Release(_ context.Context, bundleID string) error {
	if bundleID == "" {
		return ErrBundleIDEmpty
	}
	key := makeBundleKey(bundleID)
	return e.kv.Delete(context.Background(), key)
}

// Recall performs a basic keyword match over retained bundle summaries.
// P0: simple case-insensitive strings.Contains on goal + packet summaries.
func (e *Engine) Recall(ctx context.Context, query string, limit int) ([]PacketHit, error) {
	if limit <= 0 {
		limit = 10
	}
	query = strings.ToLower(query)
	var hits []PacketHit

	it, err := e.kv.ScanPrefix(ctx, bundleKeyPrefix)
	if err != nil {
		return nil, fmt.Errorf("working: scan: %w", err)
	}
	defer it.Close()

	for it.Next() {
		if len(hits) >= limit {
			break
		}
		val := it.Value()
		var bundle WorkingBundle
		if err := json.Unmarshal(val, &bundle); err != nil {
			continue // skip malformed
		}

		// Match against goal.
		if strings.Contains(strings.ToLower(bundle.Goal), query) {
			for _, pkt := range bundle.Packets {
				hits = append(hits, PacketHit{
					BundleID:  bundle.BundleID,
					PacketID:  pkt.ID,
					Type:      pkt.Type,
					Source:    pkt.Source,
					Summary:   pkt.Summary,
					Relevance: pkt.Relevance,
				})
				if len(hits) >= limit {
					break
				}
			}
			continue
		}

		// Match against packet summaries.
		for _, pkt := range bundle.Packets {
			if strings.Contains(strings.ToLower(pkt.Summary), query) ||
				strings.Contains(strings.ToLower(pkt.Source), query) {
				hits = append(hits, PacketHit{
					BundleID:  bundle.BundleID,
					PacketID:  pkt.ID,
					Type:      pkt.Type,
					Source:    pkt.Source,
					Summary:   pkt.Summary,
					Relevance: pkt.Relevance,
				})
				if len(hits) >= limit {
					break
				}
			}
		}
	}

	if err := it.Err(); err != nil {
		return nil, fmt.Errorf("working: iterate: %w", err)
	}
	return hits, nil
}

// List returns all retained bundles.
func (e *Engine) List(ctx context.Context) ([]WorkingSetItem, error) {
	var items []WorkingSetItem

	it, err := e.kv.ScanPrefix(ctx, bundleKeyPrefix)
	if err != nil {
		return nil, fmt.Errorf("working: scan: %w", err)
	}
	defer it.Close()

	for it.Next() {
		val := it.Value()
		var bundle WorkingBundle
		if err := json.Unmarshal(val, &bundle); err != nil {
			continue
		}
		items = append(items, WorkingSetItem{
			BundleID:    bundle.BundleID,
			RetainedAt:  bundle.RetainedAt,
			Reason:      bundle.Reason,
			AccessCount: bundle.AccessCount,
		})
	}

	if err := it.Err(); err != nil {
		return nil, fmt.Errorf("working: iterate: %w", err)
	}
	return items, nil
}

// Clear empties the entire working set.
func (e *Engine) Clear(ctx context.Context) error {
	it, err := e.kv.ScanPrefix(ctx, bundleKeyPrefix)
	if err != nil {
		return fmt.Errorf("working: scan for clear: %w", err)
	}
	defer it.Close()

	var keys []string
	for it.Next() {
		keys = append(keys, it.Key())
	}

	for _, key := range keys {
		if err := e.kv.Delete(ctx, key); err != nil {
			return fmt.Errorf("working: delete %q: %w", key, err)
		}
	}
	return nil
}

// Compact triggers explicit snapshot rebuild.
func (e *Engine) Compact() error {
	return e.kv.Compact()
}

// GetBundle retrieves a full bundle by ID. Not part of the Memory interface;
// exposed for callers that need to read/modify bundle contents.
func (e *Engine) GetBundle(ctx context.Context, bundleID string) (*WorkingBundle, error) {
	if bundleID == "" {
		return nil, ErrBundleIDEmpty
	}
	key := makeBundleKey(bundleID)
	return e.loadBundle(key)
}

// UpdateBundle stores a modified bundle. Not part of the Memory interface.
func (e *Engine) UpdateBundle(_ context.Context, bundleID string, bundle *WorkingBundle) error {
	if bundleID == "" {
		return ErrBundleIDEmpty
	}
	if bundle == nil {
		return fmt.Errorf("working: bundle is nil")
	}
	key := makeBundleKey(bundleID)
	return e.storeBundle(key, bundle)
}

// loadBundle unmarshals a bundle from the KV engine.
func (e *Engine) loadBundle(key string) (*WorkingBundle, error) {
	val, err := e.kv.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}
	var bundle WorkingBundle
	if err := json.Unmarshal(val, &bundle); err != nil {
		return nil, fmt.Errorf("working: unmarshal bundle: %w", err)
	}
	return &bundle, nil
}

// storeBundle marshals and stores a bundle into the KV engine.
func (e *Engine) storeBundle(key string, bundle *WorkingBundle) error {
	b, err := json.Marshal(bundle)
	if err != nil {
		return fmt.Errorf("working: marshal bundle: %w", err)
	}
	return e.kv.Put(context.Background(), key, b)
}
