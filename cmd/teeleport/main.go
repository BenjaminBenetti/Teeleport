package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/BenjaminBenetti/Teeleport/internal/aicli"
	"github.com/BenjaminBenetti/Teeleport/internal/config"
	filecopy "github.com/BenjaminBenetti/Teeleport/internal/copy"
	"github.com/BenjaminBenetti/Teeleport/internal/mount"
	"github.com/BenjaminBenetti/Teeleport/internal/packages"
	"github.com/BenjaminBenetti/Teeleport/internal/preflight"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	configFlag := flag.String("config", "", "path to teeleport config file")
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println("teeleport", version)
		os.Exit(0)
	}

	// Locate the configuration file.
	cfgPath, err := config.FindConfig(*configFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[teeleport] loading config from %s\n", cfgPath)

	// Load and validate the configuration.
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] error: %v\n", err)
		os.Exit(1)
	}

	var totalErrors int
	var warnings int

	// --- Packages ---
	pkgCount := len(cfg.Packages)
	if err := packages.Run(cfg.Packages); err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] warning: packages: %v\n", err)
		warnings++
		// Continue despite package errors.
	}

	// --- Preflight checks ---
	preflightOK := true
	if err := preflight.RunChecks(cfg); err != nil {
		if len(cfg.Mounts.Entries) > 0 {
			fmt.Fprintf(os.Stderr, "[teeleport] error: preflight: %v\n", err)
			preflightOK = false
		}
	}

	// --- Mounts ---
	mountCount := len(cfg.Mounts.Entries)
	if preflightOK && mountCount > 0 {
		if err := mount.ProcessMounts(cfg.Mounts); err != nil {
			totalErrors++
		}
	} else if !preflightOK && mountCount > 0 {
		fmt.Println("[teeleport] skipping mounts due to preflight failure")
		totalErrors++
	}

	// --- Copies ---
	copyCount := len(cfg.Copies)
	if err := filecopy.ProcessCopies(config.ExpandPath(cfg.DotfileRepo), cfg.Copies); err != nil {
		totalErrors++
	}

	// --- AI CLI ---
	if cfg.AICli.Tool != "" {
		// AI CLI errors are never fatal.
		_ = aicli.RunAICli(cfg.AICli, config.ExpandPath(cfg.DotfileRepo))
	}

	// --- Summary ---
	fmt.Printf("[teeleport] done: %d packages, %d mounts, %d copies (%d errors, %d warnings)\n",
		pkgCount, mountCount, copyCount, totalErrors, warnings)

	if totalErrors > 0 {
		os.Exit(1)
	}
}
