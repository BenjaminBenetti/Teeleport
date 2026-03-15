# Teeleport — Install Script Plan

## Overview

Teeleport is installed by adding a single line to the user's dotfile install script. When devcontainers clone the dotfile repo, they execute the install script, which downloads and runs the Teeleport binary.

## User Setup

The user adds this line to their dotfile install script (e.g., `install.sh`, `setup.sh`, or `bootstrap.sh`):

```bash
curl -fsSL https://raw.githubusercontent.com/BenjaminBenetti/Teeleport/main/install.sh | bash
```

That's it. One line. The install script handles everything else.

## What the Install Script Does

```
curl install.sh | bash
  ├── 1. Detect system architecture (amd64 / arm64)
  ├── 2. Detect OS (linux only — fail with message on others)
  ├── 3. Determine latest release version
  │     └── Query GitHub API: https://api.github.com/repos/BenjaminBenetti/Teeleport/releases/latest
  ├── 4. Download the correct binary
  │     └── https://github.com/BenjaminBenetti/Teeleport/releases/download/{version}/teeleport-linux-{arch}
  ├── 5. Make binary executable
  ├── 6. Run teeleport
  │     └── Binary auto-discovers teeleport.config from the dotfile repo
  └── 7. Clean up (optional — remove binary or keep it for manual re-runs)
```

## Install Script

```bash
#!/bin/bash
set -euo pipefail

REPO="BenjaminBenetti/Teeleport"
INSTALL_DIR="${HOME}/.local/bin"
BINARY_NAME="teeleport"

# Detect architecture
ARCH=$(uname -m)
case "${ARCH}" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    *)
        echo "[teeleport] unsupported architecture: ${ARCH}"
        exit 1
        ;;
esac

# Verify we're on Linux
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [ "${OS}" != "linux" ]; then
    echo "[teeleport] unsupported OS: ${OS} (teeleport only supports linux)"
    exit 1
fi

# Get latest release version from GitHub API
echo "[teeleport] fetching latest release..."
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "${VERSION}" ]; then
    echo "[teeleport] failed to determine latest version"
    exit 1
fi

echo "[teeleport] latest version: ${VERSION}"

# Download binary
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}-${OS}-${ARCH}"
echo "[teeleport] downloading ${DOWNLOAD_URL}..."

mkdir -p "${INSTALL_DIR}"
curl -fsSL -o "${INSTALL_DIR}/${BINARY_NAME}" "${DOWNLOAD_URL}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# Ensure install dir is in PATH
export PATH="${INSTALL_DIR}:${PATH}"

# Run teeleport
echo "[teeleport] running teeleport..."
"${INSTALL_DIR}/${BINARY_NAME}"
```

## Key Decisions

### Binary Location
Installed to `~/.local/bin/teeleport`. This is the standard user-local binary path on Linux and is often already in PATH in devcontainers. The script also exports it to PATH explicitly to ensure it's available immediately.

### Keep Binary After Run
The binary is kept in `~/.local/bin` so the user can manually re-run `teeleport` if needed (e.g., after changing their config). It's a small static binary — no reason to clean it up.

### No Version Pinning (By Default)
The script always pulls the latest release. This ensures users get bug fixes and new features automatically. If a user wants to pin a version, they can modify the curl line:

```bash
curl -fsSL https://raw.githubusercontent.com/BenjaminBenetti/Teeleport/main/install.sh | bash -s -- --version v1.2.3
```

The script can optionally accept a `--version` flag to override the latest lookup:

```bash
# At the top of the script, after set -euo pipefail:
VERSION_OVERRIDE=""
while [[ $# -gt 0 ]]; do
    case "$1" in
        --version) VERSION_OVERRIDE="$2"; shift 2 ;;
        *) shift ;;
    esac
done

# Then when determining version:
if [ -n "${VERSION_OVERRIDE}" ]; then
    VERSION="${VERSION_OVERRIDE}"
else
    VERSION=$(curl -fsSL ...)
fi
```

### No sudo
The script installs to user space (`~/.local/bin`) and never requires root. Teeleport itself may need sudo for package installation (handled internally by the Go binary), but the install script stays unprivileged.

## Release Asset Naming

The GitHub release should include these assets:

```
teeleport-linux-amd64
teeleport-linux-arm64
```

No `.tar.gz`, no `.zip` — just raw binaries. Keeps the install script simple (no extraction step).

## Example: Complete Dotfile Repo Layout

```
dotfiles/
├── install.sh                    # User's dotfile install script
├── teeleport.config              # Teeleport configuration
├── config/
│   ├── .bashrc                   # Bash config (copied by teeleport)
│   ├── .bash_aliases             # Aliases (appended by teeleport)
│   └── .gitconfig                # Git config (copied by teeleport)
└── prompts/
    └── startup.md                # AI CLI startup prompt (optional)
```

### Example `install.sh`

```bash
#!/bin/bash

# Install teeleport and run it (handles mounts, copies, packages, ai-cli)
curl -fsSL https://raw.githubusercontent.com/BenjaminBenetti/Teeleport/main/install.sh | bash

# Anything else the user wants to do after teeleport runs
echo "dotfiles setup complete!"
```

The user points their VS Code / GitHub Codespaces dotfile setting to this repo, and everything happens automatically on container creation.
