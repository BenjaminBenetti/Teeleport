package config

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands a leading ~ to the user's home directory.
func ExpandPath(path string) string {
	if path == "" {
		return path
	}
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// ResolvePath resolves a relative path against a base directory.
// If the path is already absolute, it is returned as-is (after ExpandPath).
// Otherwise it is joined to the base directory.
func ResolvePath(base, relative string) string {
	expanded := ExpandPath(relative)
	if filepath.IsAbs(expanded) {
		return expanded
	}
	return filepath.Join(ExpandPath(base), expanded)
}
