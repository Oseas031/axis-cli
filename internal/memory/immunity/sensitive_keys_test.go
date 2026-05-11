package immunity

import "testing"

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		// Direct hits.
		{"api_key", true},
		{"apikey", true},
		{"token", true},
		{"bearer", true},
		{"password", true},
		{"passwd", true},
		{"secret", true},
		{"credential", true},
		{"auth", true},
		{"private_key", true},

		// Case-insensitive.
		{"API_KEY", true},
		{"Authorization", true},
		{"X-Auth-Token", true},

		// Substring inside larger key.
		{"openai_api_key", true},
		{"user_password_hash", true},
		{"my_secret_value", true},

		// Non-sensitive.
		{"user_id", false},
		{"intent_kind", false},
		{"tool_name", false},
		{"timeout_ms", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := IsSensitiveKey(tt.key); got != tt.want {
				t.Errorf("IsSensitiveKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}
