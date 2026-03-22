package mountpresets

import "github.com/BenjaminBenetti/Teeleport/internal/domainmodel"

// Codex defines the mount preset for OpenAI's Codex CLI.
// It mounts the .codex directory which contains config, auth, agents, rules, sessions, and logs.
var Codex = []domainmodel.MountEntry{
	{Name: "codex", Source: "/var/opt/teeleport/.codex", Target: "~/.codex"},
}
