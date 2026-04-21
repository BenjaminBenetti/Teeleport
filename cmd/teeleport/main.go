// Package main is the entry point for the teeleport CLI. It loads the
// configuration file, then orchestrates package installation, preflight
// checks, SSHFS mounts, rsync syncs, file copies, and optional AI CLI setup.
// The "rsync" subcommand runs a long-lived daemon that periodically
// synchronises rsync-backend entries bidirectionally.
//
// Exit code 0 indicates success; exit code 1 indicates one or more errors.
//
// All output is tee'd to ~/teeleport/run.log for debugging.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/BenjaminBenetti/Teeleport/internal/aicli"
	"github.com/BenjaminBenetti/Teeleport/internal/config"
	filecopy "github.com/BenjaminBenetti/Teeleport/internal/copy"
	"github.com/BenjaminBenetti/Teeleport/internal/mount"
	"github.com/BenjaminBenetti/Teeleport/internal/packages"
	"github.com/BenjaminBenetti/Teeleport/internal/preflight"
	rsyncpkg "github.com/BenjaminBenetti/Teeleport/internal/rsync"
)

// version holds the build version string for the teeleport binary.
// It defaults to "dev" and is overridden at build time via
// -ldflags "-X main.version=<semver>".
var version = "dev"


// runRsyncDaemon handles the "teeleport rsync" subcommand. It loads the
// configuration, filters for rsync-backend entries, and starts the daemon
// sync loop. This function does not return until the daemon is stopped.
func runRsyncDaemon() int {
	fs := flag.NewFlagSet("rsync", flag.ContinueOnError)
	configFlag := fs.String("config", "", "path to teeleport config file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] rsync: %v\n", err)
		return 1
	}

	cfgPath, err := config.FindConfig(*configFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] rsync: %v\n", err)
		return 1
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] rsync: %v\n", err)
		return 1
	}

	entries := rsyncpkg.FilterRsyncEntries(cfg.Mounts.Entries)
	if len(entries) == 0 {
		fmt.Println("[teeleport] rsync: no rsync entries in config, nothing to do")
		return 0
	}

	daemonCfg := rsyncpkg.DaemonConfig{
		SSH:      cfg.Mounts.SSH,
		Entries:  entries,
		Interval: time.Duration(cfg.Mounts.Rsync.Interval) * time.Second,
	}

	if err := rsyncpkg.RunDaemon(daemonCfg); err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] rsync daemon error: %v\n", err)
		return 1
	}
	return 0
}

// startRsyncDaemonIfNeeded launches the rsync daemon in the background when
// the configuration contains rsync-backend entries and no daemon is already
// running. The daemon is fully detached (new session) so it survives the
// parent process exiting.
func startRsyncDaemonIfNeeded(cfgPath string) {
	if pid := rsyncpkg.IsDaemonRunning(); pid != 0 {
		fmt.Printf("[teeleport] rsync daemon already running (pid %d)\n", pid)
		return
	}

	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] warning: cannot find own executable for rsync daemon: %v\n", err)
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] warning: cannot determine home dir for rsync log: %v\n", err)
		return
	}

	logPath := filepath.Join(home, ".teeleport", "rsync.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[teeleport] warning: cannot open rsync log %s: %v\n", logPath, err)
		return
	}

	cmd := exec.Command(exePath, "rsync", "--config", cfgPath)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		fmt.Fprintf(os.Stderr, "[teeleport] warning: failed to start rsync daemon: %v\n", err)
		return
	}

	// Capture PID before releasing the process handle.
	daemonPid := cmd.Process.Pid
	cmd.Process.Release()
	logFile.Close()

	fmt.Printf("[teeleport] rsync daemon started (pid %d)\n", daemonPid)
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

	// --- AI CLI install ---
	// Install AI tools early (before mounts) so their installation does not
	// overwrite state files that mounts will provide.
	for _, cli := range cfg.AICli {
		if cli.Tool != "" {
			_ = aicli.InstallAICli(cli)
		}
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

	// --- AI CLI startup prompts ---
	// Run AI startup prompts after mounts, since prompts may depend on
	// auth info supplied by mounted files.
	for _, cli := range cfg.AICli {
		if cli.Tool != "" {
			_ = aicli.RunAICliPrompt(cli, config.ExpandPath(cfg.DotfileRepo))
		}
	}

	// --- Summary ---
	fmt.Printf("[teeleport] done: %d packages, %d mounts, %d copies (%d errors, %d warnings)\n",
		pkgCount, mountCount, copyCount, totalErrors, warnings)

	// --- SSH multiplexing ---
	// Enable multiplexing for any SSH-backed backend (sshfs or rsync) so that
	// ensureRemotePath calls, multiple mounts, and rsync cycles all reuse a
	// single master connection.
	if cfg.Mounts.SSH.Host != "" {
		if err := rsyncpkg.EnsureSSHMultiplexing(cfg.Mounts.SSH); err != nil {
			fmt.Fprintf(os.Stderr, "[teeleport] warning: ssh multiplexing: %v\n", err)
			warnings++
		}
	}

	// --- Rsync daemon ---
	// Skip the daemon when preflight failed: launching it against a broken
	// SSH connection can corrupt remote state if the failure is intermittent.
	if preflightOK && rsyncpkg.HasRsyncEntries(cfg.Mounts.Entries) {
		startRsyncDaemonIfNeeded(cfgPath)
	} else if !preflightOK && rsyncpkg.HasRsyncEntries(cfg.Mounts.Entries) {
		fmt.Println("[teeleport] skipping rsync daemon due to preflight failure")
	}

	if totalErrors > 0 {
		return 1
	}
	return 0
}

func main() {
	// Check for subcommands before flag.Parse() since the standard flag
	// package stops at the first non-flag argument.
	if len(os.Args) > 1 && os.Args[1] == "rsync" {
		os.Exit(runRsyncDaemon())
	}

	os.Exit(run())
}
