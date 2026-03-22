package mountpresets

import "github.com/BenjaminBenetti/Teeleport/internal/domainmodel"

// GH defines the mount preset for GitHub CLI (gh).
// It mounts the .config/gh directory which contains config, auth tokens, and aliases.
var GH = []domainmodel.MountEntry{
	{Name: "gh", Source: "/var/opt/teeleport/.config/gh", Target: "~/.config/gh"},
}
