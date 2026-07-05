package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadClockConfigCustomZone(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	defer loadClockConfig()

	dir := filepath.Join(home, ".goose")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "clock.json"),
		[]byte(`{"timezone":"Asia/Tokyo","label":"TOKYO"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	loadClockConfig()
	label, loc := clockZone()
	if label != "TOKYO" {
		t.Errorf("label = %q, want TOKYO", label)
	}
	if loc.String() != "Asia/Tokyo" {
		t.Errorf("loc = %q, want Asia/Tokyo", loc)
	}
}

func TestLoadClockConfigDerivesLabel(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	defer loadClockConfig()

	dir := filepath.Join(home, ".goose")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "clock.json"),
		[]byte(`{"timezone":"America/New_York"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	loadClockConfig()
	if label, _ := clockZone(); label != "New York" {
		t.Errorf("derived label = %q, want New York", label)
	}
}

func TestLoadClockConfigFallsBackOnBadZone(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	defer loadClockConfig()

	dir := filepath.Join(home, ".goose")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "clock.json"),
		[]byte(`{"timezone":"Not/AZone"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	loadClockConfig()
	label, loc := clockZone()
	if loc.String() != "America/New_York" || label != "NYC" {
		t.Errorf("bad zone should fall back to default, got %q / %q", label, loc)
	}
}
