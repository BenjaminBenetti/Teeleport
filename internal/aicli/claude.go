package aicli

import (
	"os"
	"os/exec"
)

// ClaudeCode implements the AICli interface for Anthropic's Claude Code.
type ClaudeCode struct{}

// Install installs Claude Code via npm.
func (c *ClaudeCode) Install() error {
	cmd := exec.Command("npm", "install", "-g", "@anthropic-ai/claude-code")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run executes Claude Code with the given prompt.
func (c *ClaudeCode) Run(prompt string) error {
	cmd := exec.Command("claude", "-p", prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
