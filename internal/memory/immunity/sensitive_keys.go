package immunity

import "strings"

// sensitivePatterns is the P0 local sensitive-key list referenced from the
// design (D1 reconciliation). Matching is case-insensitive substring against
// the argument map key. The list lives here, NOT in internal/safego/, because
// safego is panic-recovery; no centralised redaction utility exists in the
// repo today. Per CLAUDE.md §12 metadata-promotion rule, the list moves to
// internal/types/ only when a second package needs it.
var sensitivePatterns = []string{
	"api_key",
	"apikey",
	"token",
	"bearer",
	"password",
	"passwd",
	"secret",
	"credential",
	"auth",
	"private_key",
}

// IsSensitiveKey reports whether key matches any sensitive pattern. Match is
// case-insensitive substring; an empty key is not sensitive.
func IsSensitiveKey(key string) bool {
	if key == "" {
		return false
	}
	lower := strings.ToLower(key)
	for _, p := range sensitivePatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}
