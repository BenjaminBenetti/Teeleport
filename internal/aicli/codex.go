package aicli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Codex implements the AICli interface for OpenAI's Codex CLI.
// It wraps the "codex" command-line tool, which is installed globally via npm.
type Codex struct{}

// Install installs Codex globally via npm. It uses npmInstallGlobal so that
// sudo is only added when the npm prefix is not writable, preserving nvm's
// platform detection for optional dependencies like @openai/codex-linux-x64.
func (c *Codex) Install() error {
	if err := ensureCodexConfig(); err != nil {
		fmt.Printf("[teeleport] codex: warning: could not configure sandbox fallback: %v\n", err)
	}
	cmd := npmInstallGlobal("@openai/codex")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run executes Codex with the given prompt. The prompt parameter is passed as
// a positional argument to the "codex" command. Stdout and stderr from the
// codex process are forwarded to the current process. It returns an error if
// the command execution fails.
func (c *Codex) Run(prompt string) error {
	cmd := exec.Command("codex", prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// bwrapSupportsArgv0 checks whether the system bubblewrap supports --argv0.
// Returns true if bwrap is present and its help output contains --argv0.
func bwrapSupportsArgv0() bool {
	out, err := exec.Command("bwrap", "--help").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "--argv0")
}

// ensureCodexConfig writes a Codex config.toml that enables legacy Landlock
// sandboxing when the system bubblewrap is too old (lacks --argv0 support).
// It is idempotent: if the setting already exists, the file is left unchanged.
func ensureCodexConfig() error {
	if bwrapSupportsArgv0() {
		return nil
	}

	fmt.Println("[teeleport] codex: system bwrap lacks --argv0 support, enabling legacy landlock sandbox")

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determining home directory: %w", err)
	}

	dir := filepath.Join(home, ".codex")
	configPath := filepath.Join(dir, "config.toml")

	const marker = "use_legacy_landlock"
	const block = "\n# Written by Teeleport to work around old system bubblewrap.\n[features]\nuse_legacy_landlock = true\n"

	data, err := os.ReadFile(configPath)
	if err == nil {
		if strings.Contains(string(data), marker) {
			return nil
		}
		f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("opening config for append: %w", err)
		}
		defer f.Close()
		_, err = f.WriteString(block)
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating .codex directory: %w", err)
	}
	return os.WriteFile(configPath, []byte(block[1:]), 0o644)
}
