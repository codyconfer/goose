package game

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/game/viewkit/panels"
	"github.com/codyconfer/goose/internal/game/viewkit/theme"
	"github.com/codyconfer/goose/internal/notify"
)

var (
	specPremiums  = economy.SpecPremiums
	specLeverages = economy.SpecLeverages
)

type specScreen struct {
	prev        screen
	kind        economy.PosKind
	premiumIdx  int
	leverageIdx int
	positions   panels.ScrollState
	ledger      panels.ScrollState
}

func (ss *specScreen) simulates() bool { return true }

func (ss *specScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		_ = m.save()
		return tea.Quit
	case "esc", "d", "q":
		m.screen = ss.prev
	case "left", "h", "right", "l", "tab":
		ss.toggleKind()
	case "up", "k":
		ss.premiumIdx = panels.StepIndex(ss.premiumIdx, 1, len(specPremiums))
	case "down", "j":
		ss.premiumIdx = panels.StepIndex(ss.premiumIdx, -1, len(specPremiums))
	case "]", "+", "=":
		ss.leverageIdx = panels.StepIndex(ss.leverageIdx, 1, len(specLeverages))
	case "[", "-", "_":
		ss.leverageIdx = panels.StepIndex(ss.leverageIdx, -1, len(specLeverages))
	case "enter", " ", "spacebar":
		ss.open(m)
	case "x":
		ss.close(m)
	case "c":
		if n := m.econ.CloseAllPositions(); n > 0 {
			m.setFlash(content.Text.Spec.ClosedAllFlash)
		} else {
			m.setFlash(content.Text.Spec.NothingToClose)
		}
	case "pgup":
		ss.positions.Scroll(-positionRows, len(m.econ.Get().Positions), positionRows)
	case "pgdown":
		ss.positions.Scroll(positionRows, len(m.econ.Get().Positions), positionRows)
	case ",":
		ss.ledger.Scroll(-ledgerRows, len(m.econ.Get().Ledger), ledgerRows)
	case ".":
		ss.ledger.Scroll(ledgerRows, len(m.econ.Get().Ledger), ledgerRows)
	}
	return nil
}

func (ss *specScreen) toggleKind() {
	if ss.posKind() == economy.PosPut {
		ss.kind = economy.PosCall
	} else {
		ss.kind = economy.PosPut
	}
}

func (ss *specScreen) posKind() economy.PosKind {
	if ss.kind == economy.PosPut {
		return economy.PosPut
	}
	return economy.PosCall
}

func (ss *specScreen) premium() float64 {
	if ss.premiumIdx < 0 || ss.premiumIdx >= len(specPremiums) {
		return 0
	}
	return specPremiums[ss.premiumIdx]
}

func (ss *specScreen) leverage() float64 {
	if ss.leverageIdx < 0 || ss.leverageIdx >= len(specLeverages) {
		return 1
	}
	return specLeverages[ss.leverageIdx]
}

func (ss *specScreen) open(m *Model) {
	kind := ss.posKind()
	prem, lev := ss.premium(), ss.leverage()
	if m.econ.OpenPosition(kind, prem, lev, economy.SpecExpirySeconds) {
		desc := fmt.Sprintf("%.0fx %s", lev, specWord(kind))
		m.setFlash(fmt.Sprintf(content.Text.Spec.OpenedFmt, desc, economy.FormatNum(prem)))
	} else {
		m.setFlash(content.Text.Spec.CantAfford)
	}
}

func (ss *specScreen) close(m *Model) {
	if res, ok := m.econ.ClosePosition(ss.positions.Offset); ok {
		m.setFlash(fmt.Sprintf(content.Text.Spec.ClosedFmt, res.Pos.Desc(), economy.FormatNum(res.Payout)))
	} else {
		m.setFlash(content.Text.Spec.NothingToClose)
	}
}

