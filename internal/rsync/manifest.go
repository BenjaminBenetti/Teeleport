package rsync

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
)

// manifestDir returns the directory where per-entry manifests are stored.
// It is a variable so tests can override it with a temporary directory.
var manifestDir = func() string {
	return config.ExpandPath("~/.teeleport/manifests")
}

// manifestPath returns the manifest file path for a given mount entry name.
func manifestPath(entryName string) string {
	return filepath.Join(manifestDir(), entryName+".manifest")
}

// LoadManifest reads the manifest for the given entry name and returns the
// sorted list of relative file paths. If no manifest exists (e.g. first boot),
// it returns an empty slice and no error.
func LoadManifest(entryName string) ([]string, error) {
	data, err := os.ReadFile(manifestPath(entryName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil, nil
	}
	return strings.Split(content, "\n"), nil
}

// SaveManifest writes the sorted file list as the manifest for the given entry.
func SaveManifest(entryName string, files []string) error {
	dir := manifestDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	sorted := make([]string, len(files))
	copy(sorted, files)
	sort.Strings(sorted)
	return os.WriteFile(manifestPath(entryName), []byte(strings.Join(sorted, "\n")+"\n"), 0o644)
}

// ScanLocalFiles walks the directory at localPath and returns a sorted list of
// relative file paths. For single-file entries (isDir == false), it returns a
// slice containing just the base filename if the file exists, or an empty
// slice if it does not.
func ScanLocalFiles(localPath string, isDir bool) ([]string, error) {
	if !isDir {
		if _, err := os.Stat(localPath); err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			return nil, err
		}
		return []string{filepath.Base(localPath)}, nil
	}

	var files []string
	err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if IsDeleteLog(info.Name()) {
			return nil
		}
		rel, err := filepath.Rel(localPath, path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// DetectDeletions returns files that are present in previous but absent from
// current. Both slices must be sorted. The result is the set difference:
// previous - current.
func DetectDeletions(previous, current []string) []string {
	currentSet := make(map[string]struct{}, len(current))
	for _, f := range current {
		currentSet[f] = struct{}{}
	}
	var deleted []string
	for _, f := range previous {
		if _, exists := currentSet[f]; !exists {
			deleted = append(deleted, f)
		}
	}
	return deleted
}
