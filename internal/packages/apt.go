package packages

import (
	"fmt"
	"os"
	"os/exec"
)

// AptManager implements the PackageManager interface for Debian/Ubuntu systems
// using the apt-get command. Commands are executed via sudo so that they have
// the necessary privileges to modify the system package state.
type AptManager struct{}

// Update refreshes the local apt package index by running
// "sudo apt-get update -y". It returns an error if the command fails.
func (a *AptManager) Update() error {
	fmt.Printf("[teeleport] packages: running sudo apt-get update -y\n")
	cmd := exec.Command("sudo", "apt-get", "update", "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Install installs the specified system packages by running
// "sudo apt-get install -y" followed by the package names.
// The packages parameter is a slice of package names to install.
// It returns an error if the command fails.
func (a *AptManager) Install(packages []string) error {
	args := []string{"apt-get", "install", "-y"}
	args = append(args, packages...)
	fmt.Printf("[teeleport] packages: running sudo %v\n", args)
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
