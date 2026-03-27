package mountpresets

import "testing"

func TestGet_Claude(t *testing.T) {
	entries, err := Get("claude")
	if err != nil {
		t.Fatalf("Get(\"claude\") returned error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("Get(\"claude\") returned %d entries, want 2", len(entries))
	}
	if entries[0].Name != "claude" {
		t.Errorf("entries[0].Name = %q, want \"claude\"", entries[0].Name)
	}
	if entries[0].Source != "/var/opt/teeleport/.claude" {
		t.Errorf("entries[0].Source = %q, want \"/var/opt/teeleport/.claude\"", entries[0].Source)
	}
	if entries[1].Name != "claude-json" {
		t.Errorf("entries[1].Name = %q, want \"claude-json\"", entries[1].Name)
	}
	if entries[1].Type != "file" {
		t.Errorf("entries[1].Type = %q, want \"file\"", entries[1].Type)
	}
	if entries[1].File.DefaultContent != "{}" {
		t.Errorf("entries[1].File.DefaultContent = %q, want \"{}\"", entries[1].File.DefaultContent)
	}
	// Verify no backend is set (presets are backend-agnostic)
	if entries[0].Backend != "" {
		t.Errorf("entries[0].Backend = %q, want empty (backend-agnostic)", entries[0].Backend)
	}
}

func TestGet_Codex(t *testing.T) {
	entries, err := Get("codex")
	if err != nil {
		t.Fatalf("Get(\"codex\") returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Get(\"codex\") returned %d entries, want 1", len(entries))
	}
	if entries[0].Name != "codex" {
		t.Errorf("entries[0].Name = %q, want \"codex\"", entries[0].Name)
	}
	if entries[0].Source != "/var/opt/teeleport/.codex" {
		t.Errorf("entries[0].Source = %q, want \"/var/opt/teeleport/.codex\"", entries[0].Source)
	}
	if entries[0].Target != "~/.codex" {
		t.Errorf("entries[0].Target = %q, want \"~/.codex\"", entries[0].Target)
	}
}

func TestGet_Gemini(t *testing.T) {
	entries, err := Get("gemini")
	if err != nil {
		t.Fatalf("Get(\"gemini\") returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Get(\"gemini\") returned %d entries, want 1", len(entries))
	}
	if entries[0].Name != "gemini" {
		t.Errorf("entries[0].Name = %q, want \"gemini\"", entries[0].Name)
	}
	if entries[0].Source != "/var/opt/teeleport/.gemini" {
		t.Errorf("entries[0].Source = %q, want \"/var/opt/teeleport/.gemini\"", entries[0].Source)
	}
	if entries[0].Target != "~/.gemini" {
		t.Errorf("entries[0].Target = %q, want \"~/.gemini\"", entries[0].Target)
	}
}

func TestGet_Copilot(t *testing.T) {
	entries, err := Get("copilot")
	if err != nil {
		t.Fatalf("Get(\"copilot\") returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Get(\"copilot\") returned %d entries, want 1", len(entries))
	}
	if entries[0].Name != "copilot" {
		t.Errorf("entries[0].Name = %q, want \"copilot\"", entries[0].Name)
	}
	if entries[0].Source != "/var/opt/teeleport/.copilot" {
		t.Errorf("entries[0].Source = %q, want \"/var/opt/teeleport/.copilot\"", entries[0].Source)
	}
	if entries[0].Target != "~/.copilot" {
		t.Errorf("entries[0].Target = %q, want \"~/.copilot\"", entries[0].Target)
	}
}

func TestGet_GH(t *testing.T) {
	entries, err := Get("gh")
	if err != nil {
		t.Fatalf("Get(\"gh\") returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Get(\"gh\") returned %d entries, want 1", len(entries))
	}
	if entries[0].Name != "gh" {
		t.Errorf("entries[0].Name = %q, want \"gh\"", entries[0].Name)
	}
	if entries[0].Source != "/var/opt/teeleport/.config/gh" {
		t.Errorf("entries[0].Source = %q, want \"/var/opt/teeleport/.config/gh\"", entries[0].Source)
	}
	if entries[0].Target != "~/.config/gh" {
		t.Errorf("entries[0].Target = %q, want \"~/.config/gh\"", entries[0].Target)
	}
}

func TestGet_GitConfig(t *testing.T) {
	entries, err := Get("gitconfig")
	if err != nil {
		t.Fatalf("Get(\"gitconfig\") returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Get(\"gitconfig\") returned %d entries, want 1", len(entries))
	}
	if entries[0].Name != "gitconfig" {
		t.Errorf("entries[0].Name = %q, want \"gitconfig\"", entries[0].Name)
	}
	if entries[0].Source != "/var/opt/teeleport/.gitconfig" {
		t.Errorf("entries[0].Source = %q, want \"/var/opt/teeleport/.gitconfig\"", entries[0].Source)
	}
	if entries[0].Target != "~/.gitconfig" {
		t.Errorf("entries[0].Target = %q, want \"~/.gitconfig\"", entries[0].Target)
	}
	if entries[0].Type != "file" {
		t.Errorf("entries[0].Type = %q, want \"file\"", entries[0].Type)
	}
}

func TestGet_Unknown(t *testing.T) {
	_, err := Get("nonexistent")
	if err == nil {
		t.Fatal("Get(\"nonexistent\") should return error")
	}
}
