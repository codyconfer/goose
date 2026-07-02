package game

import (
	"fmt"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
	"github.com/codyconfer/goose/internal/game/viewkit/panels"
	"github.com/codyconfer/goose/internal/game/viewkit/theme"
	"github.com/codyconfer/goose/internal/store"
	"github.com/codyconfer/goose/internal/world"
)

type menuAction int

const (
	menuNew menuAction = iota
	menuSave
	menuExit
)

type menuItem struct {
	action menuAction
	save   store.SaveInfo
}

type menuMode int

const (
	menuModeNormal menuMode = iota
	menuModeRename
	menuModeDelete
)

type menuScreen struct {
	items  []menuItem
	cursor int
	mode   menuMode
	edit   string
	target store.SaveInfo
}

func (ms *menuScreen) simulates() bool { return false }

func (ms *menuScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	switch ms.mode {
	case menuModeRename:
		return ms.handleRename(m, msg)
	case menuModeDelete:
		return ms.handleDelete(m, msg)
	}

	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.quitting = true
		return tea.Quit
	case "n":
		return ms.startNew(m)
	case "r":
		if save, ok := ms.selectedSave(); ok {
			ms.mode = menuModeRename
			ms.target = save
			ms.edit = save.Name
			m.flash = ""
		}
	case "x", "d":
		if save, ok := ms.selectedSave(); ok {
			ms.mode = menuModeDelete
			ms.target = save
			m.flash = ""
		}
	case "up", "k":
		ms.cursor = panels.MoveIndex(ms.cursor, -1, len(ms.items))
	case "down", "j":
		ms.cursor = panels.MoveIndex(ms.cursor, 1, len(ms.items))
	case "enter", " ", "spacebar":
		return ms.choose(m)
	}
	return nil
}

func (ms *menuScreen) choose(m *Model) tea.Cmd {
	if len(ms.items) == 0 {
		return nil
	}
	ms.cursor = panels.ClampIndex(ms.cursor, len(ms.items))
	switch it := ms.items[ms.cursor]; it.action {
	case menuNew:
		return ms.startNew(m)
	case menuSave:
		econ, ev, wrld, offline, err := store.Load(it.save.ID)
		if err != nil {
			m.setFlash(content.Text.Menu.SaveError)
			m.refreshSaves(ms)
			return nil
		}
		m.econ = econ
		m.events = ev
		m.world = wrld
		m.offline = offline
		m.saveID = it.save.ID
		m.saveName = it.save.Name
		m.loadPriceChart()
		m.screen = &gameScreen{}
	case menuExit:
		m.quitting = true
		return tea.Quit
	}
	return nil
}

func (ms *menuScreen) startNew(m *Model) tea.Cmd {
	m.flash = ""
	m.screen = newSettingsScreen()
	return nil
}

func (m *Model) foundFlock(set economy.Settings, seed int64) {
	m.econ = economy.NewMachine()
	m.econ.SetSettings(set)
	m.events = events.NewMachine()
	m.world = world.Generate(seed)
	m.offline = 0
	m.saveID = 0
	m.saveName = ""
	m.loadPriceChart()
	if err := m.save(); err != nil {
		m.setFlash("Couldn't create the save.")
	}
	m.screen = &gameScreen{}
}

func (ms *menuScreen) handleRename(m *Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		name := store.CleanName(ms.edit)
		if name == "" {
			m.setFlash("Save name can't be empty.")
			return nil
		}
		if err := store.RenameSave(ms.target.ID, name); err != nil {
			m.setFlash("Couldn't rename the save.")
			return nil
		}
		if m.saveID == ms.target.ID {
			m.saveName = name
		}
		id := ms.target.ID
		ms.mode = menuModeNormal
		ms.edit = ""
		m.refreshSaves(ms)
		ms.selectSave(id)
	case "esc", "ctrl+c":
		ms.mode = menuModeNormal
		ms.edit = ""
	case "backspace", "ctrl+h":
		rs := []rune(ms.edit)
		if len(rs) > 0 {
			ms.edit = string(rs[:len(rs)-1])
		}
	default:
		for _, r := range msg.Runes {
			if unicode.IsPrint(r) {
				ms.edit += string(r)
			}
		}
	}
	return nil
}

