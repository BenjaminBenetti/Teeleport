package aicli

import (
	"fmt"
	"os"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
)

// AICli is the interface that every AI CLI backend must implement.
// Each implementation wraps a specific AI coding assistant CLI tool
// (e.g., Claude Code, Codex, Gemini CLI, or GitHub Copilot).
type AICli interface {
	// Install installs the underlying AI CLI tool onto the system.
	// It returns an error if the installation process fails.
	Install() error
	// Run executes the AI CLI tool with the given prompt.
	// The prompt parameter is the startup instruction passed to the tool.
	// It returns an error if the tool invocation fails.
	Run(prompt string) error
}

// NewAICli returns the AICli implementation corresponding to the given tool name.
// The tool parameter must be one of: "claude-code", "codex", "gemini-cli", or "copilot".
// It returns the matching AICli implementation and a nil error on success, or a nil
// AICli and a non-nil error if the tool name is not recognized.
func NewAICli(tool string) (AICli, error) {
	switch tool {
	case "claude-code":
		return &ClaudeCode{}, nil
	case "codex":
		return &Codex{}, nil
	case "gemini-cli":
		return &GeminiCli{}, nil
	case "copilot":
		return &Copilot{}, nil
	default:
		return nil, fmt.Errorf("unknown ai-cli tool: %q", tool)
	}
}

// RunAICli is the top-level entry point for the ai-cli subsystem, typically
// called from main.go. It resolves the configured backend, installs it if
// necessary, builds the startup prompt, and runs the tool.
//
// The cfg parameter supplies the AI CLI configuration (tool name, prompt, and
// prompt file path). The dotfileRepo parameter is the base path used to resolve
// relative prompt file references via config.ResolvePath.
//
// RunAICli always returns nil. Any errors encountered during installation or
// execution are logged as warnings to stdout rather than propagated, so the
// caller never needs to treat them as fatal.
func RunAICli(cfg config.AICLIConfig, dotfileRepo string) error {
	if cfg.Tool == "" {
		fmt.Println("[teeleport] ai-cli: no tool configured, skipping")
		return nil
	}

	backend, err := NewAICli(cfg.Tool)
	if err != nil {
		fmt.Printf("[teeleport] ai-cli: warning: %v\n", err)
		return nil
	}

	if err := backend.Install(); err != nil {
		fmt.Printf("[teeleport] ai-cli: warning: install failed: %v\n", err)
		return nil
	}

	// Resolve the prompt.
	var prompt string
	switch {
	case cfg.StartupPrompt != "":
		prompt = cfg.StartupPrompt
	case cfg.StartupPromptFile != "":
		path := config.ResolvePath(dotfileRepo, cfg.StartupPromptFile)
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("[teeleport] ai-cli: warning: reading prompt file: %v\n", err)
			return nil
		}
		prompt = string(data)
	default:
		fmt.Println("[teeleport] ai-cli: no startup prompt configured, skipping invocation")
		return nil
	}

	if err := backend.Run(prompt); err != nil {
		fmt.Printf("[teeleport] ai-cli: warning: run failed: %v\n", err)
		return nil
	}

	return nil
}
