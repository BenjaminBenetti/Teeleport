package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_FullyPopulated(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "teeleport.config")

	yaml := `dotfile_repo: https://github.com/user/dotfiles.git
mounts:
  ssh:
    host: myhost.example.com
    user: alice
    port: 2222
    identity_file: ~/.ssh/id_ed25519
  permissions:
    uid: 5000
    gid: 5001
  entries:
    - name: projects
      source: /home/alice/projects
      target: /workspaces/projects
      backend: sshfs
copies:
  - name: gitconfig
    source: .gitconfig
    target: ~/.gitconfig
    mode: "0644"
packages:
  - git
  - curl
ai_cli:
  - tool: claude
    startup_prompt: "Hello!"
`
	if err := os.WriteFile(cfgPath, []byte(yaml), 0o644); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.DotfileRepo != "https://github.com/user/dotfiles.git" {
		t.Errorf("DotfileRepo = %q, want %q", cfg.DotfileRepo, "https://github.com/user/dotfiles.git")
	}

	// SSH
	if cfg.Mounts.SSH.Host != "myhost.example.com" {
		t.Errorf("SSH.Host = %q, want %q", cfg.Mounts.SSH.Host, "myhost.example.com")
	}
	if cfg.Mounts.SSH.User != "alice" {
		t.Errorf("SSH.User = %q, want %q", cfg.Mounts.SSH.User, "alice")
	}
	if cfg.Mounts.SSH.Port != 2222 {
		t.Errorf("SSH.Port = %d, want %d", cfg.Mounts.SSH.Port, 2222)
	}
	if cfg.Mounts.SSH.IdentityFile != "~/.ssh/id_ed25519" {
		t.Errorf("SSH.IdentityFile = %q, want %q", cfg.Mounts.SSH.IdentityFile, "~/.ssh/id_ed25519")
	}

	// Permissions
	if cfg.Mounts.Permissions.UID == nil || *cfg.Mounts.Permissions.UID != 5000 {
		t.Errorf("Permissions.UID = %v, want %d", cfg.Mounts.Permissions.UID, 5000)
	}
	if cfg.Mounts.Permissions.GID == nil || *cfg.Mounts.Permissions.GID != 5001 {
		t.Errorf("Permissions.GID = %v, want %d", cfg.Mounts.Permissions.GID, 5001)
	}

	// Mount entries
	if len(cfg.Mounts.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(cfg.Mounts.Entries))
	}
	e := cfg.Mounts.Entries[0]
	if e.Name != "projects" {
		t.Errorf("Entry.Name = %q, want %q", e.Name, "projects")
	}
	if e.Source != "/home/alice/projects" {
		t.Errorf("Entry.Source = %q, want %q", e.Source, "/home/alice/projects")
	}
	if e.Target != "/workspaces/projects" {
		t.Errorf("Entry.Target = %q, want %q", e.Target, "/workspaces/projects")
	}
	if e.Backend != "sshfs" {
		t.Errorf("Entry.Backend = %q, want %q", e.Backend, "sshfs")
	}

	// Copies
	if len(cfg.Copies) != 1 {
		t.Fatalf("len(Copies) = %d, want 1", len(cfg.Copies))
	}
	c := cfg.Copies[0]
	if c.Name != "gitconfig" {
		t.Errorf("Copy.Name = %q, want %q", c.Name, "gitconfig")
	}
	if c.Source != ".gitconfig" {
		t.Errorf("Copy.Source = %q, want %q", c.Source, ".gitconfig")
	}
	if c.Target != "~/.gitconfig" {
		t.Errorf("Copy.Target = %q, want %q", c.Target, "~/.gitconfig")
	}
	if c.Mode != "0644" {
		t.Errorf("Copy.Mode = %q, want %q", c.Mode, "0644")
	}

	// Packages
	if len(cfg.Packages) != 2 {
		t.Fatalf("len(Packages) = %d, want 2", len(cfg.Packages))
	}
	if cfg.Packages[0] != "git" || cfg.Packages[1] != "curl" {
		t.Errorf("Packages = %v, want [git curl]", cfg.Packages)
	}

	// AI CLI
	if len(cfg.AICli) != 1 {
		t.Fatalf("len(AICli) = %d, want 1", len(cfg.AICli))
	}
	if cfg.AICli[0].Tool != "claude" {
		t.Errorf("AICli[0].Tool = %q, want %q", cfg.AICli[0].Tool, "claude")
	}
	if cfg.AICli[0].StartupPrompt != "Hello!" {
		t.Errorf("AICli[0].StartupPrompt = %q, want %q", cfg.AICli[0].StartupPrompt, "Hello!")
	}
	if cfg.AICli[0].StartupPromptFile != "" {
		t.Errorf("AICli[0].StartupPromptFile = %q, want empty", cfg.AICli[0].StartupPromptFile)
	}
}

func TestLoadConfig_MinimalConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "teeleport.config")

	yaml := `dotfile_repo: https://github.com/user/dotfiles.git
`
	if err := os.WriteFile(cfgPath, []byte(yaml), 0o644); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.DotfileRepo != "https://github.com/user/dotfiles.git" {
		t.Errorf("DotfileRepo = %q, want %q", cfg.DotfileRepo, "https://github.com/user/dotfiles.git")
	}

	// Defaults should be applied
	if cfg.Mounts.SSH.Port != 22 {
		t.Errorf("SSH.Port = %d, want default 22", cfg.Mounts.SSH.Port)
	}
	if cfg.Mounts.Permissions.UID == nil || *cfg.Mounts.Permissions.UID != 1000 {
		t.Errorf("Permissions.UID = %v, want default 1000", cfg.Mounts.Permissions.UID)
	}
	if cfg.Mounts.Permissions.GID == nil || *cfg.Mounts.Permissions.GID != 1000 {
		t.Errorf("Permissions.GID = %v, want default 1000", cfg.Mounts.Permissions.GID)
	}
}

