package rsync

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestPidFilePath(t *testing.T) {
	got := pidFilePath()
	if got == "" {
		t.Fatal("pidFilePath() returned empty string")
	}
	if !strings.Contains(got, ".teeleport/rsync.pid") {
		t.Errorf("pidFilePath() = %q, expected to contain .teeleport/rsync.pid", got)
	}
}

func TestIsDaemonRunning_NoPidFile(t *testing.T) {
	// Ensure the pid file doesn't exist (use a temp location)
	origPath := pidFilePath()
	_ = os.Remove(origPath)
	defer func() {
		// Clean up in case the test created the file
		os.Remove(origPath)
	}()

	if pid := IsDaemonRunning(); pid != 0 {
		t.Errorf("IsDaemonRunning() = %d, want 0 (no pid file)", pid)
	}
}

func TestIsDaemonRunning_StalePid(t *testing.T) {
	pidPath := pidFilePath()
	if err := os.MkdirAll(strings.TrimSuffix(pidPath, "/rsync.pid"), 0o755); err != nil {
		t.Skipf("cannot create pid dir: %v", err)
	}

	// Write a PID that almost certainly doesn't exist
	stalePid := 99999
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(stalePid)), 0o644); err != nil {
		t.Fatalf("writing stale pid: %v", err)
	}
	defer os.Remove(pidPath)

	if pid := IsDaemonRunning(); pid != 0 {
		t.Errorf("IsDaemonRunning() = %d, want 0 (stale pid)", pid)
	}
}

func TestIsDaemonRunning_CurrentProcess(t *testing.T) {
	pidPath := pidFilePath()
	if err := os.MkdirAll(strings.TrimSuffix(pidPath, "/rsync.pid"), 0o755); err != nil {
		t.Skipf("cannot create pid dir: %v", err)
	}

	// Write our own PID — we are alive
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		t.Fatalf("writing pid: %v", err)
	}
	defer os.Remove(pidPath)

	if pid := IsDaemonRunning(); pid != os.Getpid() {
		t.Errorf("IsDaemonRunning() = %d, want %d (own pid)", pid, os.Getpid())
	}
}
