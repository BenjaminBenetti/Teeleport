package rsync

import (
	"testing"

	"github.com/BenjaminBenetti/Teeleport/internal/domainmodel"
)

func TestBuildSSHCommand_Defaults(t *testing.T) {
	ssh := domainmodel.SSHConfig{Host: "example.com", Port: 22}
	got := buildSSHCommand(ssh)
	want := "ssh -p 22 -o StrictHostKeyChecking=no -o ConnectTimeout=10 -o BatchMode=yes"
	if got != want {
		t.Errorf("buildSSHCommand() = %q, want %q", got, want)
	}
}

func TestBuildSSHCommand_WithIdentityFile(t *testing.T) {
	ssh := domainmodel.SSHConfig{Host: "example.com", Port: 2222, IdentityFile: "/tmp/test_key"}
	got := buildSSHCommand(ssh)
	want := `ssh -p 2222 -o StrictHostKeyChecking=no -o ConnectTimeout=10 -o BatchMode=yes -i "/tmp/test_key"`
	if got != want {
		t.Errorf("buildSSHCommand() = %q, want %q", got, want)
	}
}

func TestBuildSSHCommand_IdentityFileWithSpaces(t *testing.T) {
	ssh := domainmodel.SSHConfig{Host: "example.com", Port: 22, IdentityFile: "/tmp/my key"}
	got := buildSSHCommand(ssh)
	want := `ssh -p 22 -o StrictHostKeyChecking=no -o ConnectTimeout=10 -o BatchMode=yes -i "/tmp/my key"`
	if got != want {
		t.Errorf("buildSSHCommand() = %q, want %q", got, want)
	}
}

func TestRemoteAddress(t *testing.T) {
	ssh := domainmodel.SSHConfig{Host: "myhost.com", User: "dev"}
	got := remoteAddress(ssh, "/home/dev/project")
	want := "dev@myhost.com:/home/dev/project"
	if got != want {
		t.Errorf("remoteAddress() = %q, want %q", got, want)
	}
}

func TestRemoteAddress_EmptyUser(t *testing.T) {
	ssh := domainmodel.SSHConfig{Host: "myhost.com"}
	got := remoteAddress(ssh, "/data")
	// Should use current OS user, just verify format
	if got == "" {
		t.Error("remoteAddress() returned empty string")
	}
	if got[len(got)-len("myhost.com:/data"):] != "myhost.com:/data" {
		t.Errorf("remoteAddress() = %q, expected to end with myhost.com:/data", got)
	}
}

func TestHasRsyncEntries(t *testing.T) {
	tests := []struct {
		name    string
		entries []domainmodel.MountEntry
		want    bool
	}{
		{"empty", nil, false},
		{"sshfs only", []domainmodel.MountEntry{{Backend: "sshfs"}}, false},
		{"rsync present", []domainmodel.MountEntry{{Backend: "sshfs"}, {Backend: "rsync"}}, true},
		{"rsync only", []domainmodel.MountEntry{{Backend: "rsync"}}, true},
		{"default backend", []domainmodel.MountEntry{{Backend: ""}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasRsyncEntries(tt.entries); got != tt.want {
				t.Errorf("HasRsyncEntries() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterRsyncEntries(t *testing.T) {
	entries := []domainmodel.MountEntry{
		{Name: "a", Backend: "sshfs"},
		{Name: "b", Backend: "rsync"},
		{Name: "c", Backend: "rsync"},
		{Name: "d", Backend: ""},
	}
	got := FilterRsyncEntries(entries)
	if len(got) != 2 {
		t.Fatalf("FilterRsyncEntries() returned %d entries, want 2", len(got))
	}
	if got[0].Name != "b" || got[1].Name != "c" {
		t.Errorf("FilterRsyncEntries() = %v, want entries b and c", got)
	}
}
