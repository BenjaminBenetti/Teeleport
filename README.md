# Teeleport

A "global home folder" for devcontainers. Teeleport gives you persistent state, consistent configuration, and your favorite tools across every ephemeral VS Code / Codespaces / devcontainer workspace.

Add one line to your dotfile repo's install script and Teeleport handles the rest: mounting remote directories, copying config files, installing packages, and launching your AI coding assistant.

## Quick Start

### 1. Add Teeleport to your dotfile repo

In your dotfile repo's install script (e.g., `install.sh`), add:

```bash
curl -fsSL https://raw.githubusercontent.com/BenjaminBenetti/Teeleport/main/install.sh | bash
```

To pin a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/BenjaminBenetti/Teeleport/main/install.sh | bash -s -- --version v1.0.0
```

### 2. Create a `teeleport.config` file

Add a `teeleport.config` to the root of your dotfile repo:

```yaml
# Install system packages
packages:
  - curl
  - jq
  - ripgrep

# Copy config files from your dotfile repo into the container
copies:
  - name: bashrc
    source: config/.bashrc
    target: ~/.bashrc
    mode: replace

  - name: git-aliases
    source: config/.git_aliases
    target: ~/.bashrc
    mode: append

  - name: gitconfig
    source: config/.gitconfig
    target: ~/.gitconfig
    mode: replace

# Mount remote directories for persistent state (requires FUSE)
mounts:
  ssh:
    host: my-server.example.com
    user: devuser
  entries:
    - name: claude
      source: /home/devuser/.claude
      target: ~/.claude
      backend: sshfs

# Install and auto-run an AI CLI tool
ai_cli:
  tool: claude-code
  startup_prompt: "Review the project and set up your memory."
```

### 3. Point your devcontainer at your dotfile repo

In VS Code or GitHub Codespaces, set your dotfiles repository in settings. Every new container will automatically run your install script, which runs Teeleport, which sets up your entire environment.

## What Teeleport Does

When Teeleport runs, it executes these steps in order:

```
1. Install packages      apt/dnf/pacman (auto-detected)
2. Mount remote dirs     SSHFS mounts for live, persistent state
3. Copy config files     From your dotfile repo to the right locations
4. Launch AI CLI         Install and run with an optional startup prompt
```

All operations are **idempotent** -- safe to run on every container creation.

## Features

### Package Installation

List the packages you need and Teeleport installs them automatically. It detects whether the container uses `apt`, `dnf`, or `pacman`.

```yaml
packages:
  - curl
  - wget
  - jq
  - ripgrep
  - htop
```

### File Copies

Copy files from your dotfile repo to the correct locations in the container.

**Replace mode** overwrites the target file entirely:

```yaml
copies:
  - name: gitconfig
    source: config/.gitconfig
    target: ~/.gitconfig
    mode: replace
```

**Append mode** adds content to an existing file using sentinel markers to ensure idempotency:

```yaml
copies:
  - name: bash-aliases
    source: config/.bash_aliases
    target: ~/.bashrc
    mode: append
```

### Remote Mounts (SSHFS)

Mount directories from a remote host into your container for live, bidirectional state. This is how you keep things like `~/.claude` persistent across workspaces -- changes sync in real-time.

```yaml
mounts:
  ssh:
    host: my-server.example.com
    user: devuser
  entries:
    - name: claude
      source: /home/devuser/.claude
      target: ~/.claude
      backend: sshfs
```

**File mounts** work the same way but for individual files. Under the hood, Teeleport mounts the remote parent directory to a staging area and symlinks the file to the target path:

```yaml
mounts:
  ssh:
    host: my-server.example.com
    user: devuser
  entries:
    - name: claude-json
      source: /home/devuser/.claude.json
      target: ~/.claude.json
      type: file
      backend: sshfs
```

**Prerequisites for mounts:**

Your `devcontainer.json` must grant FUSE access. Add one of:

```jsonc
// Option A: privileged mode
"privileged": true

