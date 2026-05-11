package immunity

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestImmunityRecord_JSONRoundTrip(t *testing.T) {
	original := ImmunityRecord{
		ImmunityID:    "imm-abc-123",
		SourceTaskID:  "task-xyz",
		Signature:     Signature{IntentKind: "build.binary"},
		SignatureHash: "deadbeef",
		Cause:         "504 timeout",
		FailureClass:  "failure.provider.timeout",
		PromotedBy:    "user:alex",
		PromotedAt:    time.Date(2026, 5, 12, 3, 0, 0, 0, time.UTC),
	}
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got ImmunityRecord
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ImmunityID != original.ImmunityID || got.PromotedAt != original.PromotedAt {
		t.Errorf("round-trip mismatch: got %+v want %+v", got, original)
	}
}

func TestImmunityRecord_DeprecatedOmittedWhenZero(t *testing.T) {
	r := ImmunityRecord{ImmunityID: "imm-1"}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(b)
	for _, field := range []string{"deprecated", "deprecated_at", "deprecate_reason", "source_digest"} {
		if contains(s, `"`+field+`"`) {
			t.Errorf("zero-value field %q should be omitted from JSON, got: %s", field, s)
		}
	}
}

func TestPromoteInput_Validate(t *testing.T) {
	tests := []struct {
		name string
		in   PromoteInput
		want error
	}{
		{
			name: "valid without class",
			in:   PromoteInput{SourceTaskID: "t1", Cause: "x", PromotedBy: "u"},
			want: nil,
		},
		{
			name: "valid with known class",
			in:   PromoteInput{SourceTaskID: "t1", Cause: "x", PromotedBy: "u", FailureClass: "failure.provider.timeout"},
			want: nil,
		},
		{
			name: "empty source task id",
			in:   PromoteInput{Cause: "x", PromotedBy: "u"},
			want: ErrSourceTaskIDRequired,
		},
		{
			name: "whitespace source task id",
			in:   PromoteInput{SourceTaskID: "   ", Cause: "x", PromotedBy: "u"},
			want: ErrSourceTaskIDRequired,
		},
		{
			name: "empty cause",
			in:   PromoteInput{SourceTaskID: "t1", PromotedBy: "u"},
			want: ErrCauseRequired,
		},
		{
			name: "empty promoted_by",
			in:   PromoteInput{SourceTaskID: "t1", Cause: "x"},
			want: ErrPromotedByRequired,
		},
		{
			name: "unknown class",
			in:   PromoteInput{SourceTaskID: "t1", Cause: "x", PromotedBy: "u", FailureClass: "failure.bogus.unknown"},
			want: ErrUnknownFailureClass,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.Validate()
			if !errors.Is(got, tt.want) {
				t.Errorf("Validate() error = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPartialSignature_ZeroValue(t *testing.T) {
	var p PartialSignature
	if p.IntentKind != "" || p.ContractToolAllow != nil {
		t.Errorf("zero PartialSignature has unexpected non-zero fields: %+v", p)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
