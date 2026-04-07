package rsync

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
	"github.com/BenjaminBenetti/Teeleport/internal/domainmodel"
)

// DaemonConfig holds everything the rsync daemon needs to run.
type DaemonConfig struct {
	SSH      domainmodel.SSHConfig
	Entries  []domainmodel.MountEntry // only rsync-backend entries
	Interval time.Duration
}

// pidFilePath returns the path to the rsync daemon PID file.
func pidFilePath() string {
	return config.ExpandPath("~/.teeleport/rsync.pid")
}

// RunDaemon starts the rsync sync loop. It writes a PID file, runs
// bidirectional syncs for every entry at the configured interval, and cleans
// up on SIGTERM/SIGINT. Errors during individual syncs are logged but never
// cause the daemon to exit — it simply retries on the next tick.
func RunDaemon(cfg DaemonConfig) error {
	if len(cfg.Entries) == 0 {
		fmt.Println("[teeleport] rsync daemon: no rsync entries configured, exiting")
		return nil
	}

	// Write PID file
	pidPath := pidFilePath()
	if err := os.MkdirAll(filepath.Dir(pidPath), 0o755); err != nil {
		return fmt.Errorf("creating pid dir: %w", err)
	}
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		return fmt.Errorf("writing pid file: %w", err)
	}
	defer os.Remove(pidPath)

	// Set up signal handling for clean shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	fmt.Printf("[teeleport] rsync daemon: started (pid %d, interval %s, %d entries)\n",
		os.Getpid(), cfg.Interval, len(cfg.Entries))

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	// Run initial sync immediately
	var cycleCount int
	syncAll(cfg)
	cycleCount++

	for {
		select {
		case <-ticker.C:
			syncAll(cfg)
			cycleCount++
			if cycleCount%cleanupCycleInterval == 0 {
				cleanupAllStaleLogs(cfg)
			}
		case sig := <-sigCh:
			fmt.Printf("[teeleport] rsync daemon: received %s, shutting down\n", sig)
			return nil
		}
	}
}

// syncAll performs a single bidirectional sync cycle across all entries.
// Errors are logged and swallowed so the daemon continues running.
func syncAll(cfg DaemonConfig) {
	var errCount int
	for _, entry := range cfg.Entries {
		if err := SyncEntry(cfg.SSH, entry); err != nil {
			fmt.Printf("[teeleport] rsync daemon: sync error for %s: %v\n", entry.Name, err)
			errCount++
		}
	}
	if errCount == 0 {
		fmt.Printf("[teeleport] rsync daemon: sync cycle complete (%d entries)\n", len(cfg.Entries))
	} else {
		fmt.Printf("[teeleport] rsync daemon: sync cycle complete (%d entries, %d errors)\n", len(cfg.Entries), errCount)
	}
}

// IsDaemonRunning checks if a rsync daemon is already running by reading the
// PID file and sending signal 0. Returns the PID if alive, 0 otherwise.
func IsDaemonRunning() int {
	data, err := os.ReadFile(pidFilePath())
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return 0
	}
	// Signal 0 tests if the process exists without actually sending a signal
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return 0
	}
	return pid
}
