package gate

import (
	"strings"

	"github.com/axis-cli/axis/internal/types"
)

// denylistKeys are matched case-insensitively as substrings of context keys.
var denylistKeys = []string{
	"api_key", "apikey", "api-key",
	"token",
	"secret",
	"password",
	"authorization",
	"private_key",
}

// ScrubMap returns a copy of m with every entry whose key matches the
// denylist removed. Values are never inspected or logged (CLAUDE.md §1.7).
func ScrubMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		if isSensitiveKey(k) {
			continue
		}
		out[k] = v
	}
	return out
}

// ScrubMeta returns a copy of m with sensitive keys removed.
func ScrubMeta(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		if isSensitiveKey(k) {
			continue
		}
		out[k] = v
	}
	return out
}

func isSensitiveKey(key string) bool {
	lowered := strings.ToLower(key)
	for _, bad := range denylistKeys {
		if strings.Contains(lowered, bad) {
			return true
		}
	}
	return false
}

// ScrubRequest sanitizes the user-controlled payload fields of a request.
func ScrubRequest(req *types.HumanCallRequest) {
	if req == nil {
		return
	}
	req.Context = ScrubMap(req.Context)
	req.Input = ScrubMap(req.Input)
	req.Metadata = ScrubMeta(req.Metadata)
}
