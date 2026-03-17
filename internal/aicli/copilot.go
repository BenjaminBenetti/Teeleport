package aicli

import (
	"os"
	"os/exec"
)

// Copilot implements the AICli interface for GitHub Copilot CLI.
// It wraps the "gh copilot" command, which is installed as a GitHub CLI extension.
type Copilot struct{}

// Install installs the GitHub Copilot CLI extension by running
// "gh extension install github/gh-copilot". Stdout and stderr from the gh
// process are forwarded to the current process. It returns an error if the
// extension installation fails.
func (c *Copilot) Install() error {
	cmd := exec.Command("gh", "extension", "install", "github/gh-copilot")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run executes GitHub Copilot CLI in suggest mode ("gh copilot suggest") with
// the given prompt. The prompt parameter is passed as a positional argument to
// the suggest subcommand. Stdout and stderr from the gh process are forwarded
// to the current process. It returns an error if the command execution fails.
func (c *Copilot) Run(prompt string) error {
	cmd := exec.Command("gh", "copilot", "suggest", prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