func (ss *specScreen) view(m *Model) string {
	vk := m.frame()
	s := m.econ.Get()
	purse := vk.Panel(content.Text.Spec.PursePanel,
		vk.Spread(theme.EggSty.Render("🪙 "+economy.FormatNum(s.Tokens))+theme.DimSty.Render(" tokens"),
			theme.ValSty.Render(economy.FormatNum(s.EggPrice())+" tokens/egg")+theme.DimSty.Render("  "+content.Text.Spec.PriceLabel)),
		vk.Row(content.Text.Spec.ExposureLabel, theme.ValSty.Render(economy.FormatNum(s.LeveragedExposure())+" 🪙")),
		vk.Row(content.Text.Spec.TrendLabel, tradeTrendLabel(s)),
	)

	sections := []panels.Section{
		{Content: vk.Header(content.Text.Spec.DeskTitle), Priority: panels.Essential},
		{Content: purse, Priority: panels.Essential},
		{Content: renderBook(m), Priority: 40},
		{Content: ss.renderTicket(m), Priority: panels.Essential},
		{Content: ss.renderPositions(m), Priority: panels.Essential},
	}
	if len(s.Positions) > 0 {
		sections = append(sections, panels.Section{Content: ss.renderPnL(m), Priority: 30})
	}
	hints := [][2]string{
		toggleHint("call/put"),
		verticalHint("premium"),
		hint("[ ]/-/+", "leverage"),
		confirmHint("open"),
		hint("x", "close"),
		hint("c", "close all"),
		hint("pgup/pgdn", "positions"),
		hint(",/.", "ledger"),
	}
	hints = append(hints, m.pageHintPairs()...)
	hints = append(hints, hint("esc/d/q", "back"))
	sections = append(sections,
		panels.Section{Content: renderLedger(m, ss.ledger), Priority: 20},
		panels.Section{Content: panels.Flash(vk.Fit(m.flash)), Priority: 10},
		panels.Section{Content: vk.HintLine(hints...), Priority: panels.Essential},
	)
	return panels.StackFit(m.bodyBudget(), sections...)
}

func (ss *specScreen) renderTicket(m *Model) string {
	vk := m.frame()
	kind := ss.posKind()
	thesis := content.Text.Spec.CallThesis
	if kind == economy.PosPut {
		thesis = content.Text.Spec.PutThesis
	}
	dir := panels.Toggle(content.Text.Spec.CallToggle, content.Text.Spec.PutToggle, kind != economy.PosPut)

	prem, lev := ss.premium(), ss.leverage()
	premSty := theme.CanSty
	if prem > m.econ.Get().Tokens {
		premSty = theme.CantSty
	}

	notional := prem * lev
	var warn string
	if lev > 1 {
		warn = theme.CantSty.Render(fmt.Sprintf(content.Text.Spec.WipeWarnFmt, economy.FormatNum(100*((1-economy.SpecMaintenanceMargin)/lev))+"%"))
	} else {
		warn = theme.DimSty.Render("1x — rides to expiry, no early margin call")
	}

	liq := theme.DimSty.Render("n/a")
	buffer := theme.DimSty.Render("n/a")
	if lev > 1 {
		price := m.econ.Get().EggPrice()
		pos := economy.Position{Kind: kind, Strike: price, Premium: prem, Leverage: lev}
		liq = theme.ValSty.Render(economy.FormatNum(pos.LiquidationPrice()) + " tokens/egg")
		buffer = theme.ValSty.Render(economy.FormatNum(pos.MarginCallMove()*100) + "%")
	}

	return vk.Panel(content.Text.Spec.TicketPanel,
		vk.Row(content.Text.Spec.DirectionLabel, dir),
		theme.DimSty.Italic(true).Render("   "+thesis),
		vk.Row(content.Text.Spec.PremiumLabel, premSty.Render(economy.FormatNum(prem)+" 🪙")),
		vk.Row(content.Text.Spec.LeverageLabel, theme.EggSty.Render(fmt.Sprintf("%.0fx", lev))),
		vk.Row(content.Text.Spec.NotionalLabel, theme.ValSty.Render(economy.FormatNum(notional)+" 🪙")),
		vk.Row(content.Text.Spec.LiqPriceLabel, liq),
		vk.Row(content.Text.Spec.BufferLabel, buffer),
		vk.Row(content.Text.Spec.ExpiryLabel, theme.ValSty.Render(fmt.Sprintf(content.Text.Spec.ExpiryFmt, economy.SpecExpirySeconds))),
		vk.Row(content.Text.Spec.RiskLabel, warn),
	)
}

