package packages

import (
	"fmt"
	"os"
	"os/exec"
)

// PacmanManager implements the PackageManager interface for Arch Linux systems
// using the pacman command. Commands are executed via sudo so that they have
// the necessary privileges to modify the system package state.
type PacmanManager struct{}

// Update refreshes the local pacman package database by running
// "sudo pacman -Sy --noconfirm". It returns an error if the command fails.
func (p *PacmanManager) Update() error {
	fmt.Printf("[teeleport] packages: running sudo pacman -Sy --noconfirm\n")
	cmd := exec.Command("sudo", "pacman", "-Sy", "--noconfirm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Install installs the specified system packages by running
// "sudo pacman -S --noconfirm" followed by the package names.
// The packages parameter is a slice of package names to install.
// It returns an error if the command fails.
func (p *PacmanManager) Install(packages []string) error {
	args := []string{"pacman", "-S", "--noconfirm"}
	args = append(args, packages...)
	fmt.Printf("[teeleport] packages: running sudo %v\n", args)
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
