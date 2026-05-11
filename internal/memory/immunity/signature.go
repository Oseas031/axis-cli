package immunity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// signatureHashLen is the number of hex characters in a canonical
// signature hash (first 128 bits of SHA-256).
const signatureHashLen = 32

// BuildSignature constructs a canonical Signature. It does NOT scrub
// sensitive args — call NormalizeArgs first if the raw map may contain
// secrets.
//
// The returned Signature has its NormalizedArgs map present (never nil),
// its ContractToolAllow slice sorted + deduplicated, and empty argument
// values dropped.
func BuildSignature(intentKind string, args map[string]string, toolAllow []string, errClass FailureClass) Signature {
	return Signature{
		IntentKind:        intentKind,
		NormalizedArgs:    canonicalArgs(args),
		ContractToolAllow: canonicalTools(toolAllow),
		ErrorClass:        errClass,
	}
}

// Hash returns the canonical hash of s: first 128 bits of SHA-256 over a
// deterministic JSON encoding (sorted map keys, sorted slices, empty
// values dropped). Output is exactly signatureHashLen (32) hex chars.
func (s Signature) Hash() string {
	canon := struct {
		IntentKind        string            `json:"intent_kind"`
		NormalizedArgs    map[string]string `json:"normalized_args"`
		ContractToolAllow []string          `json:"contract_tool_allow"`
		ErrorClass        FailureClass      `json:"error_class"`
	}{
		IntentKind:        s.IntentKind,
		NormalizedArgs:    canonicalArgs(s.NormalizedArgs),
		ContractToolAllow: canonicalTools(s.ContractToolAllow),
		ErrorClass:        s.ErrorClass,
	}
	// json.Marshal sorts map keys deterministically as of Go 1.12+;
	// our slice ordering is enforced by canonicalTools.
	b, err := json.Marshal(canon)
	if err != nil {
		// Marshal of a map[string]string + []string + strings cannot
		// fail in practice. Defensive only.
		return strings.Repeat("0", signatureHashLen)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:signatureHashLen/2])
}

// NormalizeArgs converts raw argument values to a deterministic
// map[string]string suitable for signature construction. Keys matching
// IsSensitiveKey are DROPPED ENTIRELY (not redacted) so they cannot
// influence the hash. Empty string values are dropped. Numeric and
// boolean values are stringified via fmt.Sprintf. Nested maps/slices
// are encoded via json.Marshal so the shape is captured deterministically.
func NormalizeArgs(raw map[string]any) map[string]string {
	if len(raw) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(raw))
	for k, v := range raw {
		if IsSensitiveKey(k) {
			continue
		}
		s := stringifyValue(v)
		if s == "" {
			continue
		}
		out[k] = s
	}
	return out
}

// canonicalArgs returns a non-nil map with empty values dropped. It does
// NOT touch keys (callers should pre-normalize via NormalizeArgs).
func canonicalArgs(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		if v == "" {
			continue
		}
		out[k] = v
	}
	return out
}

// canonicalTools returns the input sorted and deduplicated, empty entries
// removed. A nil/empty input yields a non-nil empty slice for stable JSON.
func canonicalTools(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, t := range in {
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// stringifyValue produces a deterministic string representation of v.
// For maps/slices we use json.Marshal so element order matters for the
// hash but does not surprise callers (they should normalize at call site).
func stringifyValue(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprintf("%v", x)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	}
}