func (ss *specScreen) renderPositions(m *Model) string {
	vk := m.frame()
	s := m.econ.Get()
	if len(s.Positions) == 0 {
		return vk.Panel(content.Text.Spec.PositionsPanel, theme.DimSty.Render(content.Text.Spec.PositionsEmpty))
	}
	price := s.EggPrice()
	var lines []string
	for i, p := range s.Positions {
		marker := panels.Cursor(i == ss.positions.Offset)
		desc := fmt.Sprintf(content.Text.Spec.PosDescFmt, fmt.Sprintf("%.0fx", p.Leverage), specWord(p.Kind), economy.FormatNum(p.Strike))

		pnl := p.PnL(price)
		pnlSty, sign := theme.CanSty, "+"
		if pnl < 0 {
			pnlSty, sign = theme.CantSty, "−"
		}
		mark := pnlSty.Render(fmt.Sprintf("%s%s 🪙", sign, economy.FormatNum(absF(pnl))))

		frac := 0.0
		if economy.SpecExpirySeconds > 0 {
			frac = p.Expiry / economy.SpecExpirySeconds
		}
		bar := panels.Meter(frac, 10)
		clock := theme.DimSty.Render(fmt.Sprintf(content.Text.Spec.ExpiresInFmt, economy.FormatNum(p.Expiry)))
		if p.Leverage > 1 {
			clock = theme.DimSty.Render(fmt.Sprintf("%s %s · %s",
				content.Text.Spec.LiqPriceLabel,
				economy.FormatNum(p.LiquidationPrice()),
				fmt.Sprintf(content.Text.Spec.ExpiresInFmt, economy.FormatNum(p.Expiry)),
			))
		}

		left := marker + theme.ValSty.Render(desc) + "  " + mark
		lines = append(lines, vk.Spread(left, bar+" "+clock))
	}
	return vk.ScrollPanel(content.Text.Spec.PositionsPanel, lines, positionRows, ss.positions.Offset)
}

func renderBook(m *Model) string {
	vk := m.frame()
	s := m.econ.Get()
	data := []panels.Datum{
		{Label: content.Text.Spec.MixCash, Value: s.Tokens},
		{Label: content.Text.Spec.MixEggs, Value: s.Eggs * s.EggPrice()},
		{Label: content.Text.Spec.MixExposure, Value: s.LeveragedExposure()},
	}
	return vk.Pie(content.Text.Spec.MixPanel, data, 48, economy.FormatNum, content.Text.Spec.MixEmpty)
}

func (ss *specScreen) renderPnL(m *Model) string {
	vk := m.frame()
	s := m.econ.Get()
	price := s.EggPrice()
	data := make([]panels.Datum, len(s.Positions))
	for i, p := range s.Positions {
		desc := fmt.Sprintf("%.0fx %s", p.Leverage, specWord(p.Kind))
		data[i] = panels.Datum{Label: desc, Value: p.PnL(price)}
	}
	return vk.BarScroll(content.Text.Spec.PnlPanel, data, meterWidth(vk.Width, 40), economy.FormatNum, content.Text.Spec.PositionsEmpty, pnlRows, ss.positions.Offset)
}

func specWord(k economy.PosKind) string {
	if k == economy.PosPut {
		return content.Text.Spec.PutWord
	}
	return content.Text.Spec.CallWord
}

func absF(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func positionSettleMsg(res economy.PosResult) string {
	if res.PnL >= 0 {
		return fmt.Sprintf(content.Text.Spec.SettledWinFmt, res.Pos.Desc(), economy.FormatNum(res.Payout))
	}
	return fmt.Sprintf(content.Text.Spec.SettledLossFmt, res.Pos.Desc(), economy.FormatNum(-res.PnL))
}

func marginCallNotif(res economy.PosResult) notify.Notification {
	msg := fmt.Sprintf("%s tripped maintenance and got liquidated. %s 🪙 came back; the rest went to the desk.", res.Pos.Desc(), economy.FormatNum(res.Payout))
	if res.Payout <= 0 {
		msg = fmt.Sprintf("%s tripped maintenance and got liquidated. The desk kept the whole premium.", res.Pos.Desc())
	}
	return notify.Notification{
		Title:   content.Text.Spec.MarginCallTitle,
		Message: msg,
		Tone:    notify.ToneNegative,
	}
}
