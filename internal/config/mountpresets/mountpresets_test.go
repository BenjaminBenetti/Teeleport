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
	// Verify no backend is set (presets are backend-agnostic)
	if entries[0].Backend != "" {
		t.Errorf("entries[0].Backend = %q, want empty (backend-agnostic)", entries[0].Backend)
	}
}

func TestGet_Unknown(t *testing.T) {
	_, err := Get("nonexistent")
	if err == nil {
		t.Fatal("Get(\"nonexistent\") should return error")
	}
}
