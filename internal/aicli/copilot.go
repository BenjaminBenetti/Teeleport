package aicli

import (
	"os"
	"os/exec"
)

// Copilot implements the AICli interface for GitHub Copilot CLI.
// It wraps the "copilot" command, installed globally via npm.
type Copilot struct{}

// Install installs GitHub Copilot CLI globally via npm by running
// "npm install -g @github/copilot". Stdout and stderr from the npm
// process are forwarded to the current process. It returns an error if the
// installation fails.
func (c *Copilot) Install() error {
	cmd := exec.Command("npm", "install", "-g", "@github/copilot")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run executes GitHub Copilot CLI with the given prompt. The prompt parameter
// is passed as a positional argument. Stdout and stderr from the copilot
// process are forwarded to the current process. It returns an error if the
// command execution fails.
func (c *Copilot) Run(prompt string) error {
	cmd := exec.Command("copilot", prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