// Option B: explicit device access
"runArgs": ["--device=/dev/fuse"]
```

You also need SSH access to the remote host. Devcontainer [SSH agent forwarding](https://code.visualstudio.com/remote/advancedcontainers/sharing-git-credentials) is the easiest way to set this up.

### AI CLI Integration

Install and auto-invoke an AI coding CLI on container start. Supported tools:

| Tool | Config value | Install method |
|---|---|---|
| Claude Code | `claude-code` | `npm install -g @anthropic-ai/claude-code` |
| OpenAI Codex | `codex` | `npm install -g @openai/codex` |
| Gemini CLI | `gemini-cli` | `npm install -g @google/gemini-cli` |
| GitHub Copilot | `copilot` | `gh extension install github/gh-copilot` |

Provide a startup prompt inline or from a file:

```yaml
ai_cli:
  tool: claude-code
  startup_prompt: "Review the project and set up your memory."
  # Or use an external file:
  # startup_prompt_file: prompts/startup.md
```

AI CLI errors are **never fatal**. On first run you may need to log in interactively -- once auth is persisted (e.g., via a `.claude` mount), subsequent runs work automatically.

## Example Dotfile Repo Layout

```
dotfiles/
├── install.sh              # Your dotfile install script
├── teeleport.config        # Teeleport configuration
├── config/
│   ├── .bashrc             # Copied to ~/.bashrc (replace)
│   ├── .bash_aliases       # Appended to ~/.bashrc (append)
│   └── .gitconfig          # Copied to ~/.gitconfig (replace)
└── prompts/
    └── startup.md          # AI CLI startup prompt (optional)
```

Example `install.sh`:

```bash
#!/bin/bash

# Run Teeleport -- handles packages, mounts, copies, and AI CLI
curl -fsSL https://raw.githubusercontent.com/BenjaminBenetti/Teeleport/main/install.sh | bash

echo "dotfiles setup complete!"
```

## Configuration Reference

### `dotfile_repo`

Path to the dotfile repo root. All copy `source` paths resolve relative to this. Supports `~`. **Optional** — defaults to `.` (current working directory), which is correct in most cases since devcontainers run your install script from the cloned dotfile repo.

### `mounts.ssh`

| Field | Required | Default | Description |
|---|---|---|---|
| `host` | Yes | -- | Remote hostname or IP |
| `user` | No | Current user | SSH username |
| `port` | No | `22` | SSH port |
| `identity_file` | No | -- | Path to SSH key (supports `~`). Omit to use SSH agent forwarding (recommended). |

### `mounts.permissions`

| Field | Required | Default | Description |
|---|---|---|---|
| `uid` | No | `1000` | UID to map mounted files to |
| `gid` | No | `1000` | GID to map mounted files to |

### `mounts.entries[]`

| Field | Required | Default | Description |
|---|---|---|---|
| `name` | Yes | -- | Human-readable label for logs |
| `source` | Yes | -- | Absolute path on the remote host |
| `target` | Yes | -- | Local mount point (supports `~`) |
| `type` | No | `directory` | `directory` or `file`. File mounts symlink a single file from a staged parent directory mount. |
| `backend` | Yes | -- | Mount backend: `sshfs` |

### `copies[]`

| Field | Required | Description |
|---|---|---|
| `name` | Yes | Label for logs and sentinel markers |
| `source` | Yes | Path relative to `dotfile_repo` |
| `target` | Yes | Destination path (supports `~`) |
| `mode` | Yes | `replace` or `append` |

### `packages`

A flat list of package names. Teeleport auto-detects `apt`, `dnf`, or `pacman`.

### `ai_cli`

| Field | Required | Description |
|---|---|---|
| `tool` | Yes | `claude-code`, `codex`, `gemini-cli`, or `copilot` |
| `startup_prompt` | No | Inline prompt string (mutually exclusive with `startup_prompt_file`) |
| `startup_prompt_file` | No | Path to prompt file, relative to `dotfile_repo` |

## CLI Flags

```
teeleport [flags]

  --config <path>    Path to config file (overrides auto-discovery)
  --version          Print version and exit
```

Teeleport auto-discovers the config file in this order:

1. `--config` flag
2. `TEELEPORT_CONFIG` environment variable
3. `./teeleport.config` (current working directory)
4. `~/dotfiles/teeleport.config`
5. `~/.dotfiles/teeleport.config`

## Building from Source

```bash
# Build for current platform
go build -o teeleport ./cmd/teeleport

# Cross-compile static binaries
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o teeleport-linux-amd64 ./cmd/teeleport
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o teeleport-linux-arm64 ./cmd/teeleport
```

## License

MIT
