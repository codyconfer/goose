package game

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/game/viewkit/panels"
	"github.com/codyconfer/goose/internal/game/viewkit/theme"
)

type gameScreen struct {
	cursor int
}

func (gs *gameScreen) simulates() bool { return true }

func (gs *gameScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.quitting = true
		_ = m.save()
		return tea.Quit
	case "enter", " ", "spacebar":
		if m.econ.Get().Frozen() {
			m.setFlash("🚫 Shut down — the subcommittee says the geese may not honk right now.")
		} else {
			m.econ.Tap()
			m.pulse = 1
		}
	case "up", "k":
		gs.cursor = m.prevVisible(gs.cursor)
	case "down", "j":
		gs.cursor = m.nextVisible(gs.cursor)
	case "b", "right", "l":
		gs.buy(m)
	case "s":
		gs.sell(m)
	case "B":
		gs.queueMaxTrade(m, economy.TxBuyEggs)
	case "S":
		gs.queueMaxTrade(m, economy.TxSellEggs)
	case "O", "C":
		gs.openMaxPosition(m, economy.PosCall)
	case "P":
		gs.openMaxPosition(m, economy.PosPut)
	case "t":
		m.screen = &tradeScreen{prev: gs, kind: economy.TxBuyEggs}
	}
	return nil
}

func (gs *gameScreen) buy(m *Model) {
	if gs.cursor < 0 || gs.cursor >= len(m.items) {
		return
	}
	it := m.items[gs.cursor]
	if it.buy(m.econ) {
		m.setFlash(it.boughtMsg(m.econ.Get()))
	} else {
		m.setFlash(it.deniedMsg())
	}
}

func (gs *gameScreen) sell(m *Model) {
	if gs.cursor < 0 || gs.cursor >= len(m.items) {
		return
	}
	it := m.items[gs.cursor]
	if it.sell(m.econ) {
		m.setFlash(it.soldMsg(m.econ.Get()))
	} else {
		m.setFlash(it.cantSellMsg())
	}
}

func (gs *gameScreen) queueMaxTrade(m *Model, kind economy.TxKind) {
	ts := tradeScreen{kind: kind, sizeIdx: len(tradeSizes)}
	ts.schedule(m)
}

func (gs *gameScreen) openMaxPosition(m *Model, kind economy.PosKind) {
	if m.econ.Get().Level() < economy.SpecUnlockLevel {
		m.setFlash(fmt.Sprintf(content.Text.Trade.SpecLockedFmt, economy.SpecUnlockLevel))
		return
	}
	premiumIdx := maxAffordableSpecPremiumIdx(m.econ.Get().Tokens)
	if premiumIdx < 0 {
		m.setFlash(content.Text.Spec.CantAfford)
		return
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
	sections := []string{
		m.renderTitleBar(),
		m.renderStatus(),
	}
	if s.Frozen() {
		sections = append(sections, m.renderShutdown())
	}
	sections = append(sections, gs.renderCapex(m))
	if s.EggsPerSecond() > 0 || s.Eggs > 0 {
		sections = append(sections, m.renderMarket())
	}
	if len(s.Transactions) > 0 || s.Demand() > 0 {
		sections = append(sections, renderTransactions(m))
	}
	sections = append(sections,
		m.renderActivity(),
		m.renderTapper(),
	)
	sections = append(sections, panels.Flash(m.flash))
	sections = append(sections, m.renderFooter())
	return panels.Stack(sections...)
}

func (gs *gameScreen) renderCapex(m *Model) string {
	vk := m.frame()
	var lines []string
	for i, it := range m.items {
		if m.unlocked(it) {
			lines = append(lines, gs.capexRow(m, i, it))
		}
	}
	if teaser, ok := m.nextLockedTeaser(); ok {
		lines = append(lines, teaser)
	}
	return vk.Panel(content.Text.Capex.Panel, lines...)
}

func (gs *gameScreen) capexRow(m *Model, i int, it capexItem) string {
	vk := m.frame()
	s := m.econ.Get()
	cost := it.cost(s)

	selected := i == gs.cursor
	cursor, nameStr := panels.Cursor(false), theme.ValSty.Render(it.name())
	if selected {
		cursor, nameStr = panels.Cursor(true), theme.TitleSty.Render(it.name())
	}

	costSty := theme.CantSty
	if s.Tokens >= cost {
		costSty = theme.CanSty
	}
	left := cursor + it.icon() + " " + nameStr
	right := theme.DimSty.Render(fmt.Sprintf("%6s", it.owned(s))) + "  " + costSty.Render(fmt.Sprintf("%9s 🪙", economy.FormatNum(cost)))

	row := vk.Spread(left, right)
	if selected {
		row += "\n" + theme.DimSty.Width(vk.Width-5).MarginLeft(5).Render(it.desc())
	}
	return row
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
