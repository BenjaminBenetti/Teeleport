package aicli

import (
	"os"
	"os/exec"
)

// GeminiCli implements the AICli interface for Google's Gemini CLI.
// It wraps the "gemini" command-line tool, which is installed globally via npm.
type GeminiCli struct{}

// Install installs Gemini CLI globally via npm by running
// "npm install -g @google/gemini-cli". Stdout and stderr from the npm process
// are forwarded to the current process. It returns an error if the npm install
// command fails.
func (g *GeminiCli) Install() error {
	cmd := npmInstallGlobal("@google/gemini-cli")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run executes Gemini CLI in print mode ("gemini -p") with the given prompt.
// The prompt parameter is passed as the argument to the -p flag. Stdout and
// stderr from the gemini process are forwarded to the current process.
// It returns an error if the command execution fails.
func (g *GeminiCli) Run(prompt string) error {
	cmd := exec.Command("gemini", "-p", prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
