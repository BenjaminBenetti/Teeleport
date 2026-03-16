package aicli

import (
	"os"
	"os/exec"
)

// Codex implements the AICli interface for OpenAI's Codex CLI.
type Codex struct{}

// Install installs Codex via npm.
func (c *Codex) Install() error {
	cmd := exec.Command("npm", "install", "-g", "@openai/codex")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run executes Codex with the given prompt.
func (c *Codex) Run(prompt string) error {
	cmd := exec.Command("codex", prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
