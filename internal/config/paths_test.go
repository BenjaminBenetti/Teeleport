package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath_TildePrefix(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not get home dir: %v", err)
	}

	got := ExpandPath("~/Documents/stuff")
	want := filepath.Join(home, "Documents/stuff")
	if got != want {
		t.Errorf("ExpandPath(\"~/Documents/stuff\") = %q, want %q", got, want)
	}
}

func TestExpandPath_TildeOnly(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not get home dir: %v", err)
	}

	got := ExpandPath("~")
	if got != home {
		t.Errorf("ExpandPath(\"~\") = %q, want %q", got, home)
	}
}

func TestExpandPath_AbsolutePath(t *testing.T) {
	input := "/usr/local/bin"
	got := ExpandPath(input)
	if got != input {
		t.Errorf("ExpandPath(%q) = %q, want unchanged", input, got)
	}
}

func TestExpandPath_RelativePath(t *testing.T) {
	input := "relative/path/to/file"
	got := ExpandPath(input)
	if got != input {
		t.Errorf("ExpandPath(%q) = %q, want unchanged", input, got)
	}
}

func TestExpandPath_EmptyString(t *testing.T) {
	got := ExpandPath("")
	if got != "" {
		t.Errorf("ExpandPath(\"\") = %q, want empty string", got)
	}
}

func TestResolvePath_RelativeJoinedToBase(t *testing.T) {
	got := ResolvePath("/home/user/dotfiles", "configs/vimrc")
	want := "/home/user/dotfiles/configs/vimrc"
	if got != want {
		t.Errorf("ResolvePath(\"/home/user/dotfiles\", \"configs/vimrc\") = %q, want %q", got, want)
	}
}

func TestResolvePath_AbsoluteRelativeReturnedAsIs(t *testing.T) {
	got := ResolvePath("/home/user/dotfiles", "/etc/config")
	want := "/etc/config"
	if got != want {
		t.Errorf("ResolvePath(\"/home/user/dotfiles\", \"/etc/config\") = %q, want %q", got, want)
	}
}

func TestResolvePath_TildeInRelative(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not get home dir: %v", err)
	}

	got := ResolvePath("/some/base", "~/.bashrc")
	want := filepath.Join(home, ".bashrc")
	if got != want {
		t.Errorf("ResolvePath(\"/some/base\", \"~/.bashrc\") = %q, want %q", got, want)
	}
}

func TestResolvePath_TildeInBase(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not get home dir: %v", err)
	}

	got := ResolvePath("~/dotfiles", "vimrc")
	want := filepath.Join(home, "dotfiles", "vimrc")
	if got != want {
		t.Errorf("ResolvePath(\"~/dotfiles\", \"vimrc\") = %q, want %q", got, want)
	}
}
