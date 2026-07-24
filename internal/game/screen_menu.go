package game

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/forms"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/list"
	"github.com/codyconfer/viewkit/panels"

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

type menuScreen struct {
	items   []menuItem
	list    list.Model
	ready   bool
	rename  *forms.Form
	confirm *forms.Confirm
	target  store.SaveInfo
}

func (ms *menuScreen) simulates() bool { return false }

func menuItemKey(it menuItem) string {
	switch it.action {
	case menuNew:
		return "new"
	case menuExit:
		return "exit"
	case menuSave:
		return fmt.Sprintf("save:%d", it.save.ID)
	}
	return ""
}

func (ms *menuScreen) ensure() {
	if ms.ready {
		return
	}
	ms.list = list.New()
	ms.list.SetFocused(true)
	ms.syncItems("")
	ms.ready = true
}

func (ms *menuScreen) syncItems(sel string) {
	items := make([]list.Item, len(ms.items))
	for i, it := range ms.items {
		items[i] = list.Item{Block: menuLabel(it), Key: menuItemKey(it), Selectable: true}
	}
	ms.list.SetItems(items)
	if sel != "" {
		selectListKey(&ms.list, sel)
	}
}

func (ms *menuScreen) selected() (menuItem, bool) {
	it, ok := ms.list.Selected()
	if !ok {
		return menuItem{}, false
	}
	for _, mi := range ms.items {
		if menuItemKey(mi) == it.Key {
			return mi, true
		}
	}
	return menuItem{}, false
}

func (ms *menuScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	ms.ensure()
	if ms.rename != nil {
		return ms.handleRename(m, msg)
	}
	if ms.confirm != nil {
		return ms.handleConfirm(m, msg)
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
			ms.rename = forms.NewForm(forms.Field{
				Key:   "name",
				Label: "Rename " + save.Name,
				Kind:  forms.FieldText,
				Text:  save.Name,
			})
			ms.target = save
			m.flash = ""
		}
	case actMenuDelete:
		if save, ok := ms.selectedSave(); ok {
			ms.confirm = &forms.Confirm{
				Title:    "DELETE SAVE",
				Message:  "Delete " + save.Name + "?",
				YesLabel: "delete",
				NoLabel:  "cancel",
				Yes:      true,
			}
			ms.target = save
			m.flash = ""
		}
	case actMenuLayout:
		m.flash = ""
		m.screen = newLayoutEditor(ms)
	case keys.Up:
		ms.list.Move(-1)
	case keys.Down:
		ms.list.Move(1)
	case keys.Confirm:
		return ms.choose(m)
	}
	return nil
}

func (ms *menuScreen) choose(m *Model) tea.Cmd {
	it, ok := ms.selected()
	if !ok {
		return nil
	}
	switch it.action {
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
		ms.rename.Insert(string(msg.Runes))
		return nil
	}
	switch action {
	case actMenuSave:
		name := store.CleanName(ms.rename.Focused().Text)
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
		ms.rename = nil
		m.refreshSaves(ms)
	case keys.Cancel:
		ms.rename = nil
	case keys.Erase:
		ms.rename.Handle(keys.Erase)
	}
	return nil
}

func (ms *menuScreen) handleConfirm(m *Model, msg tea.KeyMsg) tea.Cmd {
	action, ok := menuDeleteKeymap().Action(msg.String())
	if !ok {
		return nil
	}
	switch ms.confirm.Handle(action) {
	case forms.Submitted:
		yes := ms.confirm.Yes
		id := ms.target.ID
		ms.confirm = nil
		if !yes {
			return nil
		}
		if err := store.DeleteSave(id); err != nil {
			m.setFlash("Couldn't delete the save.")
			return nil
		}
		if m.saveID == id {
			m.saveID = 0
			m.saveName = ""
		}
		m.refreshSaves(ms)
	case forms.Cancelled:
		ms.confirm = nil
	}
	return nil
}

func (ms *menuScreen) view(m *Model) string {
	ms.ensure()
	vk := m.frame()
	ms.list.SetSize(vk.Width, 0)
	var b strings.Builder
	b.WriteString(vk.Header(content.Text.App.Title, content.Text.App.Tagline))
	b.WriteString("\n\n")

	switch {
	case ms.rename != nil:
		b.WriteString(ms.rename.Render(vk, "SAVE NAME"))
		b.WriteString("\n\n")
	case ms.confirm != nil:
		b.WriteString(ms.confirm.Render(vk))
		b.WriteString("\n\n")
	}

	b.WriteString(ms.list.View())
	b.WriteString("\n")
	if m.flash != "" {
		b.WriteString("\n")
		b.WriteString(panels.Flash(vk.Fit(m.flash)))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	switch {
	case ms.rename != nil:
		km := menuRenameKeymap()
		b.WriteString(vk.HintLine(
			[2]string{"type", "rename"},
			km.Hint(keys.Erase),
			km.Hint(actMenuSave),
			km.Hint(keys.Cancel),
		))
	case ms.confirm != nil:
		km := menuDeleteKeymap()
		b.WriteString(vk.HintLine(km.Hint(keys.Confirm), km.Hint(keys.Cancel)))
	default:
		km := menuKeymap()
		b.WriteString(vk.HintLine(
			km.Hint(keys.Up),
			km.Hint(keys.Confirm),
			km.Hint(actMenuNew),
			km.Hint(actMenuRename),
			km.Hint(actMenuDelete),
			km.Hint(actMenuLayout),
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
	it, ok := ms.selected()
	if !ok || it.action != menuSave {
		return store.SaveInfo{}, false
	}
	return it.save, true
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
