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

func TestRemoteEnsureCmd_Directory(t *testing.T) {
	got := remoteEnsureCmd("/home/user/.claude", false)
	want := `mkdir -p "/home/user/.claude"`
	if got != want {
		t.Errorf("remoteEnsureCmd(dir) = %q, want %q", got, want)
	}
}

func TestRemoteEnsureCmd_File(t *testing.T) {
	got := remoteEnsureCmd("/home/user/.claude.json", true)
	want := `mkdir -p "/home/user" && touch "/home/user/.claude.json"`
	if got != want {
		t.Errorf("remoteEnsureCmd(file) = %q, want %q", got, want)
	}
}

func TestRemoteEnsureCmd_NestedFile(t *testing.T) {
	got := remoteEnsureCmd("/home/user/.config/app/settings.json", true)
	want := `mkdir -p "/home/user/.config/app" && touch "/home/user/.config/app/settings.json"`
	if got != want {
		t.Errorf("remoteEnsureCmd(nested file) = %q, want %q", got, want)
	}
}

func TestRemoteEnsureCmd_RootFile(t *testing.T) {
	got := remoteEnsureCmd("/file.txt", true)
	want := `mkdir -p "/" && touch "/file.txt"`
	if got != want {
		t.Errorf("remoteEnsureCmd(root file) = %q, want %q", got, want)
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
