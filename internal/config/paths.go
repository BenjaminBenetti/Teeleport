package config

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands a leading tilde (~) in path to the current user's home
// directory. If path is exactly "~", it returns the home directory. If path
// begins with "~/", the prefix is replaced with the home directory. All other
// values of path, including the empty string, are returned unchanged. If the
// home directory cannot be determined, path is returned unmodified.
//
// Parameters:
//   - path: the file-system path to expand.
//
// Returns the expanded path as a string.
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

// ResolvePath resolves relative against the base directory. Both values are
// first passed through [ExpandPath] so that tilde prefixes are expanded.
// If the expanded relative path is already absolute, it is returned directly.
// Otherwise it is joined to the expanded base directory using
// [filepath.Join].
//
// Parameters:
//   - base: the directory to resolve against when relative is not absolute.
//   - relative: the path to resolve; may be absolute, relative, or
//     tilde-prefixed.
//
// Returns the resolved absolute path as a string.
func ResolvePath(base, relative string) string {
	expanded := ExpandPath(relative)
	if filepath.IsAbs(expanded) {
		return expanded
	}
	return filepath.Join(ExpandPath(base), expanded)
}
