package aicli

import (
	"os"
	"os/exec"
)

// GeminiCli implements the AICli interface for Google's Gemini CLI.
type GeminiCli struct{}

// Install installs Gemini CLI via npm.
func (g *GeminiCli) Install() error {
	cmd := exec.Command("npm", "install", "-g", "@google/gemini-cli")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run executes Gemini CLI with the given prompt.
func (g *GeminiCli) Run(prompt string) error {
	cmd := exec.Command("gemini", "-p", prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
