package aicli

import (
	"os"
	"os/exec"
)

// Copilot implements the AICli interface for GitHub Copilot CLI.
type Copilot struct{}

// Install installs the GitHub Copilot CLI extension via gh.
func (c *Copilot) Install() error {
	cmd := exec.Command("gh", "extension", "install", "github/gh-copilot")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run executes GitHub Copilot CLI with the given prompt.
func (c *Copilot) Run(prompt string) error {
	cmd := exec.Command("gh", "copilot", "suggest", prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
