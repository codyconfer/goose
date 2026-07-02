package game

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

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
	seed   string
}

func newSettingsScreen() *settingsScreen {
	def := economy.DefaultSettings()
	return &settingsScreen{
		rows: []settingRow{
			{spec: content.Settings.LevelPace, idx: def.LevelIdx()},
			{spec: content.Settings.EventPace, idx: def.EventIdx()},
			{spec: content.Settings.MarketPace, idx: def.MarketIdx()},
		},
		seed: fmt.Sprintf("%d", time.Now().UnixNano()),
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

func (ss *settingsScreen) seedValue() int64 {
	if strings.TrimSpace(ss.seed) == "" {
		return time.Now().UnixNano()
	}
	seed, err := strconv.ParseInt(ss.seed, 10, 64)
	if err != nil {
		return time.Now().UnixNano()
	}
	return seed
}

func (ss *settingsScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	total := len(ss.rows) + 1
	ss.cursor = panels.ClampIndex(ss.cursor, total)
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return tea.Quit
	case "esc", "q":
		m.screen = &menuScreen{items: menuItems(m.saves)}
	case "up", "k":
		ss.cursor = panels.MoveIndex(ss.cursor, -1, total)
	case "down", "j":
		ss.cursor = panels.MoveIndex(ss.cursor, 1, total)
	case "left", "h":
		if ss.cursor < len(ss.rows) {
			r := &ss.rows[ss.cursor]
			r.idx = panels.StepIndex(r.idx, -1, len(r.spec.Options))
		}
	case "right", "l":
		if ss.cursor < len(ss.rows) {
			r := &ss.rows[ss.cursor]
			r.idx = panels.StepIndex(r.idx, 1, len(r.spec.Options))
		}
	case "r":
		if ss.cursor == len(ss.rows) {
			ss.seed = fmt.Sprintf("%d", time.Now().UnixNano())
		}
	case "backspace", "ctrl+h":
		if ss.cursor == len(ss.rows) {
			rs := []rune(ss.seed)
			if len(rs) > 0 {
				ss.seed = string(rs[:len(rs)-1])
			}
		}
	case "enter", " ", "spacebar":
		m.foundFlock(ss.settings(), ss.seedValue())
	default:
		if ss.cursor == len(ss.rows) {
			for _, r := range msg.Runes {
				if unicode.IsDigit(r) || (r == '-' && ss.seed == "") {
					ss.seed += string(r)
				}
			}
		}
	}
	return nil
}

func (ss *settingsScreen) view(m *Model) string {
	vk := m.frame()
	cursor := panels.ClampIndex(ss.cursor, len(ss.rows)+1)
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

	seedSelected := cursor == len(ss.rows)
	seedValue := theme.ValSty.Render(ss.seed)
	if seedSelected {
		seedValue = theme.EggSty.Render(ss.seed)
	}
	b.WriteString(vk.Spread(vk.Selectable("World Seed", seedSelected), seedValue))
	b.WriteString("\n")
	if seedSelected {
		b.WriteString(theme.DimSty.Width(vk.Width-5).MarginLeft(5).Render("Type a deterministic seed or press r to reroll the generated world.") + "\n")
	}
	b.WriteString("\n")

	if m.flash != "" {
		b.WriteString(panels.Flash(vk.Fit(m.flash)) + "\n\n")
	}
	hints := [][2]string{verticalHint("setting")}
	if seedSelected {
		hints = append(hints,
			hint("digits/-", "seed"),
			hint("backspace", "erase"),
			hint("r", "reroll seed"),
		)
	} else {
		hints = append(hints, horizontalHint("change"))
	}
	hints = append(hints, confirmHint("hatch flock"))
	hints = append(hints, m.pageHintPairs()...)
	hints = append(hints, hint("esc/q", "back"))
	b.WriteString(vk.HintLine(hints...))
	return b.String()
}
