package immunity

import "context"

// Recaller queries Immunity records by signature shape. P0 uses
// field-equality matching only (no vector or embedding similarity, per
// requirements FR4).
type Recaller struct {
	store *Store
}

// NewRecaller wraps a Store for query-only access.
func NewRecaller(s *Store) *Recaller {
	return &Recaller{store: s}
}

// Recall returns Immunity records whose SignatureHash matches sig.Hash().
// Deprecated records are excluded. limit == 0 means unbounded.
func (r *Recaller) Recall(ctx context.Context, sig Signature, limit int) ([]ImmunityRecord, error) {
	target := sig.Hash()
	all, err := r.store.List(ctx, ListFilter{})
	if err != nil {
		return nil, err
	}
	out := make([]ImmunityRecord, 0, limit)
	for _, rec := range all {
		if rec.SignatureHash != target {
			continue
		}
		out = append(out, rec)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

// RecallSimilar returns records that match a partial signature shape.
// Zero-value fields in partial mean "any value matches". For
// ContractToolAllow, the record's tools MUST be a superset of partial's
// (every tool in partial appears in the record). Deprecated records are
// excluded. limit == 0 means unbounded.
func (r *Recaller) RecallSimilar(ctx context.Context, partial PartialSignature, limit int) ([]ImmunityRecord, error) {
	all, err := r.store.List(ctx, ListFilter{})
	if err != nil {
		return nil, err
	}
	out := make([]ImmunityRecord, 0, limit)
	for _, rec := range all {
		if !matchesPartial(partial, rec.Signature) {
			continue
		}
		out = append(out, rec)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

// matchesPartial reports whether sig satisfies all non-zero fields of p.
func matchesPartial(p PartialSignature, sig Signature) bool {
	if p.IntentKind != "" && sig.IntentKind != p.IntentKind {
		return false
	}
	if len(p.ContractToolAllow) > 0 {
		recTools := make(map[string]struct{}, len(sig.ContractToolAllow))
		for _, t := range sig.ContractToolAllow {
			recTools[t] = struct{}{}
		}
		for _, t := range p.ContractToolAllow {
			if _, ok := recTools[t]; !ok {
				return false
			}
		}
	}
	return true
}
