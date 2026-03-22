package mountpresets

import "github.com/BenjaminBenetti/Teeleport/internal/domainmodel"

// Claude defines the mount preset for Anthropic's Claude Code.
// It mounts the .claude directory and .claude.json settings file
// from the remote host.
var Claude = []domainmodel.MountEntry{
	{Name: "claude", Source: "/var/opt/teeleport/.claude", Target: "~/.claude"},
	{Name: "claude-json", Source: "/var/opt/teeleport/.claude.json", Target: "~/.claude.json", Type: "file"},
}
