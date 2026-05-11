package immediate

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/axis-cli/axis/internal/memory/working"
)

// seenFileName is the per-project seen-hash tracking file.
const seenFileName = ".seen"

// ContextBuilder assembles ImmediateContext from Working Memory and filesystem.
type ContextBuilder struct {
	projectRoot string
	seenPath    string
	seenCache   map[string]seenEntry // in-memory cache of .seen file
	seenDirty   bool
}

// seenEntry tracks the last observed hash and timestamp for a file.
type seenEntry struct {
	Hash      string
	Timestamp int64
}

// NewContextBuilder creates a builder for the given project root.
// The .seen file is stored at projectRoot/.axis/memory/.seen.
func NewContextBuilder(projectRoot string) *ContextBuilder {
	memDir := filepath.Join(projectRoot, ".axis", "memory")
	return &ContextBuilder{
		projectRoot: projectRoot,
		seenPath:    filepath.Join(memDir, seenFileName),
		seenCache:   make(map[string]seenEntry),
	}
}

// LoadSeen reads the .seen file into memory.
func (cb *ContextBuilder) LoadSeen() error {
	f, err := os.Open(cb.seenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no seen file yet
		}
		return fmt.Errorf("immediate: open .seen: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) != 3 {
			continue
		}
		path := parts[0]
		hash := parts[1]
		ts, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			continue
		}
		cb.seenCache[path] = seenEntry{Hash: hash, Timestamp: ts}
	}
	return scanner.Err()
}

// SaveSeen persists the in-memory seen cache to disk.
func (cb *ContextBuilder) SaveSeen() error {
	if !cb.seenDirty {
		return nil
	}
	tmp := cb.seenPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("immediate: create .seen.tmp: %w", err)
	}
	for path, entry := range cb.seenCache {
		if _, err := fmt.Fprintf(f, "%s %s %d\n", path, entry.Hash, entry.Timestamp); err != nil {
			_ = f.Close()
			return fmt.Errorf("immediate: write .seen: %w", err)
		}
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return fmt.Errorf("immediate: sync .seen: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("immediate: close .seen: %w", err)
	}
	if err := os.Rename(tmp, cb.seenPath); err != nil {
		return fmt.Errorf("immediate: rename .seen: %w", err)
	}
	cb.seenDirty = false
	return nil
}

// isFileChanged compares the current hash against the last seen entry.
// It updates the seen cache with the new hash (marking dirty).
func (cb *ContextBuilder) isFileChanged(normPath string, currentHash string) bool {
	entry, ok := cb.seenCache[normPath]
	changed := !ok || entry.Hash != currentHash
	// Always update seen cache with the current observation.
	cb.seenCache[normPath] = seenEntry{Hash: currentHash, Timestamp: time.Now().Unix()}
	cb.seenDirty = true
	return changed
}

// BuildFromWorkingSet constructs an ImmediateContext from the given working set.
// It reads file contents, generates summaries and hashes, and applies budget.
func (cb *ContextBuilder) BuildFromWorkingSet(
	taskID string,
	intent string,
	contract *ContractSnapshot,
	wm *working.Engine,
	budget TokenBudget,
) (*ImmediateContext, error) {
	if err := cb.LoadSeen(); err != nil {
		return nil, err
	}

	ctx := &ImmediateContext{
		TaskID:     taskID,
		Intent:     intent,
		Contract:   contract,
		WorkingSet: &WorkingSetSnapshot{Bundles: make([]RetainedBundleSummary, 0)},
		Budget:     budget,
	}

	items, err := wm.List(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("immediate: list working set: %w", err)
	}

	for _, item := range items {
		bundle, err := wm.GetBundle(context.TODO(), item.BundleID)
		if err != nil {
			continue // skip unreadable bundles
		}

		summary := RetainedBundleSummary{
			BundleID:    bundle.BundleID,
			Type:        "bundle",
			Source:      "",
			Summary:     bundle.Goal,
			ContentHash: "",
			FileChanged: false,
			PacketCount: len(bundle.Packets),
		}

		// Budget check for bundle metadata.
		metaTokens := EstimateTokens(summary.Summary)
		if err := ctx.Budget.Consume(metaTokens); err != nil {
			// Budget exhausted: strip summaries, keep paths only.
			summary.Summary = ""
			ctx.WorkingSet.Bundles = append(ctx.WorkingSet.Bundles, summary)
			continue
		}

		// Process each packet (file reference) inside the bundle.
		for _, pkt := range bundle.Packets {
			if pkt.Source == "" {
				continue
			}
			normPath := filepath.ToSlash(pkt.Source)
			absPath := filepath.Join(cb.projectRoot, normPath)
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				// Fallback: try as-is.
				absPath = pkt.Source
			}

			content, err := os.ReadFile(absPath)
			if err != nil {
				// File unreadable: use path-only mode.
				fileSum := RetainedBundleSummary{
					BundleID:    bundle.BundleID,
					Type:        pkt.Type,
					Source:      normPath,
					Summary:     "",
					ContentHash: "",
					FileChanged: false,
					PacketCount: 0,
				}
				ctx.WorkingSet.Bundles = append(ctx.WorkingSet.Bundles, fileSum)
				continue
			}

			hash := ContentHash(content)
			changed := cb.isFileChanged(normPath, hash)
			summaryText := TruncateSummary(string(content))

			fileTokens := EstimateTokens(summaryText) + len(normPath)/4 + 8 // hash ~8 tokens
			if err := ctx.Budget.Consume(fileTokens); err != nil {
				// Budget exhausted: degrade to path-only mode.
				fileSum := RetainedBundleSummary{
					BundleID:    bundle.BundleID,
					Type:        pkt.Type,
					Source:      normPath,
					Summary:     "",
					ContentHash: hash,
					FileChanged: changed,
					PacketCount: 0,
				}
				ctx.WorkingSet.Bundles = append(ctx.WorkingSet.Bundles, fileSum)
				continue
			}

			fileSum := RetainedBundleSummary{
				BundleID:    bundle.BundleID,
				Type:        pkt.Type,
				Source:      normPath,
				Summary:     summaryText,
				ContentHash: hash,
				FileChanged: changed,
				PacketCount: 1,
			}
			ctx.WorkingSet.Bundles = append(ctx.WorkingSet.Bundles, fileSum)
		}
	}

	_ = cb.SaveSeen() // best-effort persistence
	return ctx, nil
}
