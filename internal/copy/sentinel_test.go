package copy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyAppend_NewFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "newfile.conf")

	err := ApplyAppend("myblock", "line1\nline2\n", target)
	if err != nil {
		t.Fatalf("ApplyAppend returned error: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading target: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# BEGIN TEELEPORT: myblock") {
		t.Error("missing BEGIN sentinel marker")
	}
	if !strings.Contains(content, "# END TEELEPORT: myblock") {
		t.Error("missing END sentinel marker")
	}
	if !strings.Contains(content, "line1\nline2\n") {
		t.Error("source content not found in output")
	}
}

func TestApplyAppend_ExistingFileWithoutSentinel(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "existing.conf")

	original := "# some existing config\nkey=value\n"
	if err := os.WriteFile(target, []byte(original), 0o644); err != nil {
		t.Fatalf("writing initial file: %v", err)
	}

	err := ApplyAppend("addon", "extra=true\n", target)
	if err != nil {
		t.Fatalf("ApplyAppend returned error: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading target: %v", err)
	}

	content := string(data)

	// Original content should still be present.
	if !strings.Contains(content, "# some existing config") {
		t.Error("original content missing")
	}
	if !strings.Contains(content, "key=value") {
		t.Error("original key=value missing")
	}

	// Sentinel block should be appended.
	if !strings.Contains(content, "# BEGIN TEELEPORT: addon") {
		t.Error("missing BEGIN sentinel marker")
	}
	if !strings.Contains(content, "extra=true") {
		t.Error("source content not found")
	}
	if !strings.Contains(content, "# END TEELEPORT: addon") {
		t.Error("missing END sentinel marker")
	}
}

func TestApplyAppend_ReplacesExistingBlock(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "replace.conf")

	initial := "before\n# BEGIN TEELEPORT: block1\nold content\n# END TEELEPORT: block1\nafter\n"
	if err := os.WriteFile(target, []byte(initial), 0o644); err != nil {
		t.Fatalf("writing initial file: %v", err)
	}

	err := ApplyAppend("block1", "new content\n", target)
	if err != nil {
		t.Fatalf("ApplyAppend returned error: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading target: %v", err)
	}

	content := string(data)

	// Old content should be gone.
	if strings.Contains(content, "old content") {
		t.Error("old content should have been replaced")
	}

	// New content should be present.
	if !strings.Contains(content, "new content") {
		t.Error("new content not found")
	}

	// Surrounding content should be preserved.
	if !strings.Contains(content, "before") {
		t.Error("content before block should be preserved")
	}
	if !strings.Contains(content, "after") {
		t.Error("content after block should be preserved")
	}

	// Only one BEGIN/END pair for block1 should exist.
	if strings.Count(content, "# BEGIN TEELEPORT: block1") != 1 {
		t.Errorf("expected exactly 1 BEGIN marker, got %d", strings.Count(content, "# BEGIN TEELEPORT: block1"))
	}
	if strings.Count(content, "# END TEELEPORT: block1") != 1 {
		t.Errorf("expected exactly 1 END marker, got %d", strings.Count(content, "# END TEELEPORT: block1"))
	}
}

func TestApplyAppend_MultipleNamedBlocks(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "multi.conf")

	// Insert first block.
	if err := ApplyAppend("alpha", "alpha-content\n", target); err != nil {
		t.Fatalf("ApplyAppend alpha: %v", err)
	}

	// Insert second block.
	if err := ApplyAppend("beta", "beta-content\n", target); err != nil {
		t.Fatalf("ApplyAppend beta: %v", err)
	}

	// Now update only the alpha block.
	if err := ApplyAppend("alpha", "alpha-updated\n", target); err != nil {
		t.Fatalf("ApplyAppend alpha update: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading target: %v", err)
	}

	content := string(data)

	// Alpha should have updated content.
	if strings.Contains(content, "alpha-content") {
		t.Error("old alpha-content should have been replaced")
	}
	if !strings.Contains(content, "alpha-updated") {
		t.Error("alpha-updated not found")
	}

	// Beta should be untouched.
	if !strings.Contains(content, "beta-content") {
		t.Error("beta-content should still be present")
	}

	// One of each named block.
	if strings.Count(content, "# BEGIN TEELEPORT: alpha") != 1 {
		t.Error("expected exactly 1 alpha BEGIN marker")
	}
	if strings.Count(content, "# END TEELEPORT: alpha") != 1 {
		t.Error("expected exactly 1 alpha END marker")
	}
	if strings.Count(content, "# BEGIN TEELEPORT: beta") != 1 {
		t.Error("expected exactly 1 beta BEGIN marker")
	}
	if strings.Count(content, "# END TEELEPORT: beta") != 1 {
		t.Error("expected exactly 1 beta END marker")
	}
}

func TestApplyAppend_SentinelBlockFormat(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "format.conf")

	err := ApplyAppend("testname", "body\n", target)
	if err != nil {
		t.Fatalf("ApplyAppend returned error: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading target: %v", err)
	}

	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")

	// Verify exact sentinel format.
	if lines[0] != "# BEGIN TEELEPORT: testname" {
		t.Errorf("first line = %q, want %q", lines[0], "# BEGIN TEELEPORT: testname")
	}
	if lines[len(lines)-1] != "# END TEELEPORT: testname" {
		t.Errorf("last line = %q, want %q", lines[len(lines)-1], "# END TEELEPORT: testname")
	}
}

func TestApplyAppend_ContentWithoutTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "nonewline.conf")

	// Source content without trailing newline — should still produce valid block.
	err := ApplyAppend("notrail", "content-no-newline", target)
	if err != nil {
		t.Fatalf("ApplyAppend returned error: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading target: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "content-no-newline") {
		t.Error("content not found")
	}
	if !strings.Contains(content, "# BEGIN TEELEPORT: notrail") {
		t.Error("missing BEGIN marker")
	}
	if !strings.Contains(content, "# END TEELEPORT: notrail") {
		t.Error("missing END marker")
	}

	// The END marker should be on its own line, not concatenated to content.
	idx := strings.Index(content, "# END TEELEPORT: notrail")
	if idx > 0 && content[idx-1] != '\n' {
		t.Error("END marker is not on its own line")
	}
}

func TestApplyAppend_NestedSubdirectoryCreated(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "sub", "dir", "nested.conf")

	err := ApplyAppend("nested", "data\n", target)
	if err != nil {
		t.Fatalf("ApplyAppend should create parent directories, got error: %v", err)
	}

	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Fatal("target file was not created")
	}
}
