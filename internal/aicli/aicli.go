package aicli

import (
	"fmt"
	"os"

	"github.com/BenjaminBenetti/Teeleport/internal/config"
)

// AICli is the interface every AI CLI backend must implement.
type AICli interface {
	// Install installs the AI CLI tool.
	Install() error
	// Run executes the CLI with the given prompt.
	Run(prompt string) error
}

// NewAICli returns the AICli implementation for the given tool name.
// Supported values: "claude-code", "codex", "gemini-cli", "copilot".
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

// RunAICli is the main entry point called from main.go.
// It never returns an error that should cause the program to exit non-zero;
// all errors are logged as warnings and nil is returned.
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
