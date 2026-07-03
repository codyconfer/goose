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

func (gs *gameScreen) focusables(m *Model) layout.Ring {
	return layout.NewRing(
		layout.Focusable{Name: "capex", Interactive: true},
		layout.Focusable{Name: "feed", Interactive: m.feedScrollable()},
	)
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
	case actOpenAgents:
		m.screen = &agentsScreen{prev: gs}
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
	ss := specScreen{
		kind:        kind,
		premiumIdx:  premiumIdx,
		leverageIdx: len(specLeverages) - 1,
	}
	ss.open(m)
}

func (gs *gameScreen) view(m *Model) string {
	s := m.econ.Get()
	sections := []layout.Section{
		{Content: m.renderTitleBar()},
		{Content: m.renderStatus()},
	}
	if s.Frozen() {
		sections = append(sections, layout.Section{Content: m.renderShutdown()})
	}
	focused := gs.focusedPanel(m)
	sections = append(sections, layout.Section{Content: gs.renderCapex(m, focused == "capex")})
	if s.EggsPerSecond() > 0 || s.Eggs > 0 {
		sections = append(sections, layout.Section{Content: m.renderMarket(), MinTier: layout.TierTall})
	}
	if len(s.Transactions) > 0 || s.Demand() > 0 {
		sections = append(sections, layout.Section{Content: renderTransactions(m, layout.ScrollState{}, false), MinTier: layout.TierTall})
	}
	sections = append(sections,
		layout.Section{Content: m.renderFeed(gs.feedScroll.Offset, focused == "feed")},
		layout.Section{Content: m.renderActivity(), MinTier: layout.TierMedium},
		layout.Section{Content: lipgloss.JoinVertical(lipgloss.Left,
			m.renderTapper(),
			m.renderFooter(gs.keys(), gs.focusVerb(m), len(gs.focusables(m))),
		)},
	)
	return layout.StackFit(m.heightTier(), sections...)
}

func (gs *gameScreen) renderCapex(m *Model, focused bool) string {
	vk := m.frame()
	if focused {
		vk = vk.Focus()
	}
	var lines []string
	selStart, selEnd := 0, 0
	for i, it := range m.items {
		if !m.unlocked(it) {
			continue
		}
		if i == gs.cursor {
			selStart = len(lines)
		}
		lines = append(lines, gs.capexRow(m, i, it))
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

func (gs *gameScreen) capexRow(m *Model, i int, it capexItem) string {
	vk := m.frame()
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
