// Package immunity implements the Immunity Memory layer for Axis.
//
// Immunity Memory turns task failures into a queryable, signature-indexed
// "failure as asset" layer. It is a thin derived view over the existing
// long-term event log; promotion writes one new memory.immunity.promoted
// event into tasks.jsonl. Forgetting is soft-mark only.
//
// See docs/specs/immunity-memory/ for the full specification.
package immunity

import (
	"strings"
	"time"
)

// FailureClass is a namespaced failure category. Format:
// "failure.<subsystem>.<reason>". Use classes.go IsKnownClass to validate.
type FailureClass string

// Signature is the deterministic shape used for similarity matching.
// All slice fields MUST be sorted and deduplicated before hashing.
type Signature struct {
	IntentKind        string            `json:"intent_kind"`
	NormalizedArgs    map[string]string `json:"normalized_args"`
	ContractToolAllow []string          `json:"contract_tool_allow"`
	ErrorClass        FailureClass      `json:"error_class"`
}

// PartialSignature is the partial-match shape for RecallSimilar.
// Zero-value fields are treated as "any value matches" (per D3 in design).
type PartialSignature struct {
	IntentKind        string
	ContractToolAllow []string
}

// ImmunityRecord is the in-memory shape of a promoted failure.
type ImmunityRecord struct {
	ImmunityID      string       `json:"immunity_id"`
	SourceTaskID    string       `json:"source_task_id"`
	Signature       Signature    `json:"signature"`
	SignatureHash   string       `json:"signature_hash"`
	Cause           string       `json:"cause"`
	FailureClass    FailureClass `json:"failure_class"`
	PromotedBy      string       `json:"promoted_by"`
	PromotedAt      time.Time    `json:"promoted_at"`
	SourceDigest    string       `json:"source_digest,omitempty"`
	Deprecated      bool         `json:"deprecated,omitempty"`
	DeprecatedAt    *time.Time   `json:"deprecated_at,omitempty"`
	DeprecateReason string       `json:"deprecate_reason,omitempty"`
}

// PromoteInput is the validated input to Store.Promote.
type PromoteInput struct {
	SourceTaskID string
	Cause        string
	FailureClass FailureClass
	PromotedBy   string
}

// Validate enforces the preconditions required by tasks.md T2.2.
// Empty FailureClass is allowed at this stage (Promote may auto-derive
// it from the source task's terminal event before storage).
func (p PromoteInput) Validate() error {
	if strings.TrimSpace(p.SourceTaskID) == "" {
		return ErrSourceTaskIDRequired
	}
	if strings.TrimSpace(p.Cause) == "" {
		return ErrCauseRequired
	}
	if strings.TrimSpace(p.PromotedBy) == "" {
		return ErrPromotedByRequired
	}
	if p.FailureClass != "" && !IsKnownClass(p.FailureClass) {
		return ErrUnknownFailureClass
	}
	return nil
}

// ListFilter constrains a List() call.
type ListFilter struct {
	Class             FailureClass
	Since             *time.Time
	IncludeDeprecated bool
	Limit             int
}
