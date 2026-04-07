// Package rsync implements bidirectional file synchronisation using the rsync
// command-line tool over SSH. It provides both one-shot sync operations and a
// long-running daemon that periodically keeps local and remote paths in sync.
package rsync

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
	"github.com/BenjaminBenetti/Teeleport/internal/domainmodel"
)

// SyncEntry performs a single bidirectional sync cycle for one mount entry.
// Directory entries use a 9-step flow that detects local deletions via a
// manifest, propagates them through per-client delete logs, then pushes and
// pulls files. Single-file entries use a simpler push/pull without deletion
// tracking (config files are never considered deleted).
func SyncEntry(ssh domainmodel.SSHConfig, entry domainmodel.MountEntry) error {
	target := config.ExpandPath(entry.Target)

	isDir := false
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		isDir = true
	}

	if isDir {
		return syncDirectoryEntry(ssh, entry, target)
	}
	return syncFileEntry(ssh, entry, target)
}

// syncFileEntry handles single-file mount entries with a simple push/pull.
// Delete-log tracking is not applicable to file mounts.
func syncFileEntry(ssh domainmodel.SSHConfig, entry domainmodel.MountEntry, target string) error {
	if err := runRsync(ssh, target, entry.Source, false, "push", true); err != nil {
		return fmt.Errorf("push %s: %w", entry.Name, err)
	}
	if err := runRsync(ssh, entry.Source, target, false, "pull", true); err != nil {
		return fmt.Errorf("pull %s: %w", entry.Name, err)
	}
	return nil
}

// syncDirectoryEntry implements the full 9-step sync cycle for directory
// mount entries, including manifest-based deletion detection, per-client
// delete-log coordination, and bidirectional file synchronisation.
func syncDirectoryEntry(ssh domainmodel.SSHConfig, entry domainmodel.MountEntry, target string) error {
	// Step 1: Detect local deletions via manifest diff
	previous, _ := LoadManifest(entry.Name)
	current, _ := ScanLocalFiles(target, true)

	var newDeletions []string
	if len(previous) > 0 {
		newDeletions = DetectDeletions(previous, current)
	}

	// Step 2: Record new deletions in our per-client delete log
	if len(newDeletions) > 0 {
		if err := AppendDeletions(target, newDeletions); err != nil {
			fmt.Printf("[teeleport] rsync: append deletions warning for %s: %v\n", entry.Name, err)
		}
		fmt.Printf("[teeleport] rsync: detected %d new deletion(s) for %s\n", len(newDeletions), entry.Name)
	}

	// Step 3: SSH rm newly detected deletions from server (immediate cleanup)
	if len(newDeletions) > 0 {
		if err := deleteRemoteFiles(ssh, entry.Source, newDeletions, true); err != nil {
			fmt.Printf("[teeleport] rsync: SSH rm warning for %s: %v\n", entry.Name, err)
		}
	}

	// Step 4: Push our delete log to server
	if err := pushOurDeleteLog(ssh, target, entry.Source); err != nil {
		fmt.Printf("[teeleport] rsync: push delete log warning for %s: %v\n", entry.Name, err)
	}

	// Step 5: Pull all per-client delete logs from server
	if err := pullAllDeleteLogs(ssh, target, entry.Source); err != nil {
		fmt.Printf("[teeleport] rsync: pull delete logs warning for %s: %v\n", entry.Name, err)
	}

	// Step 6: Process all delete logs — delete local files + SSH rm from server
	logDeleted, err := processDeleteLogEntries(target)
	if err != nil {
		fmt.Printf("[teeleport] rsync: process delete logs warning for %s: %v\n", entry.Name, err)
	}
	if len(logDeleted) > 0 {
		fmt.Printf("[teeleport] rsync: delete log removed %d file(s) for %s\n", len(logDeleted), entry.Name)
		if err := deleteRemoteFiles(ssh, entry.Source, logDeleted, true); err != nil {
			fmt.Printf("[teeleport] rsync: SSH rm (log) warning for %s: %v\n", entry.Name, err)
		}
	}

	// Step 7: Push all files (no --delete, --update)
	if err := runRsync(ssh, target, entry.Source, true, "push", true); err != nil {
		return fmt.Errorf("push %s: %w", entry.Name, err)
	}

	// Step 8: Pull all files (--delete, --update)
	if err := runRsync(ssh, entry.Source, target, true, "pull", true); err != nil {
		return fmt.Errorf("pull %s: %w", entry.Name, err)
	}

	// Step 9: Save manifest reflecting post-sync state
	postSyncFiles, err := ScanLocalFiles(target, true)
	if err != nil {
		fmt.Printf("[teeleport] rsync: manifest scan warning for %s: %v\n", entry.Name, err)
	} else {
		if err := SaveManifest(entry.Name, postSyncFiles); err != nil {
			fmt.Printf("[teeleport] rsync: manifest save warning for %s: %v\n", entry.Name, err)
		}
	}

	return nil
}

// deleteRemoteFiles removes the listed relative paths from the remote host
// via a single SSH session. For directory entries, paths are relative to
// remotePath. For file entries, remotePath itself is removed.
func deleteRemoteFiles(ssh domainmodel.SSHConfig, remotePath string, files []string, isDir bool) error {
	var rmArgs []string
	for _, f := range files {
		if isDir {
			rmArgs = append(rmArgs, filepath.Join(remotePath, f))
		} else {
			rmArgs = append(rmArgs, remotePath)
		}
	}

	sshUser := ssh.User
	if sshUser == "" {
		u, err := user.Current()
		if err != nil {
			sshUser = "root"
		} else {
			sshUser = u.Username
		}
	}

	sshArgs := []string{
		"-p", fmt.Sprintf("%d", ssh.Port),
		"-o", "StrictHostKeyChecking=no",
		"-o", "ConnectTimeout=10",
		"-o", "BatchMode=yes",
	}
	if ssh.IdentityFile != "" {
		sshArgs = append(sshArgs, "-i", config.ExpandPath(ssh.IdentityFile))
	}
	sshArgs = append(sshArgs, fmt.Sprintf("%s@%s", sshUser, ssh.Host))

	// Build a single rm -f command for all files
	var rmCmd strings.Builder
	rmCmd.WriteString("rm -f")
	for _, p := range rmArgs {
		fmt.Fprintf(&rmCmd, " %q", p)
	}
	sshArgs = append(sshArgs, rmCmd.String())

	cmd := exec.Command("ssh", sshArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ssh rm failed: %w\noutput: %s", err, strings.TrimSpace(string(output)))
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
