package aicli

import (
	"os"
	"os/exec"
)

// ClaudeCode implements the AICli interface for Anthropic's Claude Code CLI.
// It wraps the "claude" command-line tool, installed via the native installer.
type ClaudeCode struct{}

// Install installs Claude Code using the native installer by running
// "curl -fsSL https://claude.ai/install.sh | bash". This is the recommended
// installation method and does not require Node.js or npm. Stdout and stderr
// are forwarded to the current process. It returns an error if the install fails.
func (c *ClaudeCode) Install() error {
	cmd := exec.Command("bash", "-c", "curl -fsSL https://claude.ai/install.sh | bash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run executes Claude Code in print mode ("claude -p") with the given prompt.
// The prompt parameter is passed as the argument to the -p flag. Stdout and
// stderr from the claude process are forwarded to the current process.
// It returns an error if the command execution fails.
func (c *ClaudeCode) Run(prompt string) error {
	cmd := exec.Command("claude", "-p", prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
