package game

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/codyconfer/viewkit/theme"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
)

func TestSettingsThemePickerSwitchesAndPersists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	defer theme.Use(theme.Default())

	m := New(economy.NewMachine(), events.NewMachine(), 0)
	ss := newSettingsScreen()
	m.screen = ss

	if !strings.Contains(renderForHints(m), "Theme") {
		t.Fatalf("settings view missing Theme row:\n%s", renderForHints(m))
	}

	ss.cursor = ss.themePos()
	startAccent := theme.Cur().Accent.GetForeground()
	ss.handleKey(&m, tea.KeyMsg{Type: tea.KeyRight})

	if theme.Cur().Accent.GetForeground() == startAccent {
		t.Fatal("stepping theme did not change the active theme")
	}
	wantKey := theme.Keys()[1]
	if theme.Cur().Accent.GetForeground() != mustAccent(t, wantKey) {
		t.Fatalf("active theme = %v, want %s", theme.Cur().Accent.GetForeground(), wantKey)
	}

	data, err := os.ReadFile(filepath.Join(home, ".goose", "theme.json"))
	if err != nil {
		t.Fatalf("theme.json not written: %v", err)
	}
	if !strings.Contains(string(data), wantKey) {
		t.Fatalf("theme.json missing %q: %s", wantKey, data)
	}

	theme.Use(theme.Default())
	loadThemeConfig()
	if theme.Cur().Accent.GetForeground() != mustAccent(t, wantKey) {
		t.Fatalf("loadThemeConfig did not restore %s", wantKey)
	}
}

func TestViewAppliesThemeBackground(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prev)
	defer theme.Use(theme.Default())

	m := New(economy.NewMachine(), events.NewMachine(), 0)
	th, _ := theme.Named("solarized-light")
	theme.Use(th)
	m.screen = newSettingsScreen()

	if !strings.Contains(renderForHints(m), "48;2;253;246;227") {
		t.Fatal("View() missing solarized-light background fill")
	}
}

func mustAccent(t *testing.T, key string) interface{} {
	t.Helper()
	th, ok := theme.Named(key)
	if !ok {
		t.Fatalf("theme %q not found", key)
	}
	return th.Accent.GetForeground()
}
