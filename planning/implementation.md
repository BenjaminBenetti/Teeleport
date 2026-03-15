# Teeleport — Implementation Plan

## Task Map

Tasks are grouped into phases. Within each phase, tasks marked with the same parallel group letter (e.g., **[A]**) can be worked on simultaneously. Tasks without a group letter are sequential prerequisites.

---

## Phase 1: Project Scaffold

Everything else depends on this. Must be done first.

### Task 1.1 — Initialize Go module and project structure

- `go mod init github.com/BenjaminBenetti/Teeleport`
- Create directory skeleton: `cmd/teeleport/`, `internal/config/`, `internal/mount/`, `internal/copy/`, `internal/packages/`, `internal/aicli/`, `internal/preflight/`
- Stub `cmd/teeleport/main.go` with version variable and basic arg parsing
- Add `gopkg.in/yaml.v3` dependency

### Task 1.2 — Config package

- `internal/config/config.go` — All config structs (`Config`, `MountConfig`, `SSHConfig`, `PermConfig`, `MountEntry`, `CopyEntry`, `AICLIConfig`)
- YAML parsing function: `LoadConfig(path string) (*Config, error)`
- Validation: required fields, mutually exclusive `startup_prompt` / `startup_prompt_file`
- `internal/config/paths.go` — `ExpandPath()` for `~` expansion, `ResolvePath()` for relative-to-dotfile-repo resolution
- Config discovery logic (CLI flag → env var → default paths)

### Task 1.3 — Example config file

- Create `teeleport.config.example` in repo root with all features demonstrated

---

## Phase 2: Core Feature Packages

All four feature packages are **independent of each other**. They only depend on Phase 1 (config types). Build them all in parallel.

### Task 2A.1 — Packages package `[A]`

- `internal/packages/packages.go` — `PackageManager` interface with `Update()` and `Install([]string)` methods
- Auto-detection function: `Detect() (PackageManager, error)` — checks PATH for `apt-get`, `dnf`, `pacman`

### Task 2A.2 — Package manager backends `[A]`

- `internal/packages/apt.go` — `AptManager` struct implementing `PackageManager`
- `internal/packages/dnf.go` — `DnfManager` struct
- `internal/packages/pacman.go` — `PacmanManager` struct
- Each backend: `Update()` runs the index refresh, `Install()` runs the install command

### Task 2B.1 — Mount package `[B]`

- `internal/mount/mount.go` — `MountBackend` interface with `Mount()` and `IsMounted()` methods
- Backend factory: `NewBackend(backendName string, ssh SSHConfig, perms PermConfig) (MountBackend, error)`

### Task 2B.2 — SSHFS backend `[B]`

- `internal/mount/sshfs.go` — `SSHFSBackend` struct implementing `MountBackend`
- Command construction with all flags (port, uid, gid, identity file, reconnect, etc.)
- `IsMounted()` checks `/proc/mounts` for the target path

### Task 2B.3 — Mount manager `[B]`

- `internal/mount/manager.go` — `ProcessMounts(cfg MountConfig) error`
- Iterates entries, skips already-mounted, creates target dirs, calls backend

### Task 2C.1 — Copy package `[C]`

- `internal/copy/copy.go` — `ProcessCopies(dotfileRepo string, entries []CopyEntry) error`
- Replace mode: read source, write to target
- Append mode: delegate to sentinel logic

### Task 2C.2 — Sentinel logic `[C]`

- `internal/copy/sentinel.go` — `ApplyAppend(name, sourceContent, targetPath string) error`
- Parse existing file for `# BEGIN TEELEPORT: <name>` / `# END TEELEPORT: <name>` blocks
- If block exists: replace content between markers
- If block doesn't exist: append block at end of file

### Task 2D.1 — AI CLI package `[D]`

- `internal/aicli/aicli.go` — `AICli` interface with `Install()` and `Run(prompt string)` methods
- Factory function: `NewAICli(tool string) (AICli, error)`
- Prompt resolution: read inline string or load from file

### Task 2D.2 — AI CLI backends `[D]`

