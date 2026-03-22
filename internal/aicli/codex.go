package aicli

import (
	"os"
	"os/exec"
)

// Codex implements the AICli interface for OpenAI's Codex CLI.
// It wraps the "codex" command-line tool, which is installed globally via npm.
type Codex struct{}

// Install installs Codex globally via npm by running
// "npm install -g @openai/codex". Stdout and stderr from the npm process are
// forwarded to the current process. It returns an error if the npm install
// command fails.
func (c *Codex) Install() error {
	cmd := buildCommand("npm", "install", "-g", "@openai/codex")
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
