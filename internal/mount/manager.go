package mount

import (
	"fmt"
	"os"
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
