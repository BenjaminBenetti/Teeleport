package aicli

import (
	"os"
	"os/exec"
)

// ClaudeCode implements the AICli interface for Anthropic's Claude Code CLI.
// It wraps the "claude" command-line tool, which is installed globally via npm.
type ClaudeCode struct{}

// Install installs Claude Code globally via npm by running
// "npm install -g @anthropic-ai/claude-code". Stdout and stderr from the npm
// process are forwarded to the current process. It returns an error if the
// npm install command fails.
func (c *ClaudeCode) Install() error {
	cmd := exec.Command("npm", "install", "-g", "@anthropic-ai/claude-code")
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
