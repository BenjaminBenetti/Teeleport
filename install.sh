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
