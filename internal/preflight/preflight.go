package preflight

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
)

// RunChecks performs preflight checks required before mounting.
// It only runs when mounts are defined. It checks for FUSE availability,
// the sshfs binary, and SSH connectivity to the configured host.
func RunChecks(cfg *config.Config) error {
	if len(cfg.Mounts.Entries) == 0 {
		return nil
	}

	// Check 1: FUSE device
	if _, err := os.Stat("/dev/fuse"); err != nil {
		return fmt.Errorf("FUSE not available. Add \"privileged\": true or \"runArgs\": [\"--device=/dev/fuse\"] to your devcontainer.json")
	}
	fmt.Println("[teeleport] preflight: FUSE ✓")

	// Check 2: SSH connectivity
	sshCfg := cfg.Mounts.SSH
	sshUser := sshCfg.User
	if sshUser == "" {
		u, err := user.Current()
		if err != nil {
			return fmt.Errorf("cannot determine current user: %w", err)
		}
		sshUser = u.Username
	}

	port := fmt.Sprintf("%d", sshCfg.Port)
	target := fmt.Sprintf("%s@%s", sshUser, sshCfg.Host)

	args := []string{
		"-o", "ConnectTimeout=5",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=no",
		"-p", port,
		target,
		"exit",
	}

	cmd := exec.Command("ssh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("SSH connectivity check failed (%s): %w", target, err)
	}
	fmt.Println("[teeleport] preflight: SSH connectivity ✓")

	return nil
}
