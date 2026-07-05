package game

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/codyconfer/viewkit/theme"
)

type themeConfig struct {
	Theme string `json:"theme"`
}

var uiTheme = themeConfig{Theme: theme.Keys()[0]}

func themeConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "goose_theme.json"
	}
	return filepath.Join(home, ".goose", "theme.json")
}

func loadThemeConfig() {
	uiTheme = themeConfig{Theme: theme.Keys()[0]}
	data, err := os.ReadFile(themeConfigPath())
	if err == nil {
		var cfg themeConfig
		if err := json.Unmarshal(data, &cfg); err == nil && cfg.Theme != "" {
			uiTheme = cfg
		}
	}
	if t, ok := theme.Named(uiTheme.Theme); ok {
		theme.Use(t)
	} else {
		uiTheme.Theme = theme.Keys()[0]
		theme.Use(theme.Default())
	}
}

func setTheme(key string) {
	uiTheme.Theme = key
	if t, ok := theme.Named(key); ok {
		theme.Use(t)
	}
}

func saveThemeConfig() error {
	path := themeConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(uiTheme, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
