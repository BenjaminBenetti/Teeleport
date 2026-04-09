package rsync

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
	"github.com/BenjaminBenetti/Teeleport/internal/domainmodel"
)

// ===========================================
// Constants
// ===========================================

const (
	deleteLogPrefix = ".delete-log."
	deleteLogSuffix = ".manifest"

	// cleanupCycleInterval is the number of sync cycles between stale
	// delete-log cleanup runs. At 30 s per cycle this is roughly 3 hours.
	cleanupCycleInterval = 360

	// staleLogMaxAge is how long a per-client delete log may go unwritten
	// before it is considered stale and eligible for removal.
	staleLogMaxAge = 24 * time.Hour

	// deleteLogMaxAge is the maximum age of a delete log entry to be
	// processed. Entries older than this are ignored to prevent unbounded
	// accumulation of delete operations across sync cycles.
	deleteLogMaxAge = 30 * time.Minute
)

// ===========================================
// Types
// ===========================================

// DeleteLogEntry represents a single file deletion event recorded in a
// per-client delete log.
type DeleteLogEntry struct {
	Path      string
	DeletedAt time.Time
}

// ===========================================
// Naming helpers
// ===========================================

// deleteLogFilename returns the delete log filename for the current client,
// identified by hostname (e.g. ".delete-log.codespace-abc.manifest").
func deleteLogFilename() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("getting hostname: %w", err)
	}
	return deleteLogPrefix + hostname + deleteLogSuffix, nil
}

// IsDeleteLog reports whether filename matches the per-client delete log
// naming convention (.delete-log.<hostname>.manifest).
func IsDeleteLog(filename string) bool {
	return strings.HasPrefix(filename, deleteLogPrefix) &&
		strings.HasSuffix(filename, deleteLogSuffix)
}

// ===========================================
// Read / Write
// ===========================================

// ReadDeleteLog reads deletion entries from a single log file. Returns nil
// and no error when the file does not exist.
func ReadDeleteLog(path string) ([]DeleteLogEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []DeleteLogEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		t, err := time.Parse(time.RFC3339, parts[0])
		if err != nil {
			continue
		}
		entries = append(entries, DeleteLogEntry{Path: parts[1], DeletedAt: t})
	}
	return entries, scanner.Err()
}

// ReadAllDeleteLogs reads every per-client delete log in targetDir and
// returns the union of all entries.
func ReadAllDeleteLogs(targetDir string) ([]DeleteLogEntry, error) {
	pattern := filepath.Join(targetDir, deleteLogPrefix+"*"+deleteLogSuffix)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var all []DeleteLogEntry
	for _, path := range matches {
		entries, err := ReadDeleteLog(path)
		if err != nil {
			fmt.Printf("[teeleport] rsync: warning reading delete log %s: %v\n", filepath.Base(path), err)
			continue
		}
		all = append(all, entries...)
	}
	return all, nil
}

