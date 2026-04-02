package mount

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/BenjaminBenetti/Teeleport/internal/domainmodel"
	"github.com/BenjaminBenetti/Teeleport/internal/packages"
	rsyncpkg "github.com/BenjaminBenetti/Teeleport/internal/rsync"
)

// RsyncBackend implements MountBackend using rsync over SSH. Unlike SSHFS,
// rsync does not create a FUSE mount — it copies files directly. The "Mount"
// operation performs an initial pull from the remote, and ongoing
// synchronisation is handled by the rsync daemon.
type RsyncBackend struct {
	SSH   domainmodel.SSHConfig
	Perms domainmodel.PermConfig
}

// EnsureInstalled checks that the rsync binary is available, installing it
// via the system package manager if necessary.
func (b *RsyncBackend) EnsureInstalled() error {
	if _, err := exec.LookPath("rsync"); err == nil {
		return nil
	}

	fmt.Println("[teeleport] mount: rsync not found, attempting install...")
	if installErr := packages.Run([]string{"rsync"}); installErr != nil {
		return fmt.Errorf("rsync not available and install failed: %w", installErr)
	}

	if _, err := exec.LookPath("rsync"); err != nil {
		return fmt.Errorf("rsync still not available after install attempt")
	}

	fmt.Println("[teeleport] mount: rsync installed ✓")
	return nil
}

// Mount performs an initial one-way pull from the remote source into the local
// target. For directories, source contents are synced into target. For files,
// the single file is synced. The --update flag ensures existing newer local
// files are not overwritten.
func (b *RsyncBackend) Mount(source, target string) error {
	isDir := false
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		isDir = true
	}

	if err := rsyncpkg.PullEntry(b.SSH, source, target, isDir); err != nil {
		return fmt.Errorf("rsync initial pull failed: %w", err)
	}
	return nil
}

// IsMounted always returns false for rsync entries. Rsync does not create a
// persistent mount — it copies files. Returning false ensures ProcessMounts
// always performs the initial sync on startup, which is cheap because rsync
// only transfers changed files.
func (b *RsyncBackend) IsMounted(_ string) (bool, error) {
	return false, nil
}

// FsType returns "rsync" as a pseudo filesystem type identifier. This value
// never appears in /proc/mounts (rsync does not create kernel mounts) but is
// used internally by ProcessMounts for backend identification.
func (b *RsyncBackend) FsType() string {
	return "rsync"
}
