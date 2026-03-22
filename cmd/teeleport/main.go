// Package main is the entry point for the teeleport CLI. It loads the
// configuration file, then orchestrates package installation, preflight
// checks, SSHFS mounts, file copies, and optional AI CLI setup. Exit
// code 0 indicates success; exit code 1 indicates one or more errors.
//
// All output is tee'd to ~/teeleport/run.log for debugging.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/BenjaminBenetti/Teeleport/internal/aicli"
	"github.com/BenjaminBenetti/Teeleport/internal/config"
	filecopy "github.com/BenjaminBenetti/Teeleport/internal/copy"
	"github.com/BenjaminBenetti/Teeleport/internal/mount"
	"github.com/BenjaminBenetti/Teeleport/internal/packages"
	"github.com/BenjaminBenetti/Teeleport/internal/preflight"
)

// version holds the build version string for the teeleport binary.
// It defaults to "dev" and is overridden at build time via
// -ldflags "-X main.version=<semver>".
var version = "dev"

// setupLogFile creates ~/teeleport/run.log and tees all stdout/stderr to it.
// It returns a cleanup function that must be called before exiting to flush
// all buffered output. If the log file cannot be created, output goes to
// the terminal only and a warning is printed.
func setupLogFile() func() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] warning: cannot determine home directory for log file: %v\n", err)
		return func() {}
	}

	logDir := filepath.Join(home, ".teeleport")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] warning: cannot create log directory %s: %v\n", logDir, err)
		return func() {}
	}

	logPath := filepath.Join(logDir, "run.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] warning: cannot create log file %s: %v\n", logPath, err)
		return func() {}
	}

	origStdout := os.Stdout
	origStderr := os.Stderr

	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()

	os.Stdout = stdoutW
	os.Stderr = stderrW

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(io.MultiWriter(origStdout, logFile), stdoutR)
	}()

	go func() {
		defer wg.Done()
		io.Copy(io.MultiWriter(origStderr, logFile), stderrR)
	}()

	return func() {
		// Close the write ends so the copy goroutines see EOF and finish.
		stdoutW.Close()
		stderrW.Close()
		// Wait for all output to be flushed to the log file.
		wg.Wait()
		logFile.Close()
	}
}

// run contains the main application logic and returns the desired exit code.
func run() int {
	configFlag := flag.String("config", "", "path to teeleport config file")
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println("teeleport", version)
		return 0
	}

	// Locate the configuration file.
	cfgPath, err := config.FindConfig(*configFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] error: %v\n", err)
		return 1
	}

	fmt.Printf("[teeleport] loading config from %s\n", cfgPath)

	// Load and validate the configuration.
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] error: %v\n", err)
		return 1
	}

	var totalErrors int
	var warnings int

	// --- Packages ---
	pkgCount := len(cfg.Packages)
	if err := packages.Run(cfg.Packages); err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] warning: packages: %v\n", err)
		warnings++
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
	for _, cli := range cfg.AICli {
		if cli.Tool != "" {
			_ = aicli.RunAICli(cli, config.ExpandPath(cfg.DotfileRepo))
		}
	}

	// --- Summary ---
	fmt.Printf("[teeleport] done: %d packages, %d mounts, %d copies (%d errors, %d warnings)\n",
		pkgCount, mountCount, copyCount, totalErrors, warnings)

	if totalErrors > 0 {
		return 1
	}
	return 0
}

func main() {
	cleanup := setupLogFile()
	exitCode := run()
	cleanup()
	os.Exit(exitCode)
}
