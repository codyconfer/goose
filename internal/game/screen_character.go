package game

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/goose/internal/characters"
	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/game/viewkit/panels"
	"github.com/codyconfer/goose/internal/game/viewkit/theme"
	"github.com/codyconfer/goose/internal/notify"
)

type characterScreen struct {
	char         *characters.Character
	cursor       int
	notification *notify.Notification
	outcomeTTL   int
	prev         *gameScreen
}

func (cs *characterScreen) simulates() bool { return true }

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
	if msg.String() == "ctrl+c" {
		m.quitting = true
		_ = m.save()
		return tea.Quit
	}

	if cs.notification != nil {
		switch msg.String() {
		case "enter", " ", "spacebar", "esc", "q":
			m.notifs.Push(*cs.notification, outcomeBeats)
			m.screen = cs.prev
		}
		return nil
	}

	switch msg.String() {
	case "up", "k":
		cs.cursor = panels.MoveIndex(cs.cursor, -1, len(cs.char.Options))
	case "down", "j":
		cs.cursor = panels.MoveIndex(cs.cursor, 1, len(cs.char.Options))
	case "enter", " ", "spacebar":
		if len(cs.char.Options) == 0 {
			return nil
		}
		cs.cursor = panels.ClampIndex(cs.cursor, len(cs.char.Options))
		out := cs.char.Options[cs.cursor].Resolve(m.econ.Get(), m.rng)
		m.econ.ApplyWindfall(out.Notif.Title, out.Cmds)
		cs.notification = &out.Notif
		cs.outcomeTTL = characterTimeoutBeats
	}
	return nil
}

func (cs *characterScreen) view(m *Model) string {
	vk := m.frame()
	cursor := panels.ClampIndex(cs.cursor, len(cs.char.Options))
	var b strings.Builder
	b.WriteString(vk.Header(cs.char.Type.Headline()))
	b.WriteString("\n\n")
	b.WriteString(vk.Box(cs.char.Pitch))
	b.WriteString("\n\n")

	if cs.notification != nil {
		b.WriteString(notificationCard(cs.notification.Title, cs.notification.Message, cs.notification.Tone, vk.Width))
		b.WriteString("\n\n")
		b.WriteString(vk.HintLine([2]string{"enter", content.Text.Character.BackHint}))
		return b.String()
	}

	b.WriteString(theme.DimSty.Render(content.Text.Character.Prompt))
	b.WriteString("\n\n")
	for i, opt := range cs.char.Options {
		b.WriteString(vk.Selectable(opt.Label, i == cursor) + "\n")
		if i == cursor {
			b.WriteString(theme.DimSty.Width(vk.Width-5).MarginLeft(5).Render(opt.Desc) + "\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(vk.HintLine(
		[2]string{"↑/↓", "weigh options"},
		[2]string{"enter", "decide"},
	))
	return b.String()
}
