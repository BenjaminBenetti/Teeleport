# Teeleport - Project Overview

## What is Teeleport?

Teeleport is a Go CLI tool that creates a "global home folder" experience across devcontainer/VS Code workspaces. It enables persistent state and consistent configuration across ephemeral development environments.

## How it works

1. User adds a `curl | bash` line to their dotfile install script
2. This downloads the latest Teeleport Go binary and executes it
3. Teeleport reads a `teeleport.config` file located in the user's dotfile repo
4. Based on the config, it performs **mount** and **copy** operations to set up the environment

## Core Operations

### Mount Operations

Mount remote directories into the local filesystem for live, bidirectional state.

- **Primary use case:** Directories that need to stay in sync across workspaces in real-time (e.g., `~/.claude` to persist login sessions)
- **Backend:** SSHFS (requires FUSE access in the container)
- **Prerequisite:** The devcontainer must be configured with FUSE access (`"privileged": true` or `"runArgs": ["--device=/dev/fuse"]`)
- **Future consideration:** The backend interface should be abstract enough to support additional mount backends (NFS, etc.) later

### Copy Operations

Copy files from the dotfile repo to the correct locations in the home directory. These are purely local operations — no remote pulls involved, since the files are already present in the cloned dotfile repo.

**Copy modes:**
- **Replace** — Overwrite the target file entirely with the source
- **Append** — Append source content to the target file (must be idempotent — use sentinel markers like `# BEGIN TEELEPORT` / `# END TEELEPORT` to prevent duplication on repeated runs)

## Architecture Decisions

- **Language:** Go — produces a single static binary with no runtime dependencies, ideal for download-and-run in containers
- **Distribution:** Single binary downloaded via URL, executed directly
- **Configuration:** `teeleport.config` file lives in the user's dotfile repo alongside their other dotfiles
- **Backend abstraction:** Mount operations should use an interface/backend pattern (strategy pattern) so new mount types can be added without changing core logic

### Package Install Operations

Install system packages into the container at startup. Uses the strategy pattern to support multiple package managers:

- **apt** — Ubuntu/Debian (most common devcontainer base)
- **dnf** — RHEL/Fedora
- **pacman** — Arch Linux

Teeleport auto-detects which package manager is available — the user just lists package names. Naturally idempotent (package managers skip already-installed packages).

### AI CLI Operations

Install and auto-invoke an AI coding CLI tool. Uses the strategy pattern to support the major AI CLIs:

- **Claude Code** — `npm install -g @anthropic-ai/claude-code`
- **Codex** — `npm install -g @openai/codex`
- **Gemini CLI** — `npm install -g @anthropic-ai/gemini-cli`
- **GitHub Copilot CLI** — `gh extension install github/gh-copilot`

After installation, Teeleport can optionally run the CLI with a startup prompt — either inline in the config or from an external file in the dotfile repo. This enables automated setup tasks on container start (e.g., "review the project and set up your memory").

**Critical design choice:** AI CLI failures are **never fatal**. The first time a container starts, the user may need to log in interactively — that's expected. Once auth is persisted (e.g., via a `.claude` mount), subsequent runs will work automatically. Teeleport logs the failure as a warning and continues.

## Key Concerns

### FUSE Requirement
SSHFS requires FUSE, which requires explicit container permissions. This is the primary friction point for mount operations and must be clearly documented. Copy operations have no such requirement.

### SSH Key Bootstrap
For SSH mounts, the container needs SSH access to the remote host. Devcontainer SSH agent forwarding helps here, but the bootstrap order matters — Teeleport should verify SSH connectivity before attempting mounts.

### Idempotency
The dotfile install script runs on every container creation. All operations (mounts and copies) must be safe to run repeatedly without side effects:
- Mounts: Check if already mounted before attempting
- Replace copies: Overwrite is naturally idempotent
- Append copies: Use sentinel markers to avoid duplication

### Mount Readiness
Other processes may try to access mounted paths before Teeleport finishes. Teeleport runs early in the dotfile script, but downstream hooks should be aware of this.

### Permissions / UID Mapping
Mounted filesystems may have different uid/gid than the container user. SSHFS provides `uid`/`gid` mapping options — these should be configurable in `teeleport.config` (default 1000/1000)
