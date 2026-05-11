package immunity

import "strings"

// knownClassPrefixes is the closed set of accepted FailureClass prefixes.
// failure.safego.* was dropped during design reconciliation: internal/safego
// is panic-recovery, not auth/secret, so panic-type failures fold into
// failure.runtime.*. See docs/specs/immunity-memory/tasks.md T2.2.
var knownClassPrefixes = []string{
	"failure.provider.",
	"failure.tool.",
	"failure.contract.",
	"failure.intent.",
	"failure.runtime.",
}

// IsKnownClass reports whether c starts with one of the accepted prefixes.
// The prefix MUST be followed by a non-empty reason segment.
func IsKnownClass(c FailureClass) bool {
	s := string(c)
	for _, p := range knownClassPrefixes {
		if strings.HasPrefix(s, p) && len(s) > len(p) {
			return true
		}
	}
	return false
}
