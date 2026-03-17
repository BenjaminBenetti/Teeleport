package packages

import (
	"fmt"
	"os"
	"os/exec"
)

// DnfManager implements the PackageManager interface for Fedora/RHEL systems
// using the dnf command. Commands are executed via sudo so that they have the
// necessary privileges to modify the system package state.
type DnfManager struct{}

// Update refreshes the local dnf package metadata by running
// "sudo dnf check-update". An exit code of 100 from dnf indicates that
// updates are available and is treated as success, not as an error.
// It returns a non-nil error only if the command fails for a real reason.
func (d *DnfManager) Update() error {
	fmt.Printf("[teeleport] packages: running sudo dnf check-update\n")
	cmd := exec.Command("sudo", "dnf", "check-update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		// Exit code 100 means updates are available — not a real error.
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 100 {
			return nil
		}
		return err
	}
	return nil
}

// Install installs the specified system packages by running
// "sudo dnf install -y" followed by the package names.
// The packages parameter is a slice of package names to install.
// It returns an error if the command fails.
func (d *DnfManager) Install(packages []string) error {
	args := []string{"dnf", "install", "-y"}
	args = append(args, packages...)
	fmt.Printf("[teeleport] packages: running sudo %v\n", args)
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
