package rsync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectDeletions_Basic(t *testing.T) {
	tests := []struct {
		name     string
		previous []string
		current  []string
		want     []string
	}{
		{"no deletions", []string{"a", "b", "c"}, []string{"a", "b", "c"}, nil},
		{"one deleted", []string{"a", "b", "c"}, []string{"a", "c"}, []string{"b"}},
		{"all deleted", []string{"a", "b"}, []string{}, []string{"a", "b"}},
		{"empty previous", nil, []string{"a"}, nil},
		{"both empty", nil, nil, nil},
		{"new file added", []string{"a"}, []string{"a", "b"}, nil},
		{"mixed add and delete", []string{"a", "b"}, []string{"b", "c"}, []string{"a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectDeletions(tt.previous, tt.current)
			if !slicesEqual(got, tt.want) {
				t.Errorf("DetectDeletions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManifestRoundTrip(t *testing.T) {
	// Override manifest dir to a temp directory
	tmpDir := t.TempDir()
	origDir := manifestDir
	manifestDir = func() string { return tmpDir }
	defer func() { manifestDir = origDir }()

	entryName := "test-entry"
	files := []string{"dir/file1.txt", "file2.txt", "zzz.txt", "aaa.txt"}

	if err := SaveManifest(entryName, files); err != nil {
		t.Fatalf("SaveManifest() error: %v", err)
	}

	loaded, err := LoadManifest(entryName)
	if err != nil {
		t.Fatalf("LoadManifest() error: %v", err)
	}

	// SaveManifest sorts, so expect sorted output
	want := []string{"aaa.txt", "dir/file1.txt", "file2.txt", "zzz.txt"}
	if !slicesEqual(loaded, want) {
		t.Errorf("LoadManifest() = %v, want %v", loaded, want)
	}
}

func TestLoadManifest_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir := manifestDir
	manifestDir = func() string { return tmpDir }
	defer func() { manifestDir = origDir }()

	files, err := LoadManifest("nonexistent")
	if err != nil {
		t.Fatalf("LoadManifest() unexpected error: %v", err)
	}
	if files != nil {
		t.Errorf("LoadManifest() = %v, want nil", files)
	}
}

func TestScanLocalFiles_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file structure
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"b.txt", "a.txt", filepath.Join("sub", "c.txt")} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := ScanLocalFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("ScanLocalFiles() error: %v", err)
	}

	want := []string{"a.txt", "b.txt", "sub/c.txt"}
	if !slicesEqual(got, want) {
		t.Errorf("ScanLocalFiles() = %v, want %v", got, want)
	}
}

func TestScanLocalFiles_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ScanLocalFiles(filePath, false)
	if err != nil {
		t.Fatalf("ScanLocalFiles() error: %v", err)
	}

	want := []string{"config.yaml"}
	if !slicesEqual(got, want) {
		t.Errorf("ScanLocalFiles() = %v, want %v", got, want)
	}
}

func TestScanLocalFiles_ExcludesDeleteLogs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create normal files and a delete log file
	for _, name := range []string{"a.txt", ".delete-log.myhost.manifest"} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := ScanLocalFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("ScanLocalFiles() error: %v", err)
	}

	want := []string{"a.txt"}
	if !slicesEqual(got, want) {
		t.Errorf("ScanLocalFiles() = %v, want %v (delete log should be excluded)", got, want)
	}
}

func TestScanLocalFiles_MissingFile(t *testing.T) {
	got, err := ScanLocalFiles("/tmp/nonexistent-file-12345", false)
	if err != nil {
		t.Fatalf("ScanLocalFiles() unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("ScanLocalFiles() = %v, want nil", got)
	}
}

// slicesEqual compares two string slices for equality, treating nil and empty
// as equivalent.
func slicesEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
