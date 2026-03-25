package aicli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// Codex implements the AICli interface for OpenAI's Codex CLI.
// It wraps the "codex" command-line tool, which is installed globally via npm.
type Codex struct{}

// Install installs Codex globally via npm. It explicitly includes the
// platform-specific native package (e.g. @openai/codex-linux-x64) because npm
// sometimes fails to resolve optional platform dependencies when running under
// sudo in devcontainer environments.
func (c *Codex) Install() error {
	args := []string{"npm", "install", "-g", "@openai/codex"}
	if pkg := nativeCodexPackage(); pkg != "" {
		args = append(args, pkg)
	}
	cmd := buildCommand(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// nativeCodexPackage returns the platform-specific npm package name for the
// current OS and architecture (e.g. "@openai/codex-linux-x64"), or an empty
// string if the platform is unrecognized.
func nativeCodexPackage() string {
	var osName string
	switch runtime.GOOS {
	case "linux":
		osName = "linux"
	case "darwin":
		osName = "darwin"
	case "windows":
		osName = "win32"
	default:
		return ""
	}

	var archName string
	switch runtime.GOARCH {
	case "amd64":
		archName = "x64"
	case "arm64":
		archName = "arm64"
	default:
		return ""
	}

	return fmt.Sprintf("@openai/codex-%s-%s", osName, archName)
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
