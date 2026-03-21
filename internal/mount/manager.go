package mount

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
)

// ProcessMounts iterates over every entry in cfg, ensuring the required
// backend is installed, creating local mount-point directories as needed,
// and performing each mount. Already-mounted targets are silently skipped.
//
// cfg contains the full mount configuration including SSH connection
// parameters, file-ownership permissions, and the list of mount entries.
//
// ProcessMounts is best-effort: if an individual mount fails the error is
// recorded and processing continues with the remaining entries. It returns
// a non-nil aggregate error summarising all failures, or nil if every mount
// succeeded (or was already mounted).
func ProcessMounts(cfg config.MountConfig) error {
	var failures []string

	// Cache backends by name so EnsureInstalled only runs once per backend type.
	backends := make(map[string]MountBackend)

	// Track staging mounts for file mount entries by remote parent directory.
	stagingMounts := make(map[string]string) // remote parent dir -> local staging path

	for _, entry := range cfg.Entries {
		backendName := entry.Backend
		if backendName == "" {
			backendName = "sshfs"
		}

		backend, ok := backends[backendName]
		if !ok {
			var err error
			backend, err = NewBackend(backendName, cfg.SSH, cfg.Permissions)
			if err != nil {
				fmt.Printf("[teeleport] mount: %s → %s ... failed: %v\n", entry.Name, entry.Target, err)
				failures = append(failures, fmt.Sprintf("%s: %v", entry.Name, err))
				continue
			}
			if err := backend.EnsureInstalled(); err != nil {
				fmt.Printf("[teeleport] mount: %s backend install failed: %v\n", backendName, err)
				failures = append(failures, fmt.Sprintf("%s: %v", entry.Name, err))
				continue
			}
			backends[backendName] = backend
		}

		target := config.ExpandPath(entry.Target)

		if isFileMount(entry) {
			// Check if symlink already exists and backing mount is active
			if linkDest, err := os.Readlink(target); err == nil {
				stagingPath := filepath.Dir(linkDest)
				if mounted, mErr := backend.IsMounted(stagingPath); mErr == nil && mounted {
					fmt.Printf("[teeleport] mount: %s → %s ... already mounted, skipping\n", entry.Name, entry.Target)
					continue
				}
			}

			// Derive remote parent and filename
			parentDir := remoteParent(entry.Source)
			basename := remoteBasename(entry.Source)

			// Determine staging path, reuse if same remote parent already mounted
			staging, exists := stagingMounts[parentDir]
			if !exists {
				staging = stagingDir(entry.Name)

				// Check if staging is already mounted
				if mounted, _ := backend.IsMounted(staging); !mounted {
					if err := os.MkdirAll(staging, 0o755); err != nil {
						fmt.Printf("[teeleport] mount: %s → %s ... failed creating staging dir: %v\n", entry.Name, entry.Target, err)
						failures = append(failures, fmt.Sprintf("%s: %v", entry.Name, err))
						continue
					}
					if err := backend.Mount(parentDir, staging); err != nil {
						fmt.Printf("[teeleport] mount: %s → %s ... failed mounting staging: %v\n", entry.Name, entry.Target, err)
						failures = append(failures, fmt.Sprintf("%s: %v", entry.Name, err))
						continue
					}
				}
				stagingMounts[parentDir] = staging
			}

			// Ensure target's parent directory exists
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				fmt.Printf("[teeleport] mount: %s → %s ... failed creating target parent dir: %v\n", entry.Name, entry.Target, err)
				failures = append(failures, fmt.Sprintf("%s: %v", entry.Name, err))
				continue
			}

			// Remove any existing file or stale symlink at target
			os.Remove(target)

			// Create symlink
			symlinkTarget := filepath.Join(staging, basename)
			if err := os.Symlink(symlinkTarget, target); err != nil {
				fmt.Printf("[teeleport] mount: %s → %s ... failed creating symlink: %v\n", entry.Name, entry.Target, err)
				failures = append(failures, fmt.Sprintf("%s: %v", entry.Name, err))
				continue
			}

			// Verify the symlink resolves (file may not exist yet on remote - that's ok)
			if _, err := os.Stat(target); err != nil {
				fmt.Printf("[teeleport] mount: %s → %s ... ok (warning: remote file not yet present)\n", entry.Name, entry.Target)
			} else {
				fmt.Printf("[teeleport] mount: %s → %s ... ok\n", entry.Name, entry.Target)
			}
			continue
		}

		// Directory mount (default behaviour)
		mounted, err := backend.IsMounted(target)
		if err != nil {
			fmt.Printf("[teeleport] mount: %s → %s ... failed checking mount: %v\n", entry.Name, entry.Target, err)
			failures = append(failures, fmt.Sprintf("%s: %v", entry.Name, err))
			continue
		}
		if mounted {
			fmt.Printf("[teeleport] mount: %s → %s ... already mounted, skipping\n", entry.Name, entry.Target)
			continue
		}

		if err := os.MkdirAll(target, 0o755); err != nil {
			fmt.Printf("[teeleport] mount: %s → %s ... failed creating directory: %v\n", entry.Name, entry.Target, err)
			failures = append(failures, fmt.Sprintf("%s: %v", entry.Name, err))
			continue
		}

		if err := backend.Mount(entry.Source, target); err != nil {
			fmt.Printf("[teeleport] mount: %s → %s ... failed: %v\n", entry.Name, entry.Target, err)
			failures = append(failures, fmt.Sprintf("%s: %v", entry.Name, err))
			continue
		}

		fmt.Printf("[teeleport] mount: %s → %s ... ok\n", entry.Name, entry.Target)
	}

	if len(failures) > 0 {
		return fmt.Errorf("mount failures: %s", strings.Join(failures, "; "))
	}
	return nil
}

// isFileMount returns true if the entry is a file mount.
func isFileMount(entry config.MountEntry) bool {
	return entry.Type == "file"
}

// stagingDir returns the staging mount directory for a file mount entry.
// File mounts stage the remote parent directory under ~/.teeleport/mounts/<name>/.
func stagingDir(name string) string {
	return config.ExpandPath(fmt.Sprintf("~/.teeleport/mounts/%s", name))
}

// remoteParent returns the parent directory of a remote path.
// Uses path.Dir (not filepath.Dir) since remote paths are always POSIX.
func remoteParent(source string) string {
	return path.Dir(source)
}

// remoteBasename returns the filename component of a remote path.
func remoteBasename(source string) string {
	return path.Base(source)
}
