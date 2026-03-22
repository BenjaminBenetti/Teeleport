package mount

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
	"github.com/BenjaminBenetti/Teeleport/internal/domainmodel"
	"github.com/BenjaminBenetti/Teeleport/internal/packages"
)

// SSHFSBackend implements MountBackend using the sshfs command-line tool.
// It connects to a remote host over SSH and exposes a remote directory as a
// local FUSE mount.
type SSHFSBackend struct {
	// SSH holds the SSH connection parameters (host, port, user, identity file)
	// used when invoking the sshfs command.
	SSH domainmodel.SSHConfig

	// Perms holds the UID and GID that will own the mounted files on the local
	// filesystem.
	Perms domainmodel.PermConfig
}

// EnsureInstalled checks whether the sshfs binary is available on the system
// PATH. If it is not found, EnsureInstalled attempts to install it via the
// host package manager. It returns a non-nil error if sshfs cannot be made
// available.
func (b *SSHFSBackend) EnsureInstalled() error {
	if _, err := exec.LookPath("sshfs"); err == nil {
		return nil
	}

	fmt.Println("[teeleport] mount: sshfs not found, attempting install...")
	if installErr := packages.Run([]string{"sshfs"}); installErr != nil {
		return fmt.Errorf("sshfs not available and install failed: %w", installErr)
	}

	if _, err := exec.LookPath("sshfs"); err != nil {
		return fmt.Errorf("sshfs still not available after install attempt")
	}

	fmt.Println("[teeleport] mount: sshfs installed ✓")
	return nil
}

// Mount mounts a remote directory onto a local path using the sshfs command.
//
// source is the absolute path of the directory on the remote SSH host.
// target is the absolute path of the local mount point directory; it must
// already exist.
//
// The SSH user is taken from b.SSH.User; if that field is empty the current
// OS user is used instead. Ownership of mounted files is set according to
// b.Perms. The connection is configured with automatic reconnection and
// periodic keepalive probes.
//
// It returns a non-nil error if the sshfs process exits with a failure status.
func (b *SSHFSBackend) Mount(source, target string) error {
	sshUser := b.SSH.User
	if sshUser == "" {
		u, err := user.Current()
		if err != nil {
			return fmt.Errorf("determining current user: %w", err)
		}
		sshUser = u.Username
	}

	remote := fmt.Sprintf("%s@%s:%s", sshUser, b.SSH.Host, source)

	args := []string{
		remote,
		target,
		"-o", fmt.Sprintf("port=%d", b.SSH.Port),
		"-o", fmt.Sprintf("uid=%d", *b.Perms.UID),
		"-o", fmt.Sprintf("gid=%d", *b.Perms.GID),
	}

	if b.SSH.IdentityFile != "" {
		args = append(args, "-o", fmt.Sprintf("IdentityFile=%s", config.ExpandPath(b.SSH.IdentityFile)))
	}

	args = append(args,
		"-o", "StrictHostKeyChecking=no",
		"-o", "reconnect",
		"-o", "ServerAliveInterval=15",
	)

	cmd := exec.Command("sshfs", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sshfs command failed: %w", err)
	}
	return nil
}

// FsType returns the filesystem type string for SSHFS mounts as it appears
// in /proc/mounts.
func (b *SSHFSBackend) FsType() string {
	return "fuse.sshfs"
}

// mountedFsType reads /proc/mounts and returns the filesystem type (third
// field) for the given mount point target, or "" if the target is not mounted
// or /proc/mounts cannot be read.
func mountedFsType(target string) string {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 && fields[1] == target {
			return fields[2]
		}
	}
	return ""
}

// IsMounted reports whether target is currently listed as a mount point in
// /proc/mounts.
//
// target is the absolute path of the local directory to check.
//
// It returns true if target appears as the second field of any line in
// /proc/mounts, false if it does not, and a non-nil error if /proc/mounts
// cannot be read or scanned.
func (b *SSHFSBackend) IsMounted(target string) (bool, error) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return false, fmt.Errorf("reading /proc/mounts: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[1] == target {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("scanning /proc/mounts: %w", err)
	}
	return false, nil
}
