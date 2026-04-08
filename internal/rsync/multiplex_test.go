package rsync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BenjaminBenetti/Teeleport/internal/domainmodel"
)

func TestBuildMultiplexBlock(t *testing.T) {
	got := buildMultiplexBlock("ssh.example.com")
	want := "Host ssh.example.com\n" +
		"    ControlMaster auto\n" +
		"    ControlPath ~/.ssh/cm-teeleport-%r@%h:%p\n" +
		"    ControlPersist 10m\n"
	if got != want {
		t.Errorf("buildMultiplexBlock() =\n%s\nwant:\n%s", got, want)
	}
}

func TestEnsureSSHMultiplexing_EmptyHost(t *testing.T) {
	ssh := domainmodel.SSHConfig{Host: ""}
	if err := EnsureSSHMultiplexing(ssh); err != nil {
		t.Errorf("EnsureSSHMultiplexing() with empty host returned error: %v", err)
	}
}

func TestEnsureSSHMultiplexing_CreatesConfig(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	configPath := filepath.Join(sshDir, "config")

	// Patch the expand function by calling the internal helpers directly.
	if err := ensureSSHDirectory(sshDir); err != nil {
		t.Fatalf("ensureSSHDirectory() error: %v", err)
	}
	if err := ensureSSHConfigFile(configPath); err != nil {
		t.Fatalf("ensureSSHConfigFile() error: %v", err)
	}

	// Verify directory permissions.
	info, err := os.Stat(sshDir)
	if err != nil {
		t.Fatalf("stat sshDir: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		t.Errorf("sshDir permissions = %o, want 700", perm)
	}

	// Verify file permissions.
	info, err = os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("config permissions = %o, want 600", perm)
	}
}

func TestEnsureSSHDirectory_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Should succeed without error.
	if err := ensureSSHDirectory(sshDir); err != nil {
		t.Errorf("ensureSSHDirectory() on existing dir: %v", err)
	}
}

func TestEnsureSSHConfigFile_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	existing := filepath.Join(tmpDir, "config")
	if err := os.WriteFile(existing, []byte("Host *\n    ServerAliveInterval 60\n"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := ensureSSHConfigFile(existing); err != nil {
		t.Errorf("ensureSSHConfigFile() on existing file: %v", err)
	}

	// Content must not be modified.
	data, _ := os.ReadFile(existing)
	if !strings.Contains(string(data), "ServerAliveInterval") {
		t.Error("ensureSSHConfigFile() clobbered existing content")
	}
}

func TestEnsureSSHDirectory_NotADirectory(t *testing.T) {
	tmpDir := t.TempDir()
	fakePath := filepath.Join(tmpDir, ".ssh")

	// Create a file where the directory should be.
	if err := os.WriteFile(fakePath, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := ensureSSHDirectory(fakePath)
	if err == nil {
		t.Error("ensureSSHDirectory() should fail when path is a file")
	}
}
