package game

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/list"
	"github.com/codyconfer/viewkit/notify"
	"github.com/codyconfer/viewkit/panels"
	"github.com/codyconfer/viewkit/theme"

	"github.com/codyconfer/goose/internal/characters"
	"github.com/codyconfer/goose/internal/content"
)

type characterScreen struct {
	char         *characters.Character
	list         list.Model
	ready        bool
	notification *notify.Notification
	outcomeTTL   int
	prev         *gameScreen
}

func (cs *characterScreen) simulates() bool { return true }

func (cs *characterScreen) ensure() {
	if cs.ready {
		return
	}
	cs.list = list.New()
	cs.list.SetFocused(true)
	items := make([]list.Item, len(cs.char.Options))
	for i, opt := range cs.char.Options {
		items[i] = list.Item{Block: opt.Label, Key: strconv.Itoa(i), Selectable: true}
	}
	cs.list.SetItems(items)
	cs.ready = true
}

func (cs *characterScreen) selectedIndex() (int, bool) {
	it, ok := cs.list.Selected()
	if !ok {
		return 0, false
	}
	idx, err := strconv.Atoi(it.Key)
	if err != nil || idx < 0 || idx >= len(cs.char.Options) {
		return 0, false
	}
	return idx, true
}

func (cs *characterScreen) tick(m *Model) {
	if cs.notification == nil || cs.outcomeTTL <= 0 {
		return
	}
	if cs.outcomeTTL--; cs.outcomeTTL == 0 {
		m.notifs.Push(*cs.notification, outcomeBeats)
		m.screen = cs.prev
	}
}

func (cs *characterScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	cs.ensure()
	if cs.notification != nil {
		action, ok := characterNotifyKeymap().Action(msg.String())
		if !ok {
			return nil
		}
		switch action {
		case keys.Quit:
			m.quitting = true
			_ = m.save()
			return tea.Quit
		case keys.Confirm:
			m.notifs.Push(*cs.notification, outcomeBeats)
			m.screen = cs.prev
		}
		return nil
	}

	action, ok := characterKeymap().Action(msg.String())
	if !ok {
		return nil
	}
	switch action {
	case keys.Quit:
		m.quitting = true
		_ = m.save()
		return tea.Quit
	case keys.Up:
		cs.list.Move(-1)
	case keys.Down:
		cs.list.Move(1)
	case keys.Confirm:
		idx, ok := cs.selectedIndex()
		if !ok {
			return nil
		}
		out := cs.char.Options[idx].Resolve(m.econ.Get(), m.rng)
		m.econ.ApplyWindfall(out.Notif.Title, out.Cmds)
		cs.notification = &out.Notif
		cs.outcomeTTL = characterTimeoutBeats
	}
	return nil
}

func (cs *characterScreen) view(m *Model) string {
	cs.ensure()
	vk := m.frame()
	cs.list.SetSize(vk.Width, 0)
	var b strings.Builder
	b.WriteString(vk.Header(cs.char.Headline))
	b.WriteString("\n\n")
	b.WriteString(vk.Box(cs.char.Pitch))
	b.WriteString("\n\n")

	if cs.notification != nil {
		b.WriteString(panels.NotificationCard(vk, *cs.notification))
		b.WriteString("\n\n")
		km := characterNotifyKeymap()
		b.WriteString(vk.HintLine(km.Hint(keys.Confirm)))
		return b.String()
	}

	b.WriteString(theme.DimSty.Render(content.Text.Character.Prompt))
	b.WriteString("\n\n")
	b.WriteString(cs.list.View())
	b.WriteString("\n")
	if idx, ok := cs.selectedIndex(); ok {
		b.WriteString(theme.DimSty.Width(vk.Width-5).MarginLeft(5).Render(cs.char.Options[idx].Desc) + "\n")
	}
	b.WriteString("\n")
	km := characterKeymap()
	b.WriteString(vk.HintLine(km.Hint(keys.Up), km.Hint(keys.Confirm)))
	return b.String()
}
