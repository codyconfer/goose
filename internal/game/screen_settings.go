package game

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/game/viewkit/panels"
	"github.com/codyconfer/goose/internal/game/viewkit/theme"
)

type settingRow struct {
	spec content.SettingSpec
	idx  int
}

type settingsScreen struct {
	rows   []settingRow
	cursor int
}

func newSettingsScreen() *settingsScreen {
	def := economy.DefaultSettings()
	return &settingsScreen{
		rows: []settingRow{
			{spec: content.Settings.LevelPace, idx: def.LevelIdx()},
			{spec: content.Settings.EventPace, idx: def.EventIdx()},
			{spec: content.Settings.MarketPace, idx: def.MarketIdx()},
		},
	}
}

func (ss *settingsScreen) simulates() bool { return false }

func (ss *settingsScreen) settings() economy.Settings {
	set := economy.DefaultSettings()
	if len(ss.rows) > 0 {
		set = set.WithLevel(ss.rows[0].idx)
	}
	if len(ss.rows) > 1 {
		set = set.WithEvent(ss.rows[1].idx)
	}
	if len(ss.rows) > 2 {
		set = set.WithMarket(ss.rows[2].idx)
	}
	return set
}

func (ss *settingsScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	ss.cursor = panels.ClampIndex(ss.cursor, len(ss.rows))
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return tea.Quit
	case "esc", "q":
		m.screen = &menuScreen{items: menuItems(m.saves)}
	case "up", "k":
		ss.cursor = panels.MoveIndex(ss.cursor, -1, len(ss.rows))
	case "down", "j":
		ss.cursor = panels.MoveIndex(ss.cursor, 1, len(ss.rows))
	case "left", "h":
		if len(ss.rows) > 0 {
			r := &ss.rows[ss.cursor]
			r.idx = panels.StepIndex(r.idx, -1, len(r.spec.Options))
		}
	case "right", "l":
		if len(ss.rows) > 0 {
			r := &ss.rows[ss.cursor]
			r.idx = panels.StepIndex(r.idx, 1, len(r.spec.Options))
		}
	case "enter", " ", "spacebar":
		m.foundFlock(ss.settings())
	}
	return nil
}

func (ss *settingsScreen) view(m *Model) string {
	vk := m.frame()
	cursor := panels.ClampIndex(ss.cursor, len(ss.rows))
	var b strings.Builder
	b.WriteString(vk.Header("NEW FLOCK", "tune the simulation before you hatch"))
	b.WriteString("\n\n")

	for i, r := range ss.rows {
		selected := i == cursor
		idx := panels.ClampIndex(r.idx, len(r.spec.Options))
		var label string
		if len(r.spec.Options) > 0 {
			label = r.spec.Options[idx].Label
		}

		valSty := theme.ValSty
		if selected {
			valSty = theme.EggSty
		}
		left, right := "  ", "  "
		if selected {
			if idx > 0 {
				left = theme.KeySty.Render("◂ ")
			}
			if idx < len(r.spec.Options)-1 {
				right = theme.KeySty.Render(" ▸")
			}
		}
		value := left + valSty.Render(label) + right

		b.WriteString(vk.Spread(vk.Selectable(r.spec.Label, selected), value))
		b.WriteString("\n")
		if selected {
			b.WriteString(theme.DimSty.Width(vk.Width-5).MarginLeft(5).Render(r.spec.Desc) + "\n")
		}
		b.WriteString("\n")
	}

	if m.flash != "" {
		b.WriteString(panels.Flash(vk.Fit(m.flash)) + "\n\n")
	}
	b.WriteString(vk.HintLine(
		[2]string{"↑/↓", "setting"},
		[2]string{"←/→", "change"},
		[2]string{"enter", "hatch flock"},
		[2]string{"esc", "back"},
	))
	return b.String()
}
