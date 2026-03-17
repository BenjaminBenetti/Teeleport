package packages

import (
	"fmt"
	"os/exec"
)

// PackageManager is the interface that each distro-specific backend must implement.
// Implementations exist for apt-get (Debian/Ubuntu), dnf (Fedora/RHEL), and
// pacman (Arch Linux).
type PackageManager interface {
	// Update refreshes the local package index from the upstream repositories.
	// It returns an error if the underlying system command fails.
	Update() error
	// Install installs one or more system packages by name.
	// The packages parameter is a slice of package names to install.
	// It returns an error if the underlying system command fails.
	Install(packages []string) error
}

// Detect auto-detects the system package manager by probing the PATH for
// known package-manager binaries. The check order is: apt-get, dnf, pacman.
// The first match wins.
//
// Detect returns a PackageManager implementation corresponding to the detected
// backend, or a non-nil error if no supported package manager is found.
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

// Run is the convenience entry-point called from main. It detects the system
// package manager via Detect, refreshes the package index, and installs the
// requested packages.
//
// The packages parameter is a slice of package names to install. If packages is
// empty, Run returns nil immediately without performing any work.
// Run returns a non-nil error if detection, index update, or installation fails.
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
