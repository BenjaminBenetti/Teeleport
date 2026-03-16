package packages

import (
	"fmt"
	"os"
	"os/exec"
)

// PacmanManager implements PackageManager for Arch Linux systems using pacman.
type PacmanManager struct{}

// Update runs "sudo pacman -Sy --noconfirm" to refresh the package index.
func (p *PacmanManager) Update() error {
	fmt.Printf("[teeleport] packages: running sudo pacman -Sy --noconfirm\n")
	cmd := exec.Command("sudo", "pacman", "-Sy", "--noconfirm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Install runs "sudo pacman -S --noconfirm <packages...>" to install the given packages.
func (p *PacmanManager) Install(packages []string) error {
	args := []string{"pacman", "-S", "--noconfirm"}
	args = append(args, packages...)
	fmt.Printf("[teeleport] packages: running sudo %v\n", args)
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
