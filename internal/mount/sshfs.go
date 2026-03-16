package mount

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
	"github.com/BenjaminBenetti/Teeleport/internal/packages"
)

// SSHFSBackend implements MountBackend using the sshfs command.
type SSHFSBackend struct {
	SSH   config.SSHConfig
	Perms config.PermConfig
}

// EnsureInstalled checks if sshfs is available and installs it if missing.
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

// Mount mounts source to target via sshfs.
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

// IsMounted checks whether target appears as a mount point in /proc/mounts.
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
