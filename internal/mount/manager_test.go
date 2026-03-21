package mount

import (
	"strings"
	"testing"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
)

func TestIsFileMount(t *testing.T) {
	tests := []struct {
		entryType string
		want      bool
	}{
		{"file", true},
		{"directory", false},
		{"", false},
	}
	for _, tt := range tests {
		entry := config.MountEntry{Type: tt.entryType}
		if got := isFileMount(entry); got != tt.want {
			t.Errorf("isFileMount(Type=%q) = %v, want %v", tt.entryType, got, tt.want)
		}
	}
}

func TestRemoteParent(t *testing.T) {
	tests := []struct {
		source string
		want   string
	}{
		{"/home/user/.claude.json", "/home/user"},
		{"/home/user/.config/some/file.txt", "/home/user/.config/some"},
		{"/file.txt", "/"},
	}
	for _, tt := range tests {
		if got := remoteParent(tt.source); got != tt.want {
			t.Errorf("remoteParent(%q) = %q, want %q", tt.source, got, tt.want)
		}
	}
}

func TestRemoteBasename(t *testing.T) {
	tests := []struct {
		source string
		want   string
	}{
		{"/home/user/.claude.json", ".claude.json"},
		{"/home/user/.config/some/file.txt", "file.txt"},
		{"/file.txt", "file.txt"},
	}
	for _, tt := range tests {
		if got := remoteBasename(tt.source); got != tt.want {
			t.Errorf("remoteBasename(%q) = %q, want %q", tt.source, got, tt.want)
		}
	}
}

func TestStagingDir(t *testing.T) {
	got := stagingDir("claude-json")
	if got == "" {
		t.Error("stagingDir returned empty string")
	}
	// Should contain .teeleport/mounts/claude-json
	if !strings.Contains(got, ".teeleport/mounts/claude-json") {
		t.Errorf("stagingDir(%q) = %q, expected to contain .teeleport/mounts/claude-json", "claude-json", got)
	}
}
