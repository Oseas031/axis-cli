package contextpack

import (
	"fmt"
	"sync"
)

type ReadinessRecord struct {
	Artifact ReadinessArtifact `json:"artifact"`
	Bundle   ContextBundle     `json:"bundle"`
}

type ReadinessRegistry struct {
	mu      sync.RWMutex
	records map[string]ReadinessRecord
	store   ReadinessStore
}

func NewReadinessRegistry() *ReadinessRegistry {
	return &ReadinessRegistry{records: make(map[string]ReadinessRecord)}
}

// NewReadinessRegistryWithStore creates a registry backed by a persistent store.
// It loads existing records from the store on creation.
func NewReadinessRegistryWithStore(store ReadinessStore) (*ReadinessRegistry, error) {
	records, err := store.LoadAll()
	if err != nil {
		return nil, err
	}
	return &ReadinessRegistry{records: records, store: store}, nil
}

func (r *ReadinessRegistry) Register(bundle *ContextBundle) (ReadinessArtifact, error) {
	artifact, err := NewReadinessArtifact(bundle)
	if err != nil {
		return ReadinessArtifact{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records[artifact.BundleID] = ReadinessRecord{Artifact: artifact, Bundle: cloneBundle(bundle)}
	if r.store != nil {
		if err := r.store.SaveAll(r.records); err != nil {
			// Persistence failure must not block the in-memory registration,
			// but the caller can observe it via logs in future observability work.
		}
	}
	return artifact, nil
}

func (r *ReadinessRegistry) Inspect(bundleID string) (ReadinessRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.store != nil {
		if err := r.syncFromStore(); err != nil {
			// Continue with in-memory data even if store sync fails.
		}
	}
	record, ok := r.records[bundleID]
	if !ok {
		return ReadinessRecord{}, fmt.Errorf("context readiness record %s not found", bundleID)
	}
	return cloneReadinessRecord(record), nil
}

// syncFromStore reloads records from the persistent store into memory.
// Caller must hold r.mu.
func (r *ReadinessRegistry) syncFromStore() error {
	if r.store == nil {
		return nil
	}
	records, err := r.store.LoadAll()
	if err != nil {
		return err
	}
	r.records = records
	return nil
}

func (r *ReadinessRegistry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = make(map[string]ReadinessRecord)
	if r.store != nil {
		_ = r.store.DeleteAll()
	}
}

var DefaultRegistry = NewReadinessRegistry()

func cloneReadinessRecord(record ReadinessRecord) ReadinessRecord {
	clone := record
	clone.Bundle = cloneBundle(&record.Bundle)
	return clone
}

func cloneBundle(bundle *ContextBundle) ContextBundle {
	if bundle == nil {
		return ContextBundle{}
	}
	clone := *bundle
	if bundle.Packets != nil {
		clone.Packets = append([]ContextPacket(nil), bundle.Packets...)
	}
	clone.Trace.Selected = append([]TraceItem(nil), bundle.Trace.Selected...)
	clone.Trace.Excluded = append([]TraceItem(nil), bundle.Trace.Excluded...)
	clone.Trace.Notes = append([]string(nil), bundle.Trace.Notes...)
	return clone
}
