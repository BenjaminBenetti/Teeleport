#!/usr/bin/env bash
set -euo pipefail

log() {
    echo "[teeleport] $*"
}

fail() {
    log "ERROR: $*" >&2
    exit 1
}

# --- Verify OS ---
OS="$(uname -s)"
if [ "$OS" != "Linux" ]; then
    fail "Unsupported operating system: ${OS}. Teeleport only supports Linux."
fi

# --- Detect architecture ---
RAW_ARCH="$(uname -m)"
case "$RAW_ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    *)       fail "Unsupported architecture: ${RAW_ARCH}. Supported: x86_64 (amd64), aarch64 (arm64)." ;;
esac
log "Detected architecture: ${RAW_ARCH} -> ${ARCH}"

# --- Parse arguments ---
VERSION=""
while [ $# -gt 0 ]; do
    case "$1" in
        --version)
            if [ -z "${2:-}" ]; then
                fail "--version requires a value (e.g. --version v1.0.0)"
            fi
            VERSION="$2"
            shift 2
            ;;
        *)
            fail "Unknown argument: $1"
            ;;
    esac
done

# --- Resolve version ---
REPO="BenjaminBenetti/Teeleport"
if [ -z "$VERSION" ]; then
    log "Querying GitHub for latest release..."
    VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' \
        | sed -E 's/.*"tag_name":\s*"([^"]+)".*/\1/')"
    if [ -z "$VERSION" ]; then
        fail "Could not determine latest release version from GitHub."
    fi
fi
log "Installing version: ${VERSION}"

# --- Download binary ---
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/teeleport-linux-${ARCH}"
INSTALL_DIR="${HOME}/.local/bin"
INSTALL_PATH="${INSTALL_DIR}/teeleport"

mkdir -p "$INSTALL_DIR"

log "Downloading from ${DOWNLOAD_URL} ..."
if ! curl -fSL -o "$INSTALL_PATH" "$DOWNLOAD_URL"; then
    fail "Download failed. Check that version ${VERSION} exists and has a binary for linux-${ARCH}."
fi

# --- Make executable ---
chmod +x "$INSTALL_PATH"
log "Installed teeleport to ${INSTALL_PATH}"

# --- Ensure install dir is on PATH ---
export PATH="${INSTALL_DIR}:${PATH}"
log "Added ${INSTALL_DIR} to PATH"

# --- Run teeleport ---
log "Starting teeleport..."
teeleport

# --- Install shell hook to re-mount after devcontainer restart ---
# Resolve the config file path now (install-time CWD is the dotfiles repo)
# so the hook can find it regardless of the terminal's CWD at runtime.
TEELEPORT_CONFIG_PATH="$(teeleport -config-path 2>/dev/null || true)"
if [ -z "$TEELEPORT_CONFIG_PATH" ]; then
    # Fallback: probe the same candidates FindConfig uses
    for _candidate in \
        "$(pwd)/teeleport.config" \
        "$(pwd)/teeleport.config.yaml" \
        "${HOME}/dotfiles/teeleport.config" \
        "${HOME}/dotfiles/teeleport.config.yaml" \
        "${HOME}/.dotfiles/teeleport.config" \
        "${HOME}/.dotfiles/teeleport.config.yaml"; do
        if [ -f "$_candidate" ]; then
            TEELEPORT_CONFIG_PATH="$_candidate"
            break
        fi
    done
    unset _candidate
fi

HOOK_MARKER="teeleport-remount-hook"
# Note: unquoted HOOKEOF so TEELEPORT_CONFIG_PATH is expanded at install time.
# Runtime variables use \$ to defer expansion.
read -r -d '' HOOK_SNIPPET << HOOKEOF || true
# teeleport-remount-hook — re-establish SSHFS mounts after container restart
# Only run in interactive shells with SSH agent forwarding available.
if [ -x "\${HOME}/.local/bin/teeleport" ] && [[ \$- == *i* ]] && [ -n "\${SSH_AUTH_SOCK:-}" ]; then
    _tp_stamp="\${HOME}/.teeleport/.remount-stamp"
    _tp_boot="\$(stat -c %Z /proc/1 2>/dev/null || echo 0)"
    _tp_last="\$(cat "\$_tp_stamp" 2>/dev/null || echo -1)"
    if [ "\$_tp_boot" != "\$_tp_last" ]; then
        mkdir -p "\${HOME}/.teeleport"
        echo "\$_tp_boot" > "\$_tp_stamp"
        TEELEPORT_CONFIG="${TEELEPORT_CONFIG_PATH}" "\${HOME}/.local/bin/teeleport" >> "\${HOME}/.teeleport/remount.log" 2>&1 &
        disown
    fi
    unset _tp_stamp _tp_boot _tp_last
fi
HOOKEOF

install_hook() {
    local rc_file="$1"
    if [ -f "$rc_file" ] && grep -q "$HOOK_MARKER" "$rc_file"; then
        log "Shell hook already present in ${rc_file}, skipping."
    else
        echo "" >> "$rc_file"
        echo "$HOOK_SNIPPET" >> "$rc_file"
        log "Installed remount hook in ${rc_file}"
    fi
}

install_hook "${HOME}/.bashrc"
[ -f "${HOME}/.zshrc" ] && install_hook "${HOME}/.zshrc"
