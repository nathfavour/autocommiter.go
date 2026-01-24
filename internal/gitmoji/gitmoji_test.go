package gitmoji

import (
	"testing"
)

func TestFindBestGitmoji(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"fix a bug in auth", "🐛"},
		{"add new feature for logging", "✨"},
		{"improve performance of diff", "⚡"},
		{"update documentation for api", "📝"},
		{"remove unused files", "🔥"},
		{"unknown change", ""}, // Should return random or nil handled by caller
	}

	for _, tt := range tests {
		got := FindBestGitmoji(tt.message)
		if tt.expected != "" {
			if got == nil || got.Emoji != tt.expected {
				t.Errorf("FindBestGitmoji(%q) = %v; want %v", tt.message, got, tt.expected)
			}
		}
	}
}

func TestGetGitmojifiedMessage(t *testing.T) {
	msg := "fix a bug"
	got := GetGitmojifiedMessage(msg)
	if !testing.Short() {
		if got == "" || got == msg {
			t.Errorf("GetGitmojifiedMessage(%q) = %q; want it to be modified", msg, got)
		}
	}
}
