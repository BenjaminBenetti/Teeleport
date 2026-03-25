package aicli

import (
	"os"
	"os/exec"
)

// Codex implements the AICli interface for OpenAI's Codex CLI.
// It wraps the "codex" command-line tool, which is installed globally via npm.
type Codex struct{}

// Install installs Codex globally via npm. It uses npmInstallGlobal so that
// sudo is only added when the npm prefix is not writable, preserving nvm's
// platform detection for optional dependencies like @openai/codex-linux-x64.
func (c *Codex) Install() error {
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
