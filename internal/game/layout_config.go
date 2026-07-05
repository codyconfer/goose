package game

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/codyconfer/viewkit/layout"
)

const (
	screenGame  = "game"
	screenTrade = "trade"
)

var configurableScreens = []string{screenGame, screenTrade}

var screenTitles = map[string]string{
	screenGame:  "Home",
	screenTrade: "Trade Desk",
}

type layoutConfig struct {
	Screens map[string]layout.ScreenSpec `json:"screens"`
}

var uiLayout = layoutConfig{Screens: map[string]layout.ScreenSpec{}}

func refs(keys ...string) []layout.PaneRef {
	out := make([]layout.PaneRef, len(keys))
	for i, k := range keys {
		out[i] = layout.PaneRef{Key: k}
	}
	return out
}

func defaultSpec(id string) layout.ScreenSpec {
	switch id {
	case screenGame:
		return layout.ScreenSpec{Layout: "single", Panes: refs("capex", "market", "spectrum", "orders", "feed", "activity")}
	case screenTrade:
		return layout.ScreenSpec{Layout: "flex-columns", Panes: refs("purse", "chart", "flow", "builder", "queue", "book", "ticket", "positions", "pnl", "roster", "ledger")}
	}
	return layout.ScreenSpec{Layout: "single"}
}

func layoutSpec(id string) layout.ScreenSpec {
	if s, ok := uiLayout.Screens[id]; ok {
		return s
	}
	return defaultSpec(id)
}

func setLayoutSpec(id string, spec layout.ScreenSpec) {
	if uiLayout.Screens == nil {
		uiLayout.Screens = map[string]layout.ScreenSpec{}
	}
	uiLayout.Screens[id] = spec
}

func layoutConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "goose_layout.json"
	}
	return filepath.Join(home, ".goose", "layout.json")
}

func loadLayoutConfig() {
	uiLayout = layoutConfig{Screens: map[string]layout.ScreenSpec{}}
	data, err := os.ReadFile(layoutConfigPath())
	if err != nil {
		return
	}
	var cfg layoutConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}
	if cfg.Screens == nil {
		cfg.Screens = map[string]layout.ScreenSpec{}
	}
	uiLayout = cfg
}

func saveLayoutConfig() error {
	path := layoutConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(uiLayout, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func paneRegistryKeys(id string) []layout.PaneInfo {
	switch id {
	case screenGame:
		return gamePanesReg.PaneKeys()
	case screenTrade:
		return tradePanesReg.PaneKeys()
	}
	return nil
}

func layoutKeys(id string) []string {
	switch id {
	case screenGame:
		return gamePanesReg.LayoutKeys()
	case screenTrade:
		return tradePanesReg.LayoutKeys()
	}
	return nil
}
