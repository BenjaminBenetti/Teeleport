// Package preflight implements pre-mount validation checks for Teeleport.
// It verifies that required system dependencies (FUSE, SSH) are available
// and properly configured before attempting any SSHFS mount operations.
package preflight

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
)

// RunChecks performs preflight checks required before attempting SSHFS mounts.
// It verifies that the FUSE device (/dev/fuse) is accessible and that SSH
// connectivity to the configured host is working. If cfg contains no mount
// entries, RunChecks returns nil immediately without performing any checks.
//
// The cfg parameter supplies the full Teeleport configuration, including SSH
// host, user, and port details used for the connectivity test.
//
// RunChecks returns a descriptive error if any check fails, or nil on success.
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
