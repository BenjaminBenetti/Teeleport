# Teeleport — CI/CD & Release Plan

## Overview

GitHub Actions handles building, testing, and publishing releases. A new release is triggered by pushing a version tag (e.g., `v1.0.0`). The pipeline cross-compiles the Go binary for both architectures and attaches them to a GitHub Release.

## Workflows

### 1. CI — Build & Test (on every push/PR)

**Trigger:** Push to any branch, pull requests to `main`

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: ["*"]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Build
        run: go build ./...

      - name: Test
        run: go test ./...

      - name: Vet
        run: go vet ./...

  build-matrix:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        arch: [amd64, arm64]
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Build binary
        env:
          CGO_ENABLED: "0"
          GOOS: linux
          GOARCH: ${{ matrix.arch }}
        run: go build -o teeleport-linux-${{ matrix.arch }} ./cmd/teeleport
```

### 2. Release — Build & Publish (on version tag)

**Trigger:** Push a tag matching `v*` (e.g., `v1.0.0`, `v0.2.1`)

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        arch: [amd64, arm64]
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Build binary
        env:
          CGO_ENABLED: "0"
          GOOS: linux
          GOARCH: ${{ matrix.arch }}
        run: go build -ldflags="-s -w -X main.version=${{ github.ref_name }}" -o teeleport-linux-${{ matrix.arch }} ./cmd/teeleport

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: teeleport-linux-${{ matrix.arch }}
          path: teeleport-linux-${{ matrix.arch }}

  publish:
    runs-on: ubuntu-latest
    needs: release
    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@v4
        with:
          merge-multiple: true

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          generate_release_notes: true
          files: |
            teeleport-linux-amd64
            teeleport-linux-arm64
```

## How to Cut a Release

```bash
# Tag the commit
git tag v1.0.0

# Push the tag — triggers the release workflow
git push origin v1.0.0
```

That's it. The workflow builds both binaries, creates a GitHub Release with auto-generated release notes, and attaches the binaries.

## Build Details

### Linker Flags

```
-ldflags="-s -w -X main.version=${{ github.ref_name }}"
```

| Flag | Purpose |
|---|---|
| `-s` | Strip symbol table — smaller binary |
| `-w` | Strip DWARF debug info — smaller binary |
| `-X main.version=...` | Embed the version tag into the binary at compile time |

This requires a `version` variable in `main.go`:

```go
var version = "dev"
```

When built locally it reads `"dev"`, when built by the release pipeline it reads `"v1.0.0"` (or whatever the tag is).

### CGO_ENABLED=0

Produces a fully static binary with no libc dependency. Essential for running in any container regardless of base image.

### Binary Size

With `-s -w` and no CGO, the binary should be ~5-10MB depending on dependencies. Small enough that the install script downloads it in seconds.

## Release Assets

Each GitHub Release contains:

```
v1.0.0
├── teeleport-linux-amd64    # x86_64 binary
├── teeleport-linux-arm64    # aarch64 binary
└── (auto-generated release notes)
```

The install script (`install.sh`) queries the GitHub API for the latest release and downloads the correct binary based on the detected architecture.

## Versioning

Follow semantic versioning:

- **Patch** (`v1.0.1`) — Bug fixes
- **Minor** (`v1.1.0`) — New features, backward compatible
- **Major** (`v2.0.0`) — Breaking changes to `teeleport.config` format