- `internal/aicli/claude.go` — `ClaudeCode` struct: `npm install -g @anthropic-ai/claude-code`, invoke with `claude -p`
- `internal/aicli/codex.go` — `Codex` struct: `npm install -g @openai/codex`, invoke with `codex`
- `internal/aicli/gemini.go` — `GeminiCli` struct: `npm install -g @anthropic-ai/gemini-cli`, invoke with `gemini -p`
- `internal/aicli/copilot.go` — `Copilot` struct: `gh extension install github/gh-copilot`, invoke with `gh copilot suggest`

---

## Phase 3: Preflight & Main Orchestration

Depends on Phase 2 packages being complete (needs mount and packages interfaces).

### Task 3.1 — Preflight checks

- `internal/preflight/preflight.go` — `RunChecks(cfg *Config) error`
- FUSE check: stat `/dev/fuse`
- SSHFS check: look in PATH, attempt install via packages if missing
- SSH connectivity: test connection to configured host

### Task 3.2 — Main orchestration (`cmd/teeleport/main.go`)

- Wire everything together in execution order:
  1. Locate and parse config
  2. Install packages
  3. Preflight checks (if mounts defined)
  4. Process mounts
  5. Process copies
  6. AI CLI (install + run prompt)
  7. Log summary and exit
- Error collection: track failures per stage, fail-forward, report at end
- Logging: `[teeleport]` prefixed output

---

## Phase 4: Install Script & CI

Independent of each other. Can be done in parallel. Can also overlap with Phase 3.

### Task 4A.1 — Install script `[A]`

- Create `install.sh` in repo root
- Architecture detection, GitHub API latest release lookup, binary download
- Optional `--version` flag for pinning
- Make sure it works with `curl | bash`

### Task 4B.1 — CI workflow `[B]`

- `.github/workflows/ci.yml` — build, test, vet on push/PR
- Build matrix for amd64 and arm64

### Task 4B.2 — Release workflow `[B]`

- `.github/workflows/release.yml` — triggered on `v*` tags
- Cross-compile with ldflags (strip symbols, embed version)
- Upload artifacts, create GitHub Release with both binaries

---

## Phase 5: Testing & Polish

### Task 5.1 — Unit tests

- Config parsing tests (valid config, missing fields, bad YAML)
- Path resolution tests (~ expansion, relative paths)
- Sentinel logic tests (first append, re-append, multiple blocks)
- Package manager detection tests (mock PATH lookups)

### Task 5.2 — Example dotfile repo

- Create `examples/dotfiles/` directory with a complete working example:
  - `install.sh`
  - `teeleport.config`
  - `config/.bashrc`, `config/.gitconfig`
  - `prompts/startup.md`

### Task 5.3 — README

- Project description and motivation
- Quick start guide
- Config reference (link to planning/config.md or inline)
- Prerequisites (FUSE for mounts)

---

## Dependency Graph

```
Phase 1: Scaffold
  └── Task 1.1 → Task 1.2 → Task 1.3

Phase 2: Core Features (all parallel after Phase 1)
  ├── [A] Task 2A.1 → Task 2A.2    (packages)
  ├── [B] Task 2B.1 → Task 2B.2 → Task 2B.3  (mounts)
  ├── [C] Task 2C.1 → Task 2C.2    (copies)
  └── [D] Task 2D.1 → Task 2D.2    (ai-cli)

Phase 3: Orchestration (after Phase 2)
  └── Task 3.1 → Task 3.2

Phase 4: Install & CI (parallel, can overlap with Phase 3)
  ├── [A] Task 4A.1    (install script)
  └── [B] Task 4B.1 → Task 4B.2    (ci/release)

Phase 5: Testing & Polish (after Phase 3)
  ├── Task 5.1 (tests)
  ├── Task 5.2 (example)
  └── Task 5.3 (readme)
```

## Parallelism Summary

| Phase | Parallel Tracks | Max Concurrent Tasks |
|---|---|---|
| 1 | None — sequential | 1 |
| 2 | 4 tracks: packages, mounts, copies, ai-cli | 4 |
| 3 | None — sequential | 1 |
| 4 | 2 tracks: install script, CI/CD | 2 (can overlap with Phase 3) |
| 5 | 3 tracks: tests, example, readme | 3 |

**Critical path:** Phase 1 → Phase 2 (any track) → Phase 3 → Phase 5.1

Phase 4 is off the critical path entirely — it can be built anytime after Phase 1.
