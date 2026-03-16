package packages

import (
	"fmt"
	"os/exec"
)

// PackageManager is the interface that each distro-specific backend must implement.
type PackageManager interface {
	// Update refreshes the package index.
	Update() error
	// Install installs the given packages.
	Install(packages []string) error
}

// Detect auto-detects the system package manager by probing PATH.
// Check order: apt-get, dnf, pacman. First match wins.
func Detect() (PackageManager, error) {
	if _, err := exec.LookPath("apt-get"); err == nil {
		fmt.Printf("[teeleport] packages: detected apt-get\n")
		return &AptManager{}, nil
	}
	if _, err := exec.LookPath("dnf"); err == nil {
		fmt.Printf("[teeleport] packages: detected dnf\n")
		return &DnfManager{}, nil
	}
	if _, err := exec.LookPath("pacman"); err == nil {
		fmt.Printf("[teeleport] packages: detected pacman\n")
		return &PacmanManager{}, nil
	}
	return nil, fmt.Errorf("no supported package manager found (looked for apt-get, dnf, pacman)")
}

// Run is the convenience entry-point called from main.
// It detects the package manager, updates the index, and installs the requested packages.
func Run(packages []string) error {
	if len(packages) == 0 {
		fmt.Printf("[teeleport] packages: no packages requested, skipping\n")
		return nil
	}

	pm, err := Detect()
	if err != nil {
		return fmt.Errorf("detecting package manager: %w", err)
	}

	fmt.Printf("[teeleport] packages: updating package index\n")
	if err := pm.Update(); err != nil {
		return fmt.Errorf("updating package index: %w", err)
	}

	fmt.Printf("[teeleport] packages: installing %v\n", packages)
	if err := pm.Install(packages); err != nil {
		return fmt.Errorf("installing packages: %w", err)
	}

	fmt.Printf("[teeleport] packages: done\n")
	return nil
}
