package mount

import (
	"fmt"

	"github.com/BenjaminBenetti/Teeleport/internal/domainmodel"
)

// MountBackend is the interface that all mount backends must implement.
// Each backend wraps a specific filesystem-mounting tool (e.g. sshfs) and
// provides methods for installing the tool, performing a mount, and querying
// whether a path is already mounted.
type MountBackend interface {
	// EnsureInstalled checks that the underlying mount tool is available on
	// the system, installing it via the package manager if it is missing.
	// It returns a non-nil error if the tool cannot be found or installed.
	EnsureInstalled() error

	// Mount mounts the remote filesystem described by source onto the local
	// directory target. source is a backend-specific location string (for
	// sshfs this is the remote path on the SSH host). target is the absolute
	// path of the local mount point directory, which must already exist.
	// It returns a non-nil error if the mount operation fails.
	Mount(source, target string) error

	// IsMounted reports whether target is currently an active mount point.
	// target is the absolute path of the local directory to check.
	// It returns true if the path is mounted, false otherwise, and a non-nil
	// error if the check itself fails.
	IsMounted(target string) (bool, error)

	// FsType returns the expected filesystem type string for this backend
	// as it appears in /proc/mounts (e.g. "fuse.sshfs").
	FsType() string
}

// NewBackend creates a MountBackend for the given backend name.
// Currently only "sshfs" is supported.
//
// backendName selects the mount backend implementation (e.g. "sshfs").
// ssh provides the SSH connection parameters used by the backend.
// perms provides the UID/GID ownership settings applied to mounted files.
//
// It returns the initialised MountBackend and a nil error on success, or a
// nil MountBackend and a descriptive error if backendName is not recognised.
func NewBackend(backendName string, ssh domainmodel.SSHConfig, perms domainmodel.PermConfig) (MountBackend, error) {
	switch backendName {
	case "sshfs":
		return &SSHFSBackend{SSH: ssh, Perms: perms}, nil
	default:
		return nil, fmt.Errorf("unknown mount backend: %q", backendName)
	}
}
