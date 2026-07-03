package game

import (
	"fmt"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/panels"
	"github.com/codyconfer/viewkit/theme"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
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

	action, ok := menuKeymap().Action(msg.String())
	if !ok {
		return nil
	}
	switch action {
	case keys.Quit:
		m.quitting = true
		return tea.Quit
	case actMenuNew:
		return ms.startNew(m)
	case actMenuRename:
		if save, ok := ms.selectedSave(); ok {
			ms.mode = menuModeRename
			ms.target = save
			ms.edit = save.Name
			m.flash = ""
		}
	case actMenuDelete:
		if save, ok := ms.selectedSave(); ok {
			ms.mode = menuModeDelete
			ms.target = save
			m.flash = ""
		}
	case keys.Up:
		ms.cursor = panels.MoveIndex(ms.cursor, -1, len(ms.items))
	case keys.Down:
		ms.cursor = panels.MoveIndex(ms.cursor, 1, len(ms.items))
	case keys.Confirm:
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
		m.setOffline(offline)
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
	m.setOffline(0)
	m.saveID = 0
	m.saveName = ""
	m.loadPriceChart()
	if err := m.save(); err != nil {
		m.setFlash("Couldn't create the save.")
	}
	m.screen = &gameScreen{}
}

func (ms *menuScreen) handleRename(m *Model, msg tea.KeyMsg) tea.Cmd {
	action, ok := menuRenameKeymap().Action(msg.String())
	if !ok {
		for _, r := range msg.Runes {
			if unicode.IsPrint(r) {
				ms.edit += string(r)
			}
		}
		return nil
	}
	switch action {
	case actMenuSave:
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
	case keys.Cancel:
		ms.mode = menuModeNormal
		ms.edit = ""
	case keys.Erase:
		rs := []rune(ms.edit)
		if len(rs) > 0 {
			ms.edit = string(rs[:len(rs)-1])
		}
	}
	return nil
}

func (ms *menuScreen) handleDelete(m *Model, msg tea.KeyMsg) tea.Cmd {
	action, ok := menuDeleteKeymap().Action(msg.String())
	if !ok {
		return nil
	}
	switch action {
	case actConfirmYes:
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
	case keys.Cancel:
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
			theme.AccentSty.Render(vk.Fit(ms.edit)),
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
		km := menuRenameKeymap()
		b.WriteString(vk.HintLine(
			[2]string{"type", "rename"},
			km.Hint(keys.Erase),
			km.Hint(actMenuSave),
			km.Hint(keys.Cancel),
		))
	case menuModeDelete:
		km := menuDeleteKeymap()
		b.WriteString(vk.HintLine(km.Hint(actConfirmYes), km.Hint(keys.Cancel)))
	default:
		km := menuKeymap()
		b.WriteString(vk.HintLine(
			km.Hint(keys.Up),
			km.Hint(keys.Confirm),
			km.Hint(actMenuNew),
			km.Hint(actMenuRename),
			km.Hint(actMenuDelete),
			km.Hint(keys.Quit),
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
