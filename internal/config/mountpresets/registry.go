// Package mountpresets provides built-in mount preset definitions for Teeleport.
// Each preset is a named collection of MountEntry values that can be referenced
// in the config file to avoid manually defining common mount configurations.
package mountpresets

import (
	"fmt"

	"github.com/BenjaminBenetti/Teeleport/internal/domainmodel"
)

var registry = map[string][]domainmodel.MountEntry{
	"claude": Claude,
}

// Get returns the mount entries for the named preset.
// It returns an error if the preset name is not recognized.
func Get(name string) ([]domainmodel.MountEntry, error) {
	entries, ok := registry[name]
	if !ok {
		available := make([]string, 0, len(registry))
		for k := range registry {
			available = append(available, k)
		}
		return nil, fmt.Errorf("unknown mount preset %q; available presets: %v", name, available)
	}
	return entries, nil
}
