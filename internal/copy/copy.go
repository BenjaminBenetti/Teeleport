package copy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
)

// ProcessCopies processes each copy entry by resolving paths and then either
// replacing or appending the source content into the target file. It logs
// progress to stdout and continues past individual failures rather than
// aborting early.
//
// dotfileRepo is the absolute path to the root of the dotfile repository, used
// to resolve relative source paths in each entry.
//
// entries is the list of CopyEntry values describing the source, target, and
// copy mode for each file operation.
//
// ProcessCopies returns a non-nil error that summarises all failed entry names
// when one or more individual copy operations fail. It returns nil when every
// entry is processed successfully.
func ProcessCopies(dotfileRepo string, entries []config.CopyEntry) error {
	var failures []string

	for _, entry := range entries {
		sourcePath := config.ResolvePath(dotfileRepo, entry.Source)
		targetPath := config.ExpandPath(entry.Target)

		fmt.Printf("[teeleport] copy: %s → %s (%s) ... ", entry.Name, entry.Target, entry.Mode)

		sourceContent, err := os.ReadFile(sourcePath)
		if err != nil {
			fmt.Println("FAIL")
			fmt.Printf("[teeleport] copy: error reading source %s: %v\n", sourcePath, err)
			failures = append(failures, entry.Name)
			continue
		}

		switch strings.ToLower(entry.Mode) {
		case "replace", "":
			if err := replaceFile(targetPath, sourceContent); err != nil {
				fmt.Println("FAIL")
				fmt.Printf("[teeleport] copy: error writing target %s: %v\n", targetPath, err)
				failures = append(failures, entry.Name)
				continue
			}
		case "append":
			if err := ApplyAppend(entry.Name, string(sourceContent), targetPath); err != nil {
				fmt.Println("FAIL")
				fmt.Printf("[teeleport] copy: error appending to %s: %v\n", targetPath, err)
				failures = append(failures, entry.Name)
				continue
			}
		default:
			fmt.Println("FAIL")
			fmt.Printf("[teeleport] copy: unknown mode %q for entry %s\n", entry.Mode, entry.Name)
			failures = append(failures, entry.Name)
			continue
		}

		fmt.Println("ok")
	}

	if len(failures) > 0 {
		return fmt.Errorf("copy failures: %s", strings.Join(failures, ", "))
	}
	return nil
}

// replaceFile writes content to targetPath, creating parent directories as needed.
func replaceFile(targetPath string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("creating parent directories: %w", err)
	}
	if err := os.WriteFile(targetPath, content, 0o644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}
