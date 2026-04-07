package rsync

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ===========================================
// IsDeleteLog
// ===========================================

func TestIsDeleteLog(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"valid log", ".delete-log.myhost.manifest", true},
		{"valid with dashes", ".delete-log.code-space-123.manifest", true},
		{"regular file", "readme.txt", false},
		{"partial prefix", ".delete-log.", false},
		{"partial suffix", ".delete-log.host", false},
		{"manifest only", ".manifest", false},
		{"empty string", "", false},
		{"our own manifest", "projects.manifest", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDeleteLog(tt.filename); got != tt.want {
				t.Errorf("IsDeleteLog(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

// ===========================================
// Read / Write round-trip
// ===========================================

func TestDeleteLogRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, ".delete-log.testhost.manifest")

	now := time.Date(2026, 4, 7, 14, 30, 0, 0, time.UTC)
	entries := []DeleteLogEntry{
		{Path: "foo.txt", DeletedAt: now},
		{Path: "sub/bar.txt", DeletedAt: now.Add(5 * time.Minute)},
	}

	if err := WriteDeleteLog(logPath, entries); err != nil {
		t.Fatalf("WriteDeleteLog() error: %v", err)
	}

	loaded, err := ReadDeleteLog(logPath)
	if err != nil {
		t.Fatalf("ReadDeleteLog() error: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("ReadDeleteLog() returned %d entries, want 2", len(loaded))
	}
	if loaded[0].Path != "foo.txt" {
		t.Errorf("entry[0].Path = %q, want %q", loaded[0].Path, "foo.txt")
	}
	if !loaded[0].DeletedAt.Equal(now) {
		t.Errorf("entry[0].DeletedAt = %v, want %v", loaded[0].DeletedAt, now)
	}
	if loaded[1].Path != "sub/bar.txt" {
		t.Errorf("entry[1].Path = %q, want %q", loaded[1].Path, "sub/bar.txt")
	}
}

func TestReadDeleteLog_NoFile(t *testing.T) {
	entries, err := ReadDeleteLog("/tmp/nonexistent-delete-log-12345")
	if err != nil {
		t.Fatalf("ReadDeleteLog() unexpected error: %v", err)
	}
	if entries != nil {
		t.Errorf("ReadDeleteLog() = %v, want nil", entries)
	}
}

func TestReadDeleteLog_MalformedLines(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	content := "bad-timestamp\tfoo.txt\n" +
		"no-tab-here\n" +
		"\n" +
		"2026-04-07T14:30:00Z\tgood.txt\n"
	if err := os.WriteFile(logPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	entries, err := ReadDeleteLog(logPath)
	if err != nil {
		t.Fatalf("ReadDeleteLog() error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("ReadDeleteLog() returned %d entries, want 1 (only the valid line)", len(entries))
	}
	if entries[0].Path != "good.txt" {
		t.Errorf("entry.Path = %q, want %q", entries[0].Path, "good.txt")
	}
}

// ===========================================
// ReadAllDeleteLogs
// ===========================================

func TestReadAllDeleteLogs(t *testing.T) {
	tmpDir := t.TempDir()

	now := time.Date(2026, 4, 7, 14, 0, 0, 0, time.UTC)

	// Client A's log
	WriteDeleteLog(filepath.Join(tmpDir, ".delete-log.client-a.manifest"), []DeleteLogEntry{
		{Path: "foo.txt", DeletedAt: now},
	})
	// Client B's log
	WriteDeleteLog(filepath.Join(tmpDir, ".delete-log.client-b.manifest"), []DeleteLogEntry{
		{Path: "bar.txt", DeletedAt: now.Add(time.Minute)},
	})
	// A regular file that should not be read
	os.WriteFile(filepath.Join(tmpDir, "regular.txt"), []byte("hello"), 0o644)

	entries, err := ReadAllDeleteLogs(tmpDir)
	if err != nil {
		t.Fatalf("ReadAllDeleteLogs() error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("ReadAllDeleteLogs() returned %d entries, want 2", len(entries))
	}

	paths := map[string]bool{}
	for _, e := range entries {
		paths[e.Path] = true
	}
	if !paths["foo.txt"] || !paths["bar.txt"] {
		t.Errorf("expected foo.txt and bar.txt, got %v", paths)
	}
}

// ===========================================
// AppendDeletions
// ===========================================

func TestAppendDeletions(t *testing.T) {
	tmpDir := t.TempDir()

	if err := AppendDeletions(tmpDir, []string{"first.txt"}); err != nil {
		t.Fatalf("first AppendDeletions() error: %v", err)
	}
	if err := AppendDeletions(tmpDir, []string{"second.txt", "third.txt"}); err != nil {
		t.Fatalf("second AppendDeletions() error: %v", err)
	}

	logName, _ := deleteLogFilename()
	entries, err := ReadDeleteLog(filepath.Join(tmpDir, logName))
	if err != nil {
		t.Fatalf("ReadDeleteLog() error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}
	if entries[0].Path != "first.txt" {
		t.Errorf("entry[0].Path = %q, want first.txt", entries[0].Path)
	}
}

// ===========================================
// ShouldDeleteFile
// ===========================================

func TestShouldDeleteFile_OlderThanDeletion(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "old.txt")
	os.WriteFile(filePath, []byte("x"), 0o644)

	// Set mtime to the past
	past := time.Now().Add(-1 * time.Hour)
	os.Chtimes(filePath, past, past)

	entry := DeleteLogEntry{Path: "old.txt", DeletedAt: time.Now()}
	if !ShouldDeleteFile(filePath, entry) {
		t.Error("ShouldDeleteFile() = false, want true (file is older than deletion)")
	}
}

func TestShouldDeleteFile_NewerThanDeletion(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "new.txt")
	os.WriteFile(filePath, []byte("x"), 0o644)

	// Deletion happened in the past, file was re-created recently
	entry := DeleteLogEntry{Path: "new.txt", DeletedAt: time.Now().Add(-1 * time.Hour)}
	if ShouldDeleteFile(filePath, entry) {
		t.Error("ShouldDeleteFile() = true, want false (file is newer than deletion)")
	}
}

func TestShouldDeleteFile_FileDoesNotExist(t *testing.T) {
	entry := DeleteLogEntry{Path: "gone.txt", DeletedAt: time.Now()}
	if ShouldDeleteFile("/tmp/nonexistent-12345", entry) {
		t.Error("ShouldDeleteFile() = true, want false (file does not exist)")
	}
}

// ===========================================
// processDeleteLogEntries
// ===========================================

func TestProcessDeleteLogEntries(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files to be deleted
	oldFile := filepath.Join(tmpDir, "old.txt")
	os.WriteFile(oldFile, []byte("x"), 0o644)
	past := time.Now().Add(-1 * time.Hour)
	os.Chtimes(oldFile, past, past)

	// Create a file that should survive (re-created after deletion)
	newFile := filepath.Join(tmpDir, "new.txt")
	os.WriteFile(newFile, []byte("x"), 0o644)

	// Write a delete log with entries for both files
	deletionTime := time.Now().Add(-30 * time.Minute) // between old mtime and new mtime
	WriteDeleteLog(filepath.Join(tmpDir, ".delete-log.test.manifest"), []DeleteLogEntry{
		{Path: "old.txt", DeletedAt: deletionTime},
		{Path: "new.txt", DeletedAt: deletionTime},
		{Path: "missing.txt", DeletedAt: deletionTime}, // doesn't exist
	})

	deleted, err := processDeleteLogEntries(tmpDir)
	if err != nil {
		t.Fatalf("processDeleteLogEntries() error: %v", err)
	}

	// old.txt should be deleted (mtime < deletionTime)
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("old.txt should have been deleted")
	}

	// new.txt should survive (mtime > deletionTime)
	if _, err := os.Stat(newFile); err != nil {
		t.Error("new.txt should NOT have been deleted (re-created after deletion)")
	}

	// Only old.txt should be in the returned list
	if len(deleted) != 1 || deleted[0] != "old.txt" {
		t.Errorf("processDeleteLogEntries() returned %v, want [old.txt]", deleted)
	}
}

func TestProcessDeleteLogEntries_LatestTimestampWins(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with mtime between T1 and T2
	filePath := filepath.Join(tmpDir, "target.txt")
	os.WriteFile(filePath, []byte("x"), 0o644)
	fileMtime := time.Now().Add(-30 * time.Minute)
	os.Chtimes(filePath, fileMtime, fileMtime)

	T1 := time.Now().Add(-1 * time.Hour)  // older than file — would NOT delete
	T2 := time.Now().Add(-10 * time.Minute) // newer than file — WOULD delete

	// Client A logged deletion at T1, Client B at T2
	WriteDeleteLog(filepath.Join(tmpDir, ".delete-log.client-a.manifest"), []DeleteLogEntry{
		{Path: "target.txt", DeletedAt: T1},
	})
	WriteDeleteLog(filepath.Join(tmpDir, ".delete-log.client-b.manifest"), []DeleteLogEntry{
		{Path: "target.txt", DeletedAt: T2},
	})

	deleted, err := processDeleteLogEntries(tmpDir)
	if err != nil {
		t.Fatalf("processDeleteLogEntries() error: %v", err)
	}

	// T2 wins → file should be deleted (mtime < T2)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("target.txt should have been deleted (latest timestamp T2 > file mtime)")
	}
	if len(deleted) != 1 || deleted[0] != "target.txt" {
		t.Errorf("returned %v, want [target.txt]", deleted)
	}
}

func TestProcessDeleteLogEntries_SkipsDeleteLogFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a delete log that references another delete log
	logFile := filepath.Join(tmpDir, ".delete-log.attacker.manifest")
	WriteDeleteLog(logFile, []DeleteLogEntry{
		{Path: ".delete-log.victim.manifest", DeletedAt: time.Now()},
	})

	// Create the "victim" delete log
	victimLog := filepath.Join(tmpDir, ".delete-log.victim.manifest")
	WriteDeleteLog(victimLog, []DeleteLogEntry{
		{Path: "legit.txt", DeletedAt: time.Now()},
	})

	deleted, _ := processDeleteLogEntries(tmpDir)

	// The victim log should NOT be deleted
	if _, err := os.Stat(victimLog); err != nil {
		t.Error(".delete-log.victim.manifest should NOT be deleted by log processing")
	}
	if len(deleted) != 0 {
		t.Errorf("expected no deletions, got %v", deleted)
	}
}
