package mountpresets

import "github.com/BenjaminBenetti/Teeleport/internal/domainmodel"

// GitConfig defines the mount preset for Git configuration.
// It mounts the .gitconfig file which contains the user's Git settings
// such as name, email, aliases, and other preferences.
var GitConfig = []domainmodel.MountEntry{
	{Name: "gitconfig", Source: "/var/opt/teeleport/.gitconfig", Target: "~/.gitconfig", Type: "file"},
}