func TestLoadConfig_DefaultsApplied(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "teeleport.config")

	// Config with zero-value port, uid, gid — defaults should fill them in.
	yaml := `dotfile_repo: https://github.com/user/dotfiles.git
`
	if err := os.WriteFile(cfgPath, []byte(yaml), 0o644); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.Mounts.SSH.Port != 22 {
		t.Errorf("default Port = %d, want 22", cfg.Mounts.SSH.Port)
	}
	if cfg.Mounts.Permissions.UID == nil || *cfg.Mounts.Permissions.UID != 1000 {
		t.Errorf("default UID = %v, want 1000", cfg.Mounts.Permissions.UID)
	}
	if cfg.Mounts.Permissions.GID == nil || *cfg.Mounts.Permissions.GID != 1000 {
		t.Errorf("default GID = %v, want 1000", cfg.Mounts.Permissions.GID)
	}
}

func TestValidate_MutuallyExclusivePrompt(t *testing.T) {
	cfg := Config{
		AICli: []AICLIConfig{{
			StartupPrompt:     "hello",
			StartupPromptFile: "/some/file",
		}},
	}
	cfg.applyDefaults()

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate should return error when both startup_prompt and startup_prompt_file are set")
	}
	if got := err.Error(); got != "ai_cli[0]: startup_prompt and startup_prompt_file are mutually exclusive; set only one" {
		t.Errorf("unexpected error message: %q", got)
	}
}

func TestValidate_MountEntriesWithoutSSHHost(t *testing.T) {
	cfg := Config{
		Mounts: MountConfig{
			Entries: []MountEntry{
				{Name: "foo", Source: "/src", Target: "/tgt"},
			},
		},
	}
	cfg.applyDefaults()

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate should return error when mount entries are defined without SSH host")
	}
	if got := err.Error(); got != "mounts.ssh.host is required when mount entries are defined" {
		t.Errorf("unexpected error: %q", got)
	}
}

func TestValidate_MountEntriesWithoutSSHUser_OK(t *testing.T) {
	cfg := Config{
		Mounts: MountConfig{
			SSH: SSHConfig{
				Host: "example.com",
			},
			Entries: []MountEntry{
				{Name: "foo", Source: "/src", Target: "/tgt"},
			},
		},
	}
	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate should pass when SSH user is empty (runtime falls back to current user), got: %v", err)
	}
}

func TestValidate_NoMountEntries_NoSSH_OK(t *testing.T) {
	cfg := Config{
		DotfileRepo: "https://github.com/user/dotfiles.git",
	}
	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate should pass with no mount entries and no SSH config, got: %v", err)
	}
}

func TestFindConfig_ExplicitPath(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "my.config")

	if err := os.WriteFile(cfgPath, []byte("dotfile_repo: x\n"), 0o644); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}

	found, err := FindConfig(cfgPath)
	if err != nil {
		t.Fatalf("FindConfig returned error: %v", err)
	}
	if found != cfgPath {
		t.Errorf("FindConfig = %q, want %q", found, cfgPath)
	}
}

func TestFindConfig_NotFound(t *testing.T) {
	// Unset environment variable so it doesn't interfere.
	t.Setenv("TEELEPORT_CONFIG", "")

	// Use a temp dir for both cwd and HOME so default paths won't find a real config.
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpDir)
	t.Setenv("HOME", tmpDir)

	_, err := FindConfig("/nonexistent/path/to/config.yaml")
	if err == nil {
		t.Fatal("FindConfig should return error when no config is found")
	}
}

func TestLoadConfig_TypeFileParsed(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "teeleport.config")

	yaml := `mounts:
  ssh:
    host: example.com
  entries:
    - name: my-secret
      source: /home/alice/.env
      target: /workspaces/.env
      type: file
`
	if err := os.WriteFile(cfgPath, []byte(yaml), 0o644); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if len(cfg.Mounts.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(cfg.Mounts.Entries))
	}
	if cfg.Mounts.Entries[0].Type != "file" {
		t.Errorf("Entry.Type = %q, want %q", cfg.Mounts.Entries[0].Type, "file")
	}
}

func TestValidate_InvalidMountType(t *testing.T) {
	cfg := Config{
		Mounts: MountConfig{
			SSH: SSHConfig{
				Host: "example.com",
			},
			Entries: []MountEntry{
				{Name: "foo", Source: "/src", Target: "/tgt", Type: "invalid"},
			},
		},
	}
	cfg.applyDefaults()

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate should return error for invalid mount type")
	}
	want := `mounts.entries[0].type must be "directory" or "file", got "invalid"`
	if got := err.Error(); got != want {
		t.Errorf("unexpected error message:\n got: %q\nwant: %q", got, want)
	}
}

func TestValidate_EmptyMountType_OK(t *testing.T) {
	cfg := Config{
		Mounts: MountConfig{
			SSH: SSHConfig{
				Host: "example.com",
			},
			Entries: []MountEntry{
				{Name: "foo", Source: "/src", Target: "/tgt"},
			},
		},
	}
	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate should pass when mount type is empty (defaults to directory), got: %v", err)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/teeleport.config")
	if err == nil {
		t.Fatal("LoadConfig should return error for nonexistent file")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad.yaml")

	if err := os.WriteFile(cfgPath, []byte(":\n\t!!invalid\n"), 0o644); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}

	_, err := LoadConfig(cfgPath)
	if err == nil {
		t.Fatal("LoadConfig should return error for invalid YAML")
	}
}
