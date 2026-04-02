package mount

import (
	"testing"

	"github.com/BenjaminBenetti/Teeleport/internal/domainmodel"
)

func TestRsyncBackend_FsType(t *testing.T) {
	b := &RsyncBackend{}
	got := b.FsType()
	want := "rsync"
	if got != want {
		t.Errorf("RsyncBackend.FsType() = %q, want %q", got, want)
	}
}

func TestRsyncBackend_IsMounted(t *testing.T) {
	b := &RsyncBackend{}
	mounted, err := b.IsMounted("/tmp/any-path")
	if err != nil {
		t.Errorf("RsyncBackend.IsMounted() error = %v, want nil", err)
	}
	if mounted {
		t.Error("RsyncBackend.IsMounted() = true, want false (rsync never reports mounted)")
	}
}

func TestNewBackend_Rsync(t *testing.T) {
	backend, err := NewBackend("rsync", defaultSSH(), defaultPerms())
	if err != nil {
		t.Fatalf("NewBackend(rsync) error = %v", err)
	}
	if backend == nil {
		t.Fatal("NewBackend(rsync) returned nil")
	}
	if _, ok := backend.(*RsyncBackend); !ok {
		t.Errorf("NewBackend(rsync) returned %T, want *RsyncBackend", backend)
	}
}

func TestNewBackend_Unknown(t *testing.T) {
	_, err := NewBackend("nfs", defaultSSH(), defaultPerms())
	if err == nil {
		t.Fatal("NewBackend(nfs) should return error for unknown backend")
	}
}

func defaultSSH() domainmodel.SSHConfig {
	return domainmodel.SSHConfig{Host: "test.example.com", Port: 22}
}

func defaultPerms() domainmodel.PermConfig {
	uid, gid := 1000, 1000
	return domainmodel.PermConfig{UID: &uid, GID: &gid}
}
