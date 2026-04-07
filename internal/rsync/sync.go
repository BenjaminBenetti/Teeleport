// Package rsync implements bidirectional file synchronisation using the rsync
// command-line tool over SSH. It provides both one-shot sync operations and a
// long-running daemon that periodically keeps local and remote paths in sync.
package rsync

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
	"github.com/BenjaminBenetti/Teeleport/internal/domainmodel"
)

// SyncEntry performs a single bidirectional sync cycle for one mount entry.
// It pushes local changes to the remote (skipping files that are newer on the
// remote), then pulls remote changes to the local side (skipping files that
// are newer locally). This achieves last-write-wins semantics.
func SyncEntry(ssh domainmodel.SSHConfig, entry domainmodel.MountEntry) error {
	target := config.ExpandPath(entry.Target)

	isDir := false
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		isDir = true
	}

	// Step 1: Push local → remote (only overwrite if local is newer)
	if err := runRsync(ssh, target, entry.Source, isDir, "push", true); err != nil {
		return fmt.Errorf("push %s: %w", entry.Name, err)
	}

	// Step 2: Pull remote → local (only overwrite if remote is newer)
	if err := runRsync(ssh, entry.Source, target, isDir, "pull", true); err != nil {
		return fmt.Errorf("pull %s: %w", entry.Name, err)
	}

	return nil
}

// PullEntry performs a one-way pull from remote to local for initial sync.
// It skips --update so the remote always overwrites local, regardless of
// timestamps. This is correct for first-boot: the remote holds the
// persisted state and freshly-created local files should not win.
func PullEntry(ssh domainmodel.SSHConfig, source, target string, isDir bool) error {
	return runRsync(ssh, source, target, isDir, "pull", false)
}

// runRsync executes a single rsync transfer in the specified direction.
// For "push", localPath is the source and remotePath is the destination.
// For "pull", remotePath is the source and localPath is the destination.
// The --delete flag is only applied to pull operations so that the local side
// mirrors the server, which accumulates files from all clients. Push never
// deletes from the server, preventing multi-client sync fighting.
// When useUpdate is true, --update is included so files newer on the
// destination are never overwritten (last-write-wins). When false, the
// source always overwrites the destination (used for initial first-boot sync).
func runRsync(ssh domainmodel.SSHConfig, fromPath, toPath string, isDir bool, direction string, useUpdate bool) error {
	sshCmd := buildSSHCommand(ssh)

	args := []string{
		"-az", // archive mode + compression
		"-e", sshCmd,
	}
	if direction == "pull" {
		args = append(args, "--delete") // only pull mirrors server state; push never deletes from server
	}
	if useUpdate {
		args = append(args, "--update") // skip files newer on the receiver
	}

	var src, dst string
	switch direction {
	case "push":
		src = fromPath
		dst = remoteAddress(ssh, toPath)
	case "pull":
		src = remoteAddress(ssh, fromPath)
		dst = toPath
	default:
		return fmt.Errorf("unknown direction %q", direction)
	}

	// For directory syncs, trailing slash means "sync contents of"
	if isDir {
		if !strings.HasSuffix(src, "/") {
			src += "/"
		}
		if !strings.HasSuffix(dst, "/") {
			dst += "/"
		}
	}

	args = append(args, src, dst)

	cmd := exec.Command("rsync", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rsync %s failed: %w\noutput: %s", direction, err, strings.TrimSpace(string(output)))
	}
	return nil
}

// buildSSHCommand constructs the SSH command string passed to rsync's -e flag.
func buildSSHCommand(ssh domainmodel.SSHConfig) string {
	parts := []string{
		"ssh",
		fmt.Sprintf("-p %d", ssh.Port),
		"-o StrictHostKeyChecking=no",
		"-o ConnectTimeout=10",
		"-o BatchMode=yes",
	}
	if ssh.IdentityFile != "" {
		parts = append(parts, fmt.Sprintf("-i \"%s\"", config.ExpandPath(ssh.IdentityFile)))
	}
	return strings.Join(parts, " ")
}

// remoteAddress formats a remote path for rsync: user@host:path.
func remoteAddress(ssh domainmodel.SSHConfig, path string) string {
	sshUser := ssh.User
	if sshUser == "" {
		u, err := user.Current()
		if err != nil {
			sshUser = "root"
		} else {
			sshUser = u.Username
		}
	}
	return fmt.Sprintf("%s@%s:%s", sshUser, ssh.Host, path)
}

// HasRsyncEntries returns true if any mount entry uses the rsync backend.
func HasRsyncEntries(entries []domainmodel.MountEntry) bool {
	for _, e := range entries {
		if e.Backend == "rsync" {
			return true
		}
	}
	return false
}

// FilterRsyncEntries returns only mount entries that use the rsync backend.
func FilterRsyncEntries(entries []domainmodel.MountEntry) []domainmodel.MountEntry {
	var result []domainmodel.MountEntry
	for _, e := range entries {
		if e.Backend == "rsync" {
			result = append(result, e)
		}
	}
	return result
}