// WriteDeleteLog writes entries to a delete log file, replacing any
// existing content.
func WriteDeleteLog(path string, entries []DeleteLogEntry) error {
	var b strings.Builder
	for _, e := range entries {
		fmt.Fprintf(&b, "%s\t%s\n", e.DeletedAt.Format(time.RFC3339), e.Path)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// AppendDeletions adds new deletion entries (timestamped now) to the
// current client's delete log inside targetDir.
func AppendDeletions(targetDir string, deletedPaths []string) error {
	logName, err := deleteLogFilename()
	if err != nil {
		return err
	}
	logPath := filepath.Join(targetDir, logName)

	existing, err := ReadDeleteLog(logPath)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	cutoff := now.Add(-deleteLogMaxAge)

	// Prune entries older than deleteLogMaxAge to prevent unbounded growth.
	var pruned []DeleteLogEntry
	for _, e := range existing {
		if !e.DeletedAt.Before(cutoff) {
			pruned = append(pruned, e)
		}
	}

	for _, p := range deletedPaths {
		pruned = append(pruned, DeleteLogEntry{Path: p, DeletedAt: now})
	}
	return WriteDeleteLog(logPath, pruned)
}

// ===========================================
// Processing
// ===========================================

// ShouldDeleteFile reports whether the file at filePath should be deleted
// according to a delete log entry. It returns true only when the file
// exists and its modification time is before the deletion timestamp,
// meaning the file predates the deletion and was not re-created afterwards.
func ShouldDeleteFile(filePath string, entry DeleteLogEntry) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return info.ModTime().Before(entry.DeletedAt)
}

// processDeleteLogEntries reads all per-client delete logs in targetDir,
// determines which local files should be removed (using the latest
// deletion timestamp per path), and deletes them. It returns the list of
// relative paths that were actually removed so the caller can also SSH rm
// them from the server.
func processDeleteLogEntries(targetDir string) ([]string, error) {
	entries, err := ReadAllDeleteLogs(targetDir)
	if err != nil {
		return nil, err
	}

	// Build a map of path -> latest deletion timestamp, ignoring entries
	// older than deleteLogMaxAge to prevent unbounded accumulation.
	cutoff := time.Now().UTC().Add(-deleteLogMaxAge)
	latest := make(map[string]time.Time)
	for _, e := range entries {
		if IsDeleteLog(e.Path) {
			continue
		}
		if e.DeletedAt.Before(cutoff) {
			continue
		}
		if t, ok := latest[e.Path]; !ok || e.DeletedAt.After(t) {
			latest[e.Path] = e.DeletedAt
		}
	}

	var deleted []string
	for path, deletedAt := range latest {
		localPath := filepath.Join(targetDir, path)
		entry := DeleteLogEntry{Path: path, DeletedAt: deletedAt}
		if ShouldDeleteFile(localPath, entry) {
			if err := os.Remove(localPath); err != nil && !os.IsNotExist(err) {
				fmt.Printf("[teeleport] rsync: warning deleting %s: %v\n", path, err)
				continue
			}
			deleted = append(deleted, path)
		}
	}
	return deleted, nil
}

// ===========================================
// Delete-log rsync helpers
// ===========================================

// pushOurDeleteLog pushes this client's delete log file to the remote
// server so other clients can discover our deletions.
func pushOurDeleteLog(ssh domainmodel.SSHConfig, targetDir, remoteDir string) error {
	logName, err := deleteLogFilename()
	if err != nil {
		return err
	}
	localPath := filepath.Join(targetDir, logName)
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return nil // nothing to push
	}
	remotePath := remoteDir + "/" + logName
	return runRsync(ssh, localPath, remotePath, false, "push", false)
}

// pullAllDeleteLogs pulls every per-client delete log from the remote
// server into the local target directory without touching other files.
func pullAllDeleteLogs(ssh domainmodel.SSHConfig, targetDir, remoteDir string) error {
	sshCmd := buildSSHCommand(ssh)

	remoteAddr := remoteAddress(ssh, remoteDir)
	if !strings.HasSuffix(remoteAddr, "/") {
		remoteAddr += "/"
	}
	localDir := targetDir
	if !strings.HasSuffix(localDir, "/") {
		localDir += "/"
	}

	args := []string{
		"-az",
		"-e", sshCmd,
		"--update",
		"--include=" + deleteLogPrefix + "*" + deleteLogSuffix,
		"--exclude=*",
		remoteAddr,
		localDir,
	}

	cmd := exec.Command("rsync", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pull delete logs failed: %w\noutput: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// ===========================================
// Cleanup
// ===========================================

// cleanupStaleLogs removes per-client delete log files whose mtime
// exceeds staleLogMaxAge, both locally and from the remote server.
func cleanupStaleLogs(ssh domainmodel.SSHConfig, targetDir, remoteDir string) {
	pattern := filepath.Join(targetDir, deleteLogPrefix+"*"+deleteLogSuffix)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	now := time.Now()
	var stale []string
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) > staleLogMaxAge {
			stale = append(stale, filepath.Base(path))
			os.Remove(path)
		}
	}

	if len(stale) > 0 {
		fmt.Printf("[teeleport] rsync: cleaning up %d stale delete log(s)\n", len(stale))
		if err := deleteRemoteFiles(ssh, remoteDir, stale, true); err != nil {
			fmt.Printf("[teeleport] rsync: warning cleaning up remote delete logs: %v\n", err)
		}
	}
}

// cleanupAllStaleLogs runs stale delete-log cleanup for every rsync
// directory entry.
func cleanupAllStaleLogs(cfg DaemonConfig) {
	for _, entry := range cfg.Entries {
		target := config.ExpandPath(entry.Target)
		info, err := os.Stat(target)
		if err != nil || !info.IsDir() {
			continue
		}
		cleanupStaleLogs(cfg.SSH, target, entry.Source)
	}
}
