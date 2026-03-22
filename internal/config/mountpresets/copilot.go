package mountpresets

import "github.com/BenjaminBenetti/Teeleport/internal/domainmodel"

// Copilot defines the mount preset for GitHub's Copilot CLI.
// It mounts the .copilot directory which contains config, MCP settings, instructions, and session state.
var Copilot = []domainmodel.MountEntry{
	{Name: "copilot", Source: "/var/opt/teeleport/.copilot", Target: "~/.copilot"},
}
