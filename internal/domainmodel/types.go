// Package domainmodel defines shared types used across Teeleport packages.
// These types are extracted here to avoid circular imports between config,
// mount, copy, aicli, and mountpresets packages.
package domainmodel

// MountConfig holds SSH connection details, permission defaults, and the list
// of directories to mount via SSHFS (or another backend).
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

// PermConfig holds the default UID and GID applied to mounted filesystems.
// Pointer types distinguish "not set" (nil → default 1000) from "set to 0" (root).
type PermConfig struct {
	UID *int `yaml:"uid"`
	GID *int `yaml:"gid"`
}

// FileConfig holds file-specific mount options.
type FileConfig struct {
	DefaultContent string `yaml:"default_content"`
}

// MountEntry represents a single remote directory or file to mount, or a
// preset that expands into multiple mount entries.
type MountEntry struct {
	Name    string     `yaml:"name"`
	Source  string     `yaml:"source"`
	Target  string     `yaml:"target"`
	Backend string     `yaml:"backend"`
	Type    string     `yaml:"type"`   // "directory" (default) or "file"
	Preset  string     `yaml:"preset"` // If set, expands to a predefined set of mounts
	File    FileConfig `yaml:"file"`   // File-specific options (only used when Type == "file")
}

// CopyEntry represents a single file to copy from the dotfile repository into
// the container during setup.
type CopyEntry struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
	Target string `yaml:"target"`
	Mode   string `yaml:"mode"`
}

// AICLIConfig controls which AI CLI tool to launch and how it should be
// initialised at startup. StartupPrompt and StartupPromptFile are mutually
// exclusive.
type AICLIConfig struct {
	Tool              string `yaml:"tool"`
	StartupPrompt     string `yaml:"startup_prompt"`
	StartupPromptFile string `yaml:"startup_prompt_file"`
}
