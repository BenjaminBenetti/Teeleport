package packages

import (
	"fmt"
	"os"
	"os/exec"
)

// DnfManager implements PackageManager for Fedora/RHEL systems using dnf.
type DnfManager struct{}

// Update runs "sudo dnf check-update" to refresh the package index.
// dnf check-update exits with code 100 when updates are available, which is not an error.
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

// Install runs "sudo dnf install -y <packages...>" to install the given packages.
func (d *DnfManager) Install(packages []string) error {
	args := []string{"dnf", "install", "-y"}
	args = append(args, packages...)
	fmt.Printf("[teeleport] packages: running sudo %v\n", args)
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
