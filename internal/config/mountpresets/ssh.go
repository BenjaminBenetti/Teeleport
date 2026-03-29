package mountpresets

import "github.com/BenjaminBenetti/Teeleport/internal/domainmodel"

// SSH defines the mount preset for SSH known hosts.
// It mounts the known_hosts file so that accepted host keys persist across containers.
var SSH = []domainmodel.MountEntry{
	{Name: "ssh_known_hosts", Source: "/var/opt/teeleport/.ssh/known_hosts", Target: "~/.ssh/known_hosts", Type: "file"},
}
