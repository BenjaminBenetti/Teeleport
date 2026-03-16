package copy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	sentinelBeginPrefix = "# BEGIN TEELEPORT: "
	sentinelEndPrefix   = "# END TEELEPORT: "
)

// ApplyAppend inserts or replaces a sentinel-delimited block in the target file.
// The block is identified by name and delimited by:
//
//	# BEGIN TEELEPORT: <name>
//	<sourceContent>
//	# END TEELEPORT: <name>
//
// If the block already exists, it is replaced (markers inclusive).
// If it does not exist, it is appended at the end of the file.
func ApplyAppend(name, sourceContent, targetPath string) error {
	beginMarker := sentinelBeginPrefix + name
	endMarker := sentinelEndPrefix + name

	block := beginMarker + "\n" + sourceContent
	// Ensure the content inside the block ends with a newline before the end marker.
	if !strings.HasSuffix(block, "\n") {
		block += "\n"
	}
	block += endMarker

	// Read existing target file content; if the file doesn't exist, start empty.
	existing, err := os.ReadFile(targetPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("reading target file: %w", err)
		}
		existing = nil
		// Ensure parent directories exist for new files.
		if mkErr := os.MkdirAll(filepath.Dir(targetPath), 0o755); mkErr != nil {
			return fmt.Errorf("creating parent directories: %w", mkErr)
		}
	}

	content := string(existing)

	beginIdx := strings.Index(content, beginMarker)
	endIdx := strings.Index(content, endMarker)

	if beginIdx >= 0 && endIdx >= 0 && endIdx > beginIdx {
		// Replace the existing block (markers inclusive).
		before := content[:beginIdx]
		after := content[endIdx+len(endMarker):]
		content = before + block + after
	} else {
		// Append the block at the end.
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += block + "\n"
	}

	if err := os.WriteFile(targetPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing target file: %w", err)
	}
	return nil
}
