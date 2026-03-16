package mount

import (
	"fmt"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
)

// MountBackend is the interface that mount backends must implement.
type MountBackend interface {
	// EnsureInstalled checks that the backend tool is available, installing it if needed.
	EnsureInstalled() error
	// Mount mounts the given source to the target path.
	Mount(source, target string) error
	// IsMounted reports whether the target path is already a mount point.
	IsMounted(target string) (bool, error)
}

// NewBackend creates a MountBackend for the given backend name.
// Currently only "sshfs" is supported.
func NewBackend(backendName string, ssh config.SSHConfig, perms config.PermConfig) (MountBackend, error) {
	switch backendName {
	case "sshfs":
		return &SSHFSBackend{SSH: ssh, Perms: perms}, nil
	default:
		return nil, fmt.Errorf("unknown mount backend: %q", backendName)
	}
}
