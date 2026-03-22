package mountpresets

import "github.com/BenjaminBenetti/Teeleport/internal/domainmodel"

// Gemini defines the mount preset for Google's Gemini CLI.
// It mounts the .gemini directory which contains settings, shell history, and project data.
var Gemini = []domainmodel.MountEntry{
	{Name: "gemini", Source: "/var/opt/teeleport/.gemini", Target: "~/.gemini"},
}
