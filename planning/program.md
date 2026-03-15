# Teeleport — Go Program Plan

## Project Structure

```
teeleport/
├── cmd/
│   └── teeleport/
│       └── main.go              # Entry point — load config, run operations
├── internal/
│   ├── config/
│   │   ├── config.go            # Config structs and YAML parsing
│   │   └── paths.go             # Path resolution (~, relative paths)
│   ├── mount/
│   │   ├── mount.go             # MountBackend interface (strategy pattern)
│   │   ├── sshfs.go             # SSHFS backend implementation
│   │   └── manager.go           # Processes mount entries, checks existing mounts
│   ├── copy/
│   │   ├── copy.go              # Copy logic — replace and append modes
│   │   └── sentinel.go          # Sentinel marker parsing for idempotent appends
│   ├── packages/
│   │   ├── packages.go          # PackageManager interface (strategy pattern)
│   │   ├── apt.go               # apt backend (Ubuntu/Debian)
│   │   ├── dnf.go               # dnf backend (RHEL/Fedora)
│   │   └── pacman.go            # pacman backend (Arch)
│   ├── aicli/
│   │   ├── aicli.go             # AICli interface (strategy pattern)
│   │   ├── claude.go            # Claude Code backend
│   │   ├── codex.go             # OpenAI Codex CLI backend
│   │   ├── gemini.go            # Gemini CLI backend
│   │   └── copilot.go           # GitHub Copilot CLI backend
│   └── preflight/
│       └── preflight.go         # Pre-run checks (FUSE available, SSH connectivity)
├── teeleport.config              # Example config (not shipped in binary)
├── go.mod
└── go.sum
```

## Dependencies

- `gopkg.in/yaml.v3` — YAML config parsing
- Standard library only for everything else (os/exec for sshfs, os for file ops)

No framework needed. This is a short-lived CLI that runs and exits — not a long-running service.

## Execution Flow

```
main()
  ├── 1. Locate config file
  │     └── Look for teeleport.config in known dotfile locations
  │         (configurable via CLI flag or env var as fallback)
  │
  ├── 2. Parse config
  │     └── Deserialize YAML into Config struct
  │     └── Resolve all paths (expand ~)
  │     └── Validate required fields
  │
  ├── 3. Install packages
  │     ├── If packages defined: detect package manager (apt/dnf/pacman)
  │     ├── Run update (apt-get update, dnf check-update, pacman -Sy)
  │     └── Install all packages in a single command
  │
  ├── 4. Preflight checks
  │     ├── If mounts defined: check FUSE availability (/dev/fuse exists)
  │     ├── If mounts defined: check sshfs binary available (install if missing)
  │     └── If mounts defined: verify SSH connectivity to host
  │
  ├── 5. Process mounts
  │     └── For each mount entry:
  │           ├── Check if target is already mounted (skip if so)
  │           ├── Create target directory if it doesn't exist
  │           └── Execute mount via backend (sshfs)
  │
  ├── 6. Process copies
  │     └── For each copy entry:
  │           ├── Resolve source path (dotfile_repo + source)
  │           ├── If mode=replace: write source content to target
  │           └── If mode=append: insert/update sentinel block in target
  │
  ├── 7. AI CLI
  │     ├── If ai_cli defined: install the tool via its backend
  │     ├── Resolve startup prompt (inline string or read from file)
  │     └── If prompt provided: invoke the CLI with the prompt
  │     └── All errors are warnings only — never fatal
  │
  └── 8. Exit
        └── Log summary of what was done (X packages, Y mounts, Z copies)
        └── Exit 0 on success, non-zero on failure (AI CLI errors excluded)
```

## Key Types

### Config Structs

```go
type Config struct {
    DotfileRepo string      `yaml:"dotfile_repo"`
    Mounts      MountConfig `yaml:"mounts"`
    Copies      []CopyEntry `yaml:"copies"`
    Packages    []string    `yaml:"packages"`
    AICli       AICLIConfig `yaml:"ai_cli"`
}

type AICLIConfig struct {
    Tool              string `yaml:"tool"`
    StartupPrompt     string `yaml:"startup_prompt"`
    StartupPromptFile string `yaml:"startup_prompt_file"`
}

type MountConfig struct {
    SSH         SSHConfig      `yaml:"ssh"`
    Permissions PermConfig     `yaml:"permissions"`
    Entries     []MountEntry   `yaml:"entries"`
}

type SSHConfig struct {
    Host         string `yaml:"host"`
    User         string `yaml:"user"`
    Port         int    `yaml:"port"`
    IdentityFile string `yaml:"identity_file"`
}

type PermConfig struct {
    UID int `yaml:"uid"`
    GID int `yaml:"gid"`
}

type MountEntry struct {
    Name    string `yaml:"name"`
    Source  string `yaml:"source"`
    Target  string `yaml:"target"`
    Backend string `yaml:"backend"`
}

type CopyEntry struct {
    Name   string `yaml:"name"`
    Source string `yaml:"source"`
    Target string `yaml:"target"`
    Mode   string `yaml:"mode"`
}
```

