package config

import (
	"fmt"
	"os"

	"github.com/BenjaminBenetti/Teeleport/internal/config/mountpresets"
	"github.com/BenjaminBenetti/Teeleport/internal/domainmodel"
	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration for Teeleport. It aggregates every
// section of the YAML configuration file: the dotfile repository location,
// SSHFS mount definitions, file-copy rules, OS package lists, and AI CLI
// preferences.
//
// Fields:
//   - DotfileRepo: Path to the dotfiles repository. Defaults to "." (cwd).
//   - Mounts:      SSH-based filesystem mount configuration.
//   - Copies:      Files to copy from the dotfile repo into the container.
//   - Packages:    OS packages to install during setup.
//   - AICli:       List of AI CLI tools and their startup behaviour.
type Config struct {
	DotfileRepo string                    `yaml:"dotfile_repo"`
	Mounts      domainmodel.MountConfig   `yaml:"mounts"`
	Copies      []domainmodel.CopyEntry   `yaml:"copies"`
	Packages    []string                  `yaml:"packages"`
	AICli       []domainmodel.AICLIConfig `yaml:"ai_cli"`
}

// LoadConfig reads the YAML configuration file at path, unmarshals it into a
// Config struct, applies sensible defaults for any unset fields (see
// applyDefaults), expands mount presets, and validates the result with
// Config.Validate.
//
// Parameters:
//   - path: absolute or relative filesystem path to the YAML config file.
//
// It returns the fully populated Config or an error if the file cannot be read,
// contains invalid YAML, or fails validation.
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

	if err := cfg.expandPresets(); err != nil {
		return nil, fmt.Errorf("expanding presets: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

// applyDefaults fills in zero-value fields with sensible defaults.
func (c *Config) applyDefaults() {
	if c.DotfileRepo == "" {
		c.DotfileRepo = "."
	}
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

// expandPresets replaces mount entries that reference a preset with the
// concrete mount entries defined by that preset.
func (c *Config) expandPresets() error {
	var expanded []domainmodel.MountEntry
	for _, entry := range c.Mounts.Entries {
		if entry.Preset != "" {
			entries, err := mountpresets.Get(entry.Preset)
			if err != nil {
				return err
			}
			// Apply user-defined overrides from the preset entry onto each expanded entry
			for _, e := range entries {
				if entry.Backend != "" {
					e.Backend = entry.Backend
				}
				if entry.ForceMount {
					e.ForceMount = true
				}
				if entry.File.DefaultContent != "" {
					e.File.DefaultContent = entry.File.DefaultContent
				}
				expanded = append(expanded, e)
			}
		} else {
			expanded = append(expanded, entry)
		}
	}
	c.Mounts.Entries = expanded
	return nil
}

// Validate checks that the Config is internally consistent and that all
// required fields are present. Specifically it enforces that:
//   - mounts.ssh.host is set when any mount entries exist,
//   - every MountEntry has a non-empty Name, Source, and Target,
//   - every CopyEntry has a non-empty Name, Source, and Target,
//   - StartupPrompt and StartupPromptFile are not both set in AICLIConfig.
//
// It returns nil when the configuration is valid, or a descriptive error
// identifying the first rule that was violated.
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
		if e.Type != "" && e.Type != "directory" && e.Type != "file" {
			return fmt.Errorf("mounts.entries[%d].type must be \"directory\" or \"file\", got %q", i, e.Type)
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
	for i, cli := range c.AICli {
		if cli.StartupPrompt != "" && cli.StartupPromptFile != "" {
			return fmt.Errorf("ai_cli[%d]: startup_prompt and startup_prompt_file are mutually exclusive; set only one", i)
		}
	}

	return nil
}

// FindConfig locates a Teeleport configuration file by probing a series of
// candidate paths in the following precedence order:
//  1. The explicit path supplied via the --config CLI flag (flagPath).
//  2. The path in the TEELEPORT_CONFIG environment variable.
//  3. ./teeleport.config (current working directory)
//  4. ~/dotfiles/teeleport.config
//  5. ~/.dotfiles/teeleport.config
//
// Each candidate is expanded with ExpandPath before being checked.
//
// Parameters:
//   - flagPath: optional config path from a CLI flag; pass "" to skip.
//
// It returns the absolute path of the first candidate that exists on disk, or
// an error listing all searched paths if none are found.
func FindConfig(flagPath string) (string, error) {
	candidates := []string{}

	if flagPath != "" {
		candidates = append(candidates, ExpandPath(flagPath))
	}

	if envPath := os.Getenv("TEELEPORT_CONFIG"); envPath != "" {
		candidates = append(candidates, ExpandPath(envPath))
	}

	candidates = append(candidates,
		"teeleport.config",
		"teeleport.config.yaml",
		ExpandPath("~/dotfiles/teeleport.config"),
		ExpandPath("~/dotfiles/teeleport.config.yaml"),
		ExpandPath("~/.dotfiles/teeleport.config"),
		ExpandPath("~/.dotfiles/teeleport.config.yaml"),
	)

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("no config file found; searched: %v", candidates)
}
