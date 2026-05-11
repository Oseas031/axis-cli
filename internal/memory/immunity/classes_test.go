package immunity

import "testing"

func TestIsKnownClass(t *testing.T) {
	tests := []struct {
		c    FailureClass
		want bool
	}{
		// Accepted: every known prefix with a non-empty reason segment.
		{"failure.provider.timeout", true},
		{"failure.provider.unreachable", true},
		{"failure.tool.permission_denied", true},
		{"failure.contract.unsatisfied", true},
		{"failure.intent.ambiguous", true},
		{"failure.runtime.panic", true},

		// Rejected: bare prefix with no reason.
		{"failure.provider.", false},
		{"failure.runtime.", false},

		// Rejected: dropped or never-defined namespaces.
		{"failure.safego.recovered", false}, // dropped during reconciliation
		{"failure.bogus.unknown", false},
		{"failure.network.timeout", false},

		// Rejected: malformed.
		{"", false},
		{"failure", false},
		{"failure.provider", false},
		{"provider.timeout", false},
		{"FAILURE.PROVIDER.TIMEOUT", false}, // case-sensitive
	}
	for _, tt := range tests {
		t.Run(string(tt.c), func(t *testing.T) {
			if got := IsKnownClass(tt.c); got != tt.want {
				t.Errorf("IsKnownClass(%q) = %v, want %v", tt.c, got, tt.want)
			}
		})
	}
}