### Mount Backend Interface

```go
// MountBackend is the strategy interface for mount operations.
type MountBackend interface {
    // Mount mounts source to target. Returns an error if the mount fails.
    Mount(source, target string) error

    // IsMounted checks if the target path is already mounted.
    IsMounted(target string) (bool, error)
}
```

The SSHFS implementation constructs and execs the `sshfs` command using the SSH config and permissions from `MountConfig`.

### Package Manager Interface

```go
// PackageManager is the strategy interface for package installation.
type PackageManager interface {
    // Update refreshes the package index.
    Update() error

    // Install installs the given packages. Skips already-installed packages.
    Install(packages []string) error
}
```

Auto-detection checks which binary exists in PATH:

| Binary   | Backend |
|----------|---------|
| `apt-get`| AptManager  |
| `dnf`    | DnfManager  |
| `pacman` | PacmanManager |

First match wins. If none found, log a warning and skip package installation.

### Package Manager Commands

| Backend | Update               | Install                              |
|---------|----------------------|--------------------------------------|
| apt     | `apt-get update -y`  | `apt-get install -y <packages...>`   |
| dnf     | `dnf check-update`   | `dnf install -y <packages...>`       |
| pacman  | `pacman -Sy`         | `pacman -S --noconfirm <packages...>`|

All packages are installed in a single command to minimize overhead.

### AI CLI Interface

```go
// AICli is the strategy interface for AI CLI tool operations.
type AICli interface {
    // Install installs the AI CLI tool.
    Install() error

    // Run executes the CLI with the given prompt. Returns error on failure.
    Run(prompt string) error
}
```

Selection is based on the `tool` config value:

| `tool` value | Backend |
|---|---|
| `claude-code` | ClaudeCode |
| `codex` | Codex |
| `gemini-cli` | GeminiCli |
| `copilot` | Copilot |

AI CLI errors are logged as warnings and **never** cause Teeleport to exit non-zero.


## SSHFS Command Construction

The SSHFS backend builds a command like:

```bash
sshfs user@host:/remote/path /local/path \
  -o port=22 \
  -o uid=1000 \
  -o gid=1000 \
  -o IdentityFile=/path/to/key \   # only if configured
  -o StrictHostKeyChecking=no \
  -o reconnect \
  -o ServerAliveInterval=15
```

- `reconnect` and `ServerAliveInterval` keep the mount alive if the connection drops briefly
- `StrictHostKeyChecking=no` avoids interactive prompts in a non-interactive container startup context

## Preflight Checks

Before running any operations, verify prerequisites:

1. **FUSE check** — Stat `/dev/fuse`. If missing, log a clear error explaining the devcontainer needs FUSE access and exit
2. **SSHFS binary** — Check if `sshfs` is in PATH. If not, attempt to install it via the detected package manager. Fail with a clear message if unavailable
3. **SSH connectivity** — Run `ssh -o ConnectTimeout=5 -o BatchMode=yes user@host exit` to verify the connection works before attempting mounts

If no mounts are defined, all preflight checks are skipped — package-only and copy-only configs work without any prerequisites.

## Error Handling

- **Fail-forward on individual operations** — If one mount fails, log the error and continue with remaining mounts and copies. Don't let one broken entry block everything else.
- **Exit code** — Exit 0 if all operations succeeded. Exit 1 if any operation failed (after completing all others).
- **Logging** — Print clear, prefixed log lines to stdout:
  ```
  [teeleport] loading config from ~/dotfiles/teeleport.config
  [teeleport] packages: detected apt
  [teeleport] packages: installing curl, wget, jq, ripgrep, htop ... ok
  [teeleport] preflight: FUSE available ✓
  [teeleport] preflight: sshfs installed ✓
  [teeleport] preflight: SSH connection to my-server.example.com ✓
  [teeleport] mount: claude → ~/.claude ... ok
  [teeleport] mount: ssh-keys → ~/.ssh ... ok
  [teeleport] copy: bashrc → ~/.bashrc (replace) ... ok
  [teeleport] copy: bash-aliases → ~/.bashrc (append) ... ok
  [teeleport] ai-cli: installing claude-code ... ok
  [teeleport] ai-cli: running startup prompt ... ok
  [teeleport] done: 5 packages, 2 mounts, 2 copies, ai-cli ✓ (0 errors)
  ```

## Config File Discovery

The program finds `teeleport.config` using this priority:

1. `--config` CLI flag (explicit path)
2. `TEELEPORT_CONFIG` environment variable
3. `~/dotfiles/teeleport.config` (common devcontainer dotfile clone location)
4. `~/.dotfiles/teeleport.config`

First match wins. If none found, exit with an error explaining where it looked.

## Build & Distribution

- Build with `CGO_ENABLED=0` for a fully static binary
- Cross-compile for `linux/amd64` and `linux/arm64` (covers all common devcontainer hosts)
- Release binaries via GitHub Releases
- Install script detects arch and downloads the correct binary
