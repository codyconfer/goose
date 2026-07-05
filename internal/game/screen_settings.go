package game

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/forms"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/panels"
	"github.com/codyconfer/viewkit/theme"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
)

type settingRow struct {
	spec content.SettingSpec
	idx  int
}

type settingsScreen struct {
	rows      []settingRow
	cursor    int
	seed      *forms.Form
	themeKeys []string
	themeIdx  int
}

func newSettingsScreen() *settingsScreen {
	def := economy.DefaultSettings()
	tk := theme.Keys()
	return &settingsScreen{
		rows: []settingRow{
			{spec: content.Settings.LevelPace, idx: def.LevelIdx()},
			{spec: content.Settings.EventPace, idx: def.EventIdx()},
			{spec: content.Settings.MarketPace, idx: def.MarketIdx()},
		},
		seed: forms.NewForm(forms.Field{
			Key:  "seed",
			Kind: forms.FieldText,
			Text: fmt.Sprintf("%d", time.Now().UnixNano()),
		}),
		themeKeys: tk,
		themeIdx:  currentThemeIndex(tk),
	}
}

func currentThemeIndex(keys []string) int {
	for i, k := range keys {
		if k == uiTheme.Theme {
			return i
		}
	}
	return 0
}

func (ss *settingsScreen) themePos() int { return len(ss.rows) }
func (ss *settingsScreen) seedPos() int  { return len(ss.rows) + 1 }
func (ss *settingsScreen) rowTotal() int { return len(ss.rows) + 2 }

func (ss *settingsScreen) stepTheme(delta int) {
	if len(ss.themeKeys) == 0 {
		return
	}
	ss.themeIdx = panels.StepIndex(ss.themeIdx, delta, len(ss.themeKeys))
	setTheme(ss.themeKeys[ss.themeIdx])
	_ = saveThemeConfig()
}

func (ss *settingsScreen) seedForm() *forms.Form {
	if ss.seed == nil {
		ss.seed = forms.NewForm(forms.Field{Key: "seed", Kind: forms.FieldText})
	}
	return ss.seed
}

func (ss *settingsScreen) seedText() string { return ss.seedForm().Focused().Text }

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
	txt := strings.TrimSpace(ss.seedText())
	if txt == "" {
		return time.Now().UnixNano()
	}
	seed, err := strconv.ParseInt(txt, 10, 64)
	if err != nil {
		return time.Now().UnixNano()
	}
	return seed
}

func (ss *settingsScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	total := ss.rowTotal()
	ss.cursor = panels.ClampIndex(ss.cursor, total)
	action, ok := settingsKeymap().Action(msg.String())
	if !ok {
		if ss.cursor == ss.seedPos() {
			ss.seedForm().Insert(string(msg.Runes))
		}
		return nil
	}
	switch action {
	case keys.Quit:
		m.quitting = true
		return tea.Quit
	case keys.Cancel:
		m.screen = &menuScreen{items: menuItems(m.saves)}
	case keys.Up:
		ss.cursor = panels.MoveIndex(ss.cursor, -1, total)
	case keys.Down:
		ss.cursor = panels.MoveIndex(ss.cursor, 1, total)
	case keys.Left:
		if ss.cursor < len(ss.rows) {
			r := &ss.rows[ss.cursor]
			r.idx = panels.StepIndex(r.idx, -1, len(r.spec.Options))
		} else if ss.cursor == ss.themePos() {
			ss.stepTheme(-1)
		}
	case keys.Right:
		if ss.cursor < len(ss.rows) {
			r := &ss.rows[ss.cursor]
			r.idx = panels.StepIndex(r.idx, 1, len(r.spec.Options))
		} else if ss.cursor == ss.themePos() {
			ss.stepTheme(1)
		}
	case actReroll:
		if ss.cursor == ss.seedPos() {
			ss.seedForm().Focused().Text = fmt.Sprintf("%d", time.Now().UnixNano())
		}
	case keys.Erase:
		if ss.cursor == ss.seedPos() {
			ss.seedForm().Handle(keys.Erase)
		}
	case keys.Confirm:
		m.foundFlock(ss.settings(), ss.seedValue())
	}
	return nil
}

func (ss *settingsScreen) view(m *Model) string {
	vk := m.frame()
	cursor := panels.ClampIndex(ss.cursor, ss.rowTotal())
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
			valSty = theme.AccentSty
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

	themeSelected := cursor == ss.themePos()
	var themeLabel string
	if len(ss.themeKeys) > 0 {
		tidx := panels.ClampIndex(ss.themeIdx, len(ss.themeKeys))
		themeLabel = theme.DisplayName(ss.themeKeys[tidx])
	}
	themeValSty := theme.ValSty
	if themeSelected {
		themeValSty = theme.AccentSty
	}
	tleft, tright := "  ", "  "
	if themeSelected {
		if ss.themeIdx > 0 {
			tleft = theme.KeySty.Render("◂ ")
		}
		if ss.themeIdx < len(ss.themeKeys)-1 {
			tright = theme.KeySty.Render(" ▸")
		}
	}
	b.WriteString(vk.Spread(vk.Selectable("Theme", themeSelected), tleft+themeValSty.Render(themeLabel)+tright))
	b.WriteString("\n")
	if themeSelected {
		b.WriteString(theme.DimSty.Width(vk.Width-5).MarginLeft(5).Render("Color theme for the whole interface. Applies immediately.") + "\n")
	}
	b.WriteString("\n")

	seedSelected := cursor == ss.seedPos()
	seedTxt := ss.seedText()
	seedValue := theme.ValSty.Render(seedTxt)
	if seedSelected {
		seedValue = theme.AccentSty.Render(seedTxt)
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
	km := settingsKeymap()
	hints := [][2]string{km.Hint(keys.Up)}
	if seedSelected {
		hints = append(hints,
			[2]string{"digits/-", "seed"},
			km.Hint(keys.Erase),
			km.Hint(actReroll),
		)
	} else {
		hints = append(hints, km.Hint(keys.Left))
	}
	hints = append(hints, km.Hint(keys.Confirm))
	hints = append(hints, km.Hint(keys.Cancel))
	b.WriteString(vk.HintLine(hints...))
	return b.String()
}
