package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration for Teeleport.
type Config struct {
	DotfileRepo string      `yaml:"dotfile_repo"`
	Mounts      MountConfig `yaml:"mounts"`
	Copies      []CopyEntry `yaml:"copies"`
	Packages    []string    `yaml:"packages"`
	AICli       AICLIConfig `yaml:"ai_cli"`
}

// MountConfig holds SSH connection details, permission defaults, and mount entries.
type MountConfig struct {
	SSH         SSHConfig    `yaml:"ssh"`
	Permissions PermConfig   `yaml:"permissions"`
	Entries     []MountEntry `yaml:"entries"`
}

// SSHConfig describes how to connect to the remote host for SSHFS mounts.
type SSHConfig struct {
	Host         string `yaml:"host"`
	User         string `yaml:"user"`
	Port         int    `yaml:"port"`
	IdentityFile string `yaml:"identity_file"`
}

// PermConfig holds default UID/GID for mounted filesystems.
// Pointer types allow distinguishing "not set" (nil) from "explicitly set to 0" (root).
type PermConfig struct {
	UID *int `yaml:"uid"`
	GID *int `yaml:"gid"`
}

// MountEntry represents a single directory to mount via SSHFS (or another backend).
type MountEntry struct {
	Name    string `yaml:"name"`
	Source  string `yaml:"source"`
	Target  string `yaml:"target"`
	Backend string `yaml:"backend"`
}

// CopyEntry represents a file to copy from the dotfile repo into the container.
type CopyEntry struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
	Target string `yaml:"target"`
	Mode   string `yaml:"mode"`
}

// AICLIConfig controls which AI CLI tool to launch and its startup behaviour.
type AICLIConfig struct {
	Tool              string `yaml:"tool"`
	StartupPrompt     string `yaml:"startup_prompt"`
	StartupPromptFile string `yaml:"startup_prompt_file"`
}

// LoadConfig reads the YAML file at path, parses it into a Config, applies
// sensible defaults, and validates the result.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

// applyDefaults fills in zero-value fields with sensible defaults.
func (c *Config) applyDefaults() {
	if c.Mounts.SSH.Port == 0 {
		c.Mounts.SSH.Port = 22
	}
	if c.Mounts.Permissions.UID == nil {
		defaultUID := 1000
		c.Mounts.Permissions.UID = &defaultUID
	}
	if c.Mounts.Permissions.GID == nil {
		defaultGID := 1000
		c.Mounts.Permissions.GID = &defaultGID
	}
}

// Validate checks that the configuration is internally consistent.
func (c *Config) Validate() error {
	// If mount entries are defined, SSH host is required.
	if len(c.Mounts.Entries) > 0 {
		if c.Mounts.SSH.Host == "" {
			return fmt.Errorf("mounts.ssh.host is required when mount entries are defined")
		}
	}

	// Validate each mount entry.
	for i, e := range c.Mounts.Entries {
		if e.Name == "" {
			return fmt.Errorf("mounts.entries[%d].name is required", i)
		}
		if e.Source == "" {
			return fmt.Errorf("mounts.entries[%d].source is required", i)
		}
		if e.Target == "" {
			return fmt.Errorf("mounts.entries[%d].target is required", i)
		}
	}

	// Validate each copy entry.
	for i, e := range c.Copies {
		if e.Name == "" {
			return fmt.Errorf("copies[%d].name is required", i)
		}
		if e.Source == "" {
			return fmt.Errorf("copies[%d].source is required", i)
		}
		if e.Target == "" {
			return fmt.Errorf("copies[%d].target is required", i)
		}
	}

	// startup_prompt and startup_prompt_file are mutually exclusive.
	if c.AICli.StartupPrompt != "" && c.AICli.StartupPromptFile != "" {
		return fmt.Errorf("ai_cli: startup_prompt and startup_prompt_file are mutually exclusive; set only one")
	}

	return nil
}

// FindConfig locates a configuration file using the following precedence:
//  1. Explicit path from CLI flag (flagPath)
//  2. TEELEPORT_CONFIG environment variable
//  3. ~/dotfiles/teeleport.config
//  4. ~/.dotfiles/teeleport.config
//
// It returns the first path that exists, or an error if none are found.
func FindConfig(flagPath string) (string, error) {
	candidates := []string{}

	if flagPath != "" {
		candidates = append(candidates, ExpandPath(flagPath))
	}

	if envPath := os.Getenv("TEELEPORT_CONFIG"); envPath != "" {
		candidates = append(candidates, ExpandPath(envPath))
	}

	candidates = append(candidates,
		ExpandPath("~/dotfiles/teeleport.config"),
		ExpandPath("~/.dotfiles/teeleport.config"),
	)

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("no config file found; searched: %v", candidates)
}