func (ms *menuScreen) handleDelete(m *Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y", "Y":
		id := ms.target.ID
		if err := store.DeleteSave(id); err != nil {
			m.setFlash("Couldn't delete the save.")
			return nil
		}
		if m.saveID == id {
			m.saveID = 0
			m.saveName = ""
		}
		ms.mode = menuModeNormal
		m.refreshSaves(ms)
	case "n", "N", "esc", "q", "ctrl+c":
		ms.mode = menuModeNormal
	}
	return nil
}

func (ms *menuScreen) view(m *Model) string {
	vk := m.frame()
	cursor := panels.ClampIndex(ms.cursor, len(ms.items))
	var b strings.Builder
	b.WriteString(vk.Header(content.Text.App.Title, content.Text.App.Tagline))
	b.WriteString("\n\n")

	switch ms.mode {
	case menuModeRename:
		b.WriteString(vk.Panel("SAVE NAME",
			theme.DimSty.Render("Rename "+ms.target.Name),
			theme.EggSty.Render(vk.Fit(ms.edit)),
		))
		b.WriteString("\n\n")
	case menuModeDelete:
		b.WriteString(vk.Panel("DELETE SAVE",
			theme.CantSty.Render(vk.Fit("Delete "+ms.target.Name+"?")),
		))
		b.WriteString("\n\n")
	}

	for i, it := range ms.items {
		b.WriteString(vk.Selectable(menuLabel(it), i == cursor) + "\n")
	}
	if m.flash != "" {
		b.WriteString("\n")
		b.WriteString(panels.Flash(vk.Fit(m.flash)))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	switch ms.mode {
	case menuModeRename:
		b.WriteString(vk.HintLine(
			[2]string{"enter", "save"},
			[2]string{"esc", "cancel"},
		))
	case menuModeDelete:
		b.WriteString(vk.HintLine(
			[2]string{"y", "delete"},
			[2]string{"n", "cancel"},
		))
	default:
		b.WriteString(vk.HintLine(
			[2]string{"↑/↓", "select"},
			[2]string{"enter", "choose"},
			[2]string{"n", "new"},
			[2]string{"r", "rename"},
			[2]string{"x", "delete"},
			[2]string{"q", "quit"},
		))
	}
	return b.String()
}

func menuItems(saves []store.SaveInfo) []menuItem {
	items := make([]menuItem, 0, len(saves)+2)
	for _, save := range saves {
		items = append(items, menuItem{action: menuSave, save: save})
	}
	items = append(items, menuItem{action: menuNew})
	items = append(items, menuItem{action: menuExit})
	return items
}

func (ms *menuScreen) selectedSave() (store.SaveInfo, bool) {
	if ms.cursor < 0 || ms.cursor >= len(ms.items) {
		return store.SaveInfo{}, false
	}
	it := ms.items[ms.cursor]
	if it.action != menuSave {
		return store.SaveInfo{}, false
	}
	return it.save, true
}

func (ms *menuScreen) selectSave(id int64) {
	for i, it := range ms.items {
		if it.action == menuSave && it.save.ID == id {
			ms.cursor = i
			return
		}
	}
}

func menuLabel(it menuItem) string {
	switch it.action {
	case menuNew:
		return content.Text.Menu.NewGame
	case menuSave:
		return saveLabel(it.save)
	case menuExit:
		return content.Text.Menu.Exit
	}
	return ""
}

func saveLabel(save store.SaveInfo) string {
	name := save.Name
	if name == "" {
		name = fmt.Sprintf("Save %d", save.ID)
	}
	return fmt.Sprintf("%s  ·  Lv.%d  ·  %s 🪙", name, save.Level, economy.FormatNum(save.Tokens))
}
