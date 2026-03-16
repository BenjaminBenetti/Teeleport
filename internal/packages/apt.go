package packages

import (
	"fmt"
	"os"
	"os/exec"
)

// AptManager implements PackageManager for Debian/Ubuntu systems using apt-get.
type AptManager struct{}

// Update runs "sudo apt-get update -y" to refresh the package index.
func (a *AptManager) Update() error {
	fmt.Printf("[teeleport] packages: running sudo apt-get update -y\n")
	cmd := exec.Command("sudo", "apt-get", "update", "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Install runs "sudo apt-get install -y <packages...>" to install the given packages.
func (a *AptManager) Install(packages []string) error {
	args := []string{"apt-get", "install", "-y"}
	args = append(args, packages...)
	fmt.Printf("[teeleport] packages: running sudo %v\n", args)
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
