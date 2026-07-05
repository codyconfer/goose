package game

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
)

type gameScreen struct {
	cursor     int
	capex      layout.ScrollState
	feedScroll layout.ScrollState
	focus      int
}

func (gs *gameScreen) simulates() bool { return true }

func (gs *gameScreen) build(m *Model) layout.Screen {
	return buildScreen(screenGame, gamePaneCtx{m: m, gs: gs}, gamePanesReg)
}

func (gs *gameScreen) focusables(m *Model) layout.Ring {
	return gs.build(m).Ring()
}

func (gs *gameScreen) focusedPanel(m *Model) string {
	return gs.focusables(m).At(gs.focus)
}

func (gs *gameScreen) focusMove(m *Model, delta int) {
	switch gs.focusedPanel(m) {
	case "capex":
		if delta < 0 {
			gs.cursor = m.prevVisible(gs.cursor)
		} else {
			gs.cursor = m.nextVisible(gs.cursor)
		}
	case "feed":
		gs.feedScroll.Scroll(delta, m.feed.size(), m.panelRows(feedRows))
	}
}

func (gs *gameScreen) focusVerb(m *Model) string {
	if gs.focusedPanel(m) == "feed" {
		return "scroll feed"
	}
	return "select"
}

func (gs *gameScreen) keys() *keys.Map { return gameKeymap() }

func (gs *gameScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	action, ok := gs.keys().Action(msg.String())
	if !ok {
		return nil
	}
	switch action {
	case keys.Quit:
		m.quitting = true
		_ = m.save()
		return tea.Quit
	case keys.Confirm:
		if m.econ.Get().Frozen() {
			m.feed.push("🚫 Shut down — the subcommittee says the geese may not honk right now.")
		} else {
			m.econ.Tap()
			m.pulse = 1
		}
	case keys.Up:
		gs.focusMove(m, -1)
	case keys.Down:
		gs.focusMove(m, 1)
	case keys.FocusNext:
		gs.focus = gs.focusables(m).Step(gs.focus, 1)
	case keys.FocusPrev:
		gs.focus = gs.focusables(m).Step(gs.focus, -1)
	case actBuy:
		gs.buy(m)
	case actSell:
		gs.sell(m)
	case actMaxBuy:
		gs.queueMaxTrade(m, economy.TxBuyEggs)
	case actMaxSell:
		gs.queueMaxTrade(m, economy.TxSellEggs)
	case actMaxCall:
		gs.openMaxPosition(m, economy.PosCall)
	case actMaxPut:
		gs.openMaxPosition(m, economy.PosPut)
	case actOpenTrade:
		m.screen = &tradeScreen{prev: gs, kind: economy.TxBuyEggs}
	case actOpenLayout:
		m.screen = newLayoutEditor(gs)
	}
	return nil
}

func (gs *gameScreen) buy(m *Model) {
	if gs.cursor < 0 || gs.cursor >= len(m.items) {
		return
	}
	it := m.items[gs.cursor]
	if it.buy(m.econ) {
		m.feed.push(it.boughtMsg(m.econ.Get()))
	} else {
		m.feed.push(it.deniedMsg())
	}
}

func (gs *gameScreen) sell(m *Model) {
	if gs.cursor < 0 || gs.cursor >= len(m.items) {
		return
	}
	it := m.items[gs.cursor]
	if it.sell(m.econ) {
		m.feed.push(it.soldMsg(m.econ.Get()))
	} else {
		m.feed.push(it.cantSellMsg())
	}
}

func (gs *gameScreen) queueMaxTrade(m *Model, kind economy.TxKind) {
	ts := tradeScreen{kind: kind, sizeIdx: len(tradeSizes)}
	ts.schedule(m)
}

func (gs *gameScreen) openMaxPosition(m *Model, kind economy.PosKind) {
	if m.econ.Get().Level() < economy.SpecUnlockLevel {
		m.feed.push(fmt.Sprintf(content.Text.Trade.SpecLockedFmt, economy.SpecUnlockLevel))
		return
	}
	premiumIdx := maxAffordableSpecPremiumIdx(m.econ.Get().Tokens)
	if premiumIdx < 0 {
		premiumIdx = 0
	}
	ts := tradeScreen{
		pos:         kind,
		premiumIdx:  premiumIdx,
		leverageIdx: len(specLeverages) - 1,
	}
	ts.open(m)
}

func (gs *gameScreen) view(m *Model) string {
	sections := []string{m.renderTitleBar(), m.renderStatus()}
	if m.econ.Get().Frozen() {
		sections = append(sections, m.renderShutdown())
	}
	sections = append(sections, gs.build(m).Render(m.bodyFrame(), m.heightTier(), gs.focus))
	sections = append(sections, lipgloss.JoinVertical(lipgloss.Left,
		m.renderTapper(),
		m.renderFooter(gs.keys(), gs.focusVerb(m), len(gs.focusables(m))),
	))
	return layout.Stack(sections...)
}

func (gs *gameScreen) renderCapex(m *Model, vk layout.Frame) string {
	var lines []string
	selStart, selEnd := 0, 0
	for i, it := range m.items {
		if !m.unlocked(it) {
			continue
		}
		if i == gs.cursor {
			selStart = len(lines)
		}
		lines = append(lines, gs.capexRow(m, vk, i, it))
		if i == gs.cursor {
			lines = append(lines, theme.DimSty.Width(vk.Width-5).MarginLeft(5).Render(it.desc()))
			selEnd = len(lines) - 1
		}
	}
	if teaser, ok := m.nextLockedTeaser(); ok {
		lines = append(lines, teaser)
	}
	rows := m.panelRows(capexRows)
	gs.capex.Reveal(selEnd, len(lines), rows)
	gs.capex.Reveal(selStart, len(lines), rows)
	return vk.ScrollPanel(content.Text.Capex.Panel, lines, rows, gs.capex.Offset)
}

func (gs *gameScreen) capexRow(m *Model, vk layout.Frame, i int, it capexItem) string {
	s := m.econ.Get()
	cost := it.cost(s)

	selected := i == gs.cursor
	cursor, nameStr := layout.Cursor(false), theme.ValSty.Render(it.name())
	if selected {
		cursor, nameStr = layout.Cursor(true), theme.TitleSty.Render(it.name())
	}

	costSty := theme.CantSty
	if s.Tokens >= cost {
		costSty = theme.CanSty
	}
	left := cursor + it.icon() + " " + nameStr
	right := theme.DimSty.Render(fmt.Sprintf("%6s", it.owned(s))) + "  " + costSty.Render(fmt.Sprintf("%9s 🪙", economy.FormatNum(cost)))

	return vk.Spread(left, right)
}

func (m Model) nextLockedTeaser() (string, bool) {
	s := m.econ.Get()
	for _, it := range m.items {
		if teaser, ok := it.lockedTeaser(s); ok {
			return theme.DimSty.Render(teaser), true
		}
	}
	return "", false
}

func maxAffordableSpecPremiumIdx(tokens float64) int {
	for i := len(specPremiums) - 1; i >= 0; i-- {
		if specPremiums[i] <= tokens {
			return i
		}
	}
	return -1
}
