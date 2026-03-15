# teeleport.config — Configuration Reference

## Format

YAML. The config file lives in the root of the user's dotfile repo alongside their other dotfiles.

## Example

```yaml
# Where the dotfile repo is cloned in the container
# Devcontainers clone dotfiles to a known location, but this may vary
dotfile_repo: ~/dotfiles

# Mount operations — remote directories mounted into the local filesystem
mounts:
  # SSH connection for the remote host
  ssh:
    host: my-server.example.com
    user: devuser
    port: 22
    # Identity file (optional — if omitted, relies on SSH agent forwarding)
    identity_file: ~/.ssh/id_ed25519

  # UID/GID mapping for mounted filesystems
  # Defaults to 1000/1000 (standard devcontainer user)
  permissions:
    uid: 1000
    gid: 1000

  entries:
    - name: claude
      source: /home/devuser/.claude
      target: ~/.claude
      backend: sshfs

    - name: ssh-keys
      source: /home/devuser/.ssh
      target: ~/.ssh
      backend: sshfs

# Copy operations — local files from the dotfile repo copied to target locations
copies:
  - name: bashrc
    source: config/.bashrc
    target: ~/.bashrc
    mode: replace

  - name: bash-aliases
    source: config/.bash_aliases
    target: ~/.bashrc
    mode: append

  - name: gitconfig
    source: config/.gitconfig
    target: ~/.gitconfig
    mode: replace

# Packages to install — auto-detects package manager (apt, dnf, pacman)
packages:
  - curl
  - wget
  - jq
  - ripgrep
  - htop

# AI CLI tool — install and optionally run with a startup prompt
ai_cli:
  tool: claude-code
  # Startup prompt — inline string or path to a file in the dotfile repo
  startup_prompt: "Review the project, set up your memory, and check for any issues."
  # Or point to an external file:
  # startup_prompt_file: prompts/startup.md
```

## Top-Level Fields

### `dotfile_repo`

| | |
|---|---|
| **Type** | `string` (path) |
| **Required** | Yes |
| **Description** | Absolute or `~`-relative path to where the dotfile repo is cloned in the container. All copy `source` paths are resolved relative to this directory. |

### `mounts`

Container for all mount-related configuration.

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `ssh` | `object` | Yes (if entries are defined) | — | SSH connection settings for the remote host |
| `permissions` | `object` | No | `uid: 1000, gid: 1000` | UID/GID mapping for mounted filesystems |
| `entries` | `list` | No | `[]` | List of mount entries |

#### `mounts.ssh`

SSH connection to the single remote host. All mount entries use this connection.

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `host` | `string` | Yes | — | Remote hostname or IP |
| `user` | `string` | No | Current user | SSH username |
| `port` | `int` | No | `22` | SSH port |
| `identity_file` | `string` (path) | No | — | Path to SSH private key. If omitted, relies on SSH agent forwarding |

#### `mounts.permissions`

UID/GID mapping applied to all mounted filesystems.

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `uid` | `int` | No | `1000` | Local UID to map files to |
| `gid` | `int` | No | `1000` | Local GID to map files to |

#### Mount Entry Fields

Each entry under `mounts.entries` defines a remote directory to mount locally. All entries share the same SSH connection.

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `name` | `string` | Yes | — | Human-readable identifier for logging and error messages |
| `source` | `string` (path) | Yes | — | Absolute path on the remote host |
| `target` | `string` (path) | Yes | — | Local path to mount to (supports `~`) |
| `backend` | `string` | Yes | — | Mount backend to use. Currently supported: `sshfs` |

## Copy Entry Fields

Each entry under `copies:` defines a file to copy from the dotfile repo into the container.

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `name` | `string` | Yes | — | Human-readable identifier for logging and error messages |
| `source` | `string` (path) | Yes | — | Path relative to `dotfile_repo` |
| `target` | `string` (path) | Yes | — | Absolute or `~`-relative destination path |
| `mode` | `string` | Yes | — | Copy strategy: `replace` or `append` |

### Copy Modes

**`replace`**
Overwrites the target file entirely with the source file contents. Naturally idempotent — running multiple times produces the same result.

**`append`**
Appends the source file contents to the target file. To ensure idempotency, content is wrapped in sentinel markers:

```bash
# BEGIN TEELEPORT: bash-aliases
<appended content>
# END TEELEPORT: bash-aliases
```

The `name` field is used in the sentinel markers. On subsequent runs, the existing block is replaced in-place rather than appended again.

### `packages`

A flat list of package names to install. Teeleport auto-detects the system's package manager and uses it to install all listed packages.

| | |
|---|---|
| **Type** | `list` of `string` |
| **Required** | No |
| **Description** | Package names to install. The appropriate package manager is auto-detected: `apt` (Ubuntu/Debian), `dnf` (RHEL/Fedora), or `pacman` (Arch). If no supported package manager is found, this step is skipped with a warning. |

Note: Package names may differ across distros (e.g., `ripgrep` on apt vs `ripgrep` on pacman). Users are responsible for listing the correct names for their devcontainer base image.

### `ai_cli`

Configuration for AI CLI tool installation and startup prompt.

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `tool` | `string` | Yes | — | AI CLI to install. Supported: `claude-code`, `codex`, `gemini-cli`, `copilot` |
| `startup_prompt` | `string` | No | — | Inline prompt to pass to the CLI after installation. Mutually exclusive with `startup_prompt_file` |
| `startup_prompt_file` | `string` (path) | No | — | Path (relative to `dotfile_repo`) to a text file containing the startup prompt. Mutually exclusive with `startup_prompt` |

#### Supported Tools

| `tool` value | Install command | Invoke command |
|---|---|---|
| `claude-code` | `npm install -g @anthropic-ai/claude-code` | `claude -p "<prompt>"` |
| `codex` | `npm install -g @openai/codex` | `codex "<prompt>"` |
| `gemini-cli` | `npm install -g @anthropic-ai/gemini-cli` | `gemini -p "<prompt>"` |
| `copilot` | `gh extension install github/gh-copilot` | `gh copilot suggest "<prompt>"` |

#### Error Handling

AI CLI operations **never cause Teeleport to fail**. If installation or the startup prompt fails:
- The error is logged as a warning
- Teeleport continues with remaining operations
- Exit code is still 0 (AI CLI errors don't count as failures)

This is intentional — on the very first run, the user likely needs to log in interactively. Once auth state is persisted (e.g., via a `.claude` SSHFS mount), subsequent runs work automatically.

## Execution Order

1. **Packages are installed first** — Ensures dependencies (like `sshfs`) are available before mount operations
2. **Mounts are processed second** — Remote filesystems are mounted before any copies, in case a copy target is within a mounted path
3. **Copies are processed third** — Entries are processed top-to-bottom as listed in the config
4. **AI CLI runs last** — Install and invoke after everything else is set up (mounts in place, config files copied). Failures are non-fatal

## Path Resolution

- `~` is expanded to the container user's home directory
- Copy `source` paths are relative to `dotfile_repo`
- Mount `source` paths are absolute paths on the remote host
- All `target` paths are local to the container
