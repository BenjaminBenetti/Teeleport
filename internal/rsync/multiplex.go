package rsync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
	filecopy "github.com/BenjaminBenetti/Teeleport/internal/copy"
	"github.com/BenjaminBenetti/Teeleport/internal/domainmodel"
)

// ===========================================
// Constants
// ===========================================

const (
	// sshMultiplexSentinel is the sentinel name used by ApplyAppend to
	// identify the Teeleport-managed block inside ~/.ssh/config.
	sshMultiplexSentinel = "ssh-multiplex"

	// controlPathPrefix distinguishes Teeleport-managed control sockets
	// from any user-configured multiplexing.
	controlPathPrefix = "~/.ssh/cm-teeleport-%r@%h:%p"

	// controlPersistDuration keeps the master connection alive after the
	// last multiplexed session disconnects. 10 minutes comfortably covers
	// the default 30-second sync interval.
	controlPersistDuration = "10m"
)

// ===========================================
// Public API
// ===========================================

// EnsureSSHMultiplexing writes a ControlMaster configuration block for the
// given SSH host into ~/.ssh/config. The block is wrapped in sentinel
// markers so that repeated calls replace rather than duplicate the entry.
//
// SSH resolves options with a first-match-wins strategy, so if the user
// already has a Host block for the same host above the sentinel block,
// their settings take precedence.
func EnsureSSHMultiplexing(ssh domainmodel.SSHConfig) error {
	if ssh.Host == "" {
		return nil
	}

	sshDir := config.ExpandPath("~/.ssh")
	sshConfig := filepath.Join(sshDir, "config")

	if err := ensureSSHDirectory(sshDir); err != nil {
		return fmt.Errorf("ensuring ~/.ssh directory: %w", err)
	}

	if err := ensureSSHConfigFile(sshConfig); err != nil {
		return fmt.Errorf("ensuring ~/.ssh/config file: %w", err)
	}

	block := buildMultiplexBlock(ssh.Host)

	if err := filecopy.ApplyAppend(sshMultiplexSentinel, block, sshConfig); err != nil {
		return fmt.Errorf("writing ssh multiplex config: %w", err)
	}

	fmt.Printf("[teeleport] ssh multiplexing configured for %s\n", ssh.Host)
	return nil
}

// ===========================================
// Config generation
// ===========================================

// buildMultiplexBlock returns the SSH config text placed between sentinel
// markers. It configures ControlMaster multiplexing for the given host.
func buildMultiplexBlock(host string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Host %s\n", host)
	fmt.Fprintf(&b, "    ControlMaster auto\n")
	fmt.Fprintf(&b, "    ControlPath %s\n", controlPathPrefix)
	fmt.Fprintf(&b, "    ControlPersist %s\n", controlPersistDuration)
	return b.String()
}

// ===========================================
// Filesystem helpers
// ===========================================

// ensureSSHDirectory creates ~/.ssh with the correct 0700 permissions if
// it does not already exist.
func ensureSSHDirectory(sshDir string) error {
	info, err := os.Stat(sshDir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s exists but is not a directory", sshDir)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	return os.MkdirAll(sshDir, 0o700)
}

// ensureSSHConfigFile creates an empty ~/.ssh/config with 0600 permissions
// if the file does not already exist. Existing files are left untouched.
func ensureSSHConfigFile(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, nil, 0o600)
}
