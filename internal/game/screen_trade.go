package game

import (
	"fmt"
	"slices"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/panels"
	"github.com/codyconfer/viewkit/theme"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
)

var tradeSizes = content.TradeSizes

const (
	priceChartHeight  = 10
	priceChartHistory = 160
	candleSamples     = 3
)

type candle struct{ open, high, low, close float64 }

type tradeScreen struct {
	prev *gameScreen

	kind        economy.TxKind  // spot buy/sell
	pos         economy.PosKind // derivatives call/put
	sizeIdx     int
	premiumIdx  int
	leverageIdx int

	queue        layout.ScrollState
	ledger       layout.ScrollState
	positions    layout.ScrollState
	roster       layout.ScrollState
	rosterCursor int

	focus int
}

func (ts *tradeScreen) simulates() bool { return true }

func (ts *tradeScreen) build(m *Model) layout.Screen {
	return buildScreen(screenTrade, tradePaneCtx{m: m, ts: ts}, tradePanesReg)
}

func (ts *tradeScreen) focusables(m *Model) layout.Ring {
	return ts.build(m).Ring()
}

func (ts *tradeScreen) focusedPanel(m *Model) string {
	return ts.focusables(m).At(ts.focus)
}

func (ts *tradeScreen) specUnlocked(m *Model) bool {
	return m.econ.Get().Level() >= economy.SpecUnlockLevel
}

func (ts *tradeScreen) focusMove(m *Model, delta int) {
	s := m.econ.Get()
	switch ts.focusedPanel(m) {
	case "builder":
		ts.sizeIdx = panels.StepIndex(ts.sizeIdx, -delta, len(tradeSizes)+1)
	case "ticket":
		ts.premiumIdx = panels.StepIndex(ts.premiumIdx, -delta, len(specPremiums))
	case "queue":
		ts.queue.Scroll(delta, len(s.Transactions), m.panelRows(queueRows))
	case "positions":
		ts.positions.Scroll(delta, len(s.Positions), m.panelRows(positionRows))
	case "ledger":
		ts.ledger.Scroll(delta, len(s.Ledger), m.panelRows(ledgerRows))
	case "roster":
		ts.rosterCursor = panels.StepIndex(ts.rosterCursor, delta, len(s.Agents))
	}
}

func (ts *tradeScreen) focusVerb(m *Model) string {
	switch ts.focusedPanel(m) {
	case "ticket":
		return "premium"
	case "queue":
		return "queue"
	case "positions":
		return "positions"
	case "ledger":
		return "ledger"
	case "roster":
		return "select"
	default:
		return "amount"
	}
}

func (ts *tradeScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	action, ok := deskKeymap().Action(msg.String())
	if !ok {
		return nil
	}
	switch action {
	case keys.Quit:
		m.quitting = true
		_ = m.save()
		return tea.Quit
	case keys.Cancel:
		m.screen = ts.prev
	case keys.Up:
		ts.focusMove(m, -1)
	case keys.Down:
		ts.focusMove(m, 1)
	case keys.FocusNext:
		ts.focus = ts.focusables(m).Step(ts.focus, 1)
	case keys.FocusPrev:
		ts.focus = ts.focusables(m).Step(ts.focus, -1)
	case keys.Left:
		ts.adjustLeftRight(m, -1)
	case keys.Right:
		ts.adjustLeftRight(m, 1)
	case keys.Inc:
		ts.adjustIncDec(m, 1)
	case keys.Dec:
		ts.adjustIncDec(m, -1)
	case keys.Confirm:
		ts.confirm(m)
	case actCancelOrder:
		ts.cancelOrClose(m)
	case actClearQueue:
		ts.clearOrCloseAll(m)
	}
	return nil
}

func (ts *tradeScreen) adjustLeftRight(m *Model, delta int) {
	switch ts.focusedPanel(m) {
	case "builder":
		ts.toggleKind()
	case "ticket":
		ts.toggleSpecKind()
	case "roster":
		ts.cycleSize(m, delta)
	}
}

func (ts *tradeScreen) adjustIncDec(m *Model, delta int) {
	switch ts.focusedPanel(m) {
	case "ticket":
		ts.leverageIdx = panels.StepIndex(ts.leverageIdx, delta, len(specLeverages))
	case "roster":
		ts.nudgeThreshold(m, delta)
	}
}

func (ts *tradeScreen) confirm(m *Model) {
	switch ts.focusedPanel(m) {
	case "ticket":
		ts.open(m)
	case "roster":
		m.econ.ToggleAgent(ts.rosterCursor)
	default:
		ts.schedule(m)
	}
}

func (ts *tradeScreen) cancelOrClose(m *Model) {
	if ts.focusedPanel(m) == "positions" {
		ts.close(m)
		return
	}
	if m.econ.CancelTransaction(ts.queue.Offset) {
		m.setFlash(content.Text.Trade.CancelledFlash)
	}
}

func (ts *tradeScreen) clearOrCloseAll(m *Model) {
	if ts.focusedPanel(m) == "positions" {
		if n := m.econ.CloseAllPositions(); n > 0 {
			m.setFlash(content.Text.Spec.ClosedAllFlash)
		} else {
			m.setFlash(content.Text.Spec.NothingToClose)
		}
		return
	}
	if len(m.econ.Get().Transactions) > 0 {
		m.econ.ClearTransactions()
		m.setFlash(content.Text.Trade.ClearedFlash)
	}
}

func (ts *tradeScreen) toggleKind() {
	if ts.kind == economy.TxBuyEggs {
		ts.kind = economy.TxSellEggs
	} else {
		ts.kind = economy.TxBuyEggs
	}
}

func (ts *tradeScreen) amount(m *Model) float64 {
	if ts.sizeIdx >= 0 && ts.sizeIdx < len(tradeSizes) {
		return tradeSizes[ts.sizeIdx]
	}
	s := m.econ.Get()
	if ts.kind == economy.TxBuyEggs {
		if p := s.BuyPrice(); p > 0 {
			return s.Tokens / p
		}
		return 0
	}
	return s.Eggs
}

func (ts *tradeScreen) schedule(m *Model) {
	amt := ts.amount(m)
	if m.econ.ScheduleTrade(ts.kind, amt) {
		m.setFlash(fmt.Sprintf(content.Text.Trade.QueuedFmt, tradeVerb(ts.kind), economy.FormatNum(amt)))
	} else {
		m.setFlash(content.Text.Trade.NothingToSchedule)
	}
}

func (ts *tradeScreen) view(m *Model) string {
	vk := m.frame()
	km := deskKeymap()
	pane := ts.focusedPanel(m)

	hints := [][2]string{km.HintLabeled(keys.Up, ts.focusVerb(m))}
	switch pane {
	case "builder":
		hints = append(hints, km.HintLabeled(keys.Left, "buy/sell"))
	case "ticket":
		hints = append(hints, km.HintLabeled(keys.Left, "call/put"), km.HintLabeled(keys.Inc, "leverage"))
	case "roster":
		hints = append(hints, km.HintLabeled(keys.Left, "size"), km.HintLabeled(keys.Inc, "threshold"))
	}
	if len(ts.focusables(m)) > 1 {
		hints = append(hints, km.Hint(keys.FocusNext))
	}
	hints = append(hints, km.HintLabeled(keys.Confirm, confirmLabel(pane)))
	if pane == "positions" {
		hints = append(hints, km.HintLabeled(actCancelOrder, "close"), km.HintLabeled(actClearQueue, "close all"))
	} else {
		hints = append(hints, km.Hint(actCancelOrder), km.Hint(actClearQueue))
	}
	hints = append(hints, km.Hint(keys.Cancel))

	body := ts.build(m).Render(m.bodyFrame(), m.heightTier(), ts.focus)
	return layout.Stack(
		vk.Header(content.Text.Trade.DeskTitle),
		body,
		panels.Flash(vk.Fit(m.flash)),
		vk.HintLine(hints...),
	)
}

func confirmLabel(pane string) string {
	switch pane {
	case "ticket":
		return "open"
	case "roster":
		return "hire/bench"
	default:
		return "queue"
	}
}

func (ts *tradeScreen) renderPurse(m *Model, vk layout.Frame) string {
	s := m.econ.Get()
	return vk.Panel(content.Text.Trade.PursePanel,
		vk.Spread(theme.AccentSty.Render("🪙 "+economy.FormatNum(s.Tokens))+theme.DimSty.Render(" tokens"),
			theme.AccentSty.Render("🥚 "+economy.FormatNum(s.Eggs))+theme.DimSty.Render(" eggs")),
		vk.Row(content.Text.Trade.MarketPriceLabel, theme.ValSty.Render(economy.FormatNum(s.EggPrice())+" tokens/egg")),
		vk.Row(content.Text.Trade.ConsumersPayLabel, theme.CanSty.Render(economy.FormatNum(s.SellPrice())+" tokens/egg")),
		vk.Row(content.Text.Trade.TrendLabel, tradeTrendLabel(s)),
		vk.Row(content.Text.Trade.TrendStrengthLabel, tradeTrendStrength(s)),
	)
}

func (ts *tradeScreen) renderBuilder(m *Model, vk layout.Frame) string {
	dir := panels.Toggle(content.Text.Trade.BuyToggle, content.Text.Trade.SellToggle, ts.kind == economy.TxBuyEggs)

	amt := ts.amount(m)
	s := m.econ.Get()
	var note string
	if ts.kind == economy.TxBuyEggs {
		worth := amt * s.BuyPrice()
		sty := theme.CanSty
		if worth > s.Tokens {
			sty = theme.CantSty
		}
		note = sty.Render(fmt.Sprintf(content.Text.Trade.SpendFmt, economy.FormatNum(worth)))
	} else {
		worth := amt * s.SellPrice()
		sty := theme.CanSty
		if amt > s.Eggs {
			sty = theme.CantSty
		}
		note = sty.Render(fmt.Sprintf(content.Text.Trade.ProceedsFmt, economy.FormatNum(worth)))
	}

	return vk.Panel(content.Text.Trade.NewOrderPanel,
		vk.Row(content.Text.Trade.DirectionLabel, dir),
		vk.Row(content.Text.Trade.AmountLabel, theme.AccentSty.Render(ts.sizeLabel(m)+" 🥚")),
		vk.Row(content.Text.Trade.EstimateLabel, note),
	)
}

func (ts *tradeScreen) sizeLabel(m *Model) string {
	if ts.sizeIdx >= 0 && ts.sizeIdx < len(tradeSizes) {
		return economy.FormatNum(tradeSizes[ts.sizeIdx])
	}
	return fmt.Sprintf(content.Text.Trade.MaxFmt, economy.FormatNum(ts.amount(m)))
}

// --- Derivatives (formerly specScreen) ---

func (ts *tradeScreen) toggleSpecKind() {
	if ts.posKind() == economy.PosPut {
		ts.pos = economy.PosCall
	} else {
		ts.pos = economy.PosPut
	}
}

func (ts *tradeScreen) posKind() economy.PosKind {
	if ts.pos == economy.PosPut {
		return economy.PosPut
	}
	return economy.PosCall
}

func (ts *tradeScreen) premium() float64 {
	if ts.premiumIdx < 0 || ts.premiumIdx >= len(specPremiums) {
		return 0
	}
	return specPremiums[ts.premiumIdx]
}

func (ts *tradeScreen) leverage() float64 {
	if ts.leverageIdx < 0 || ts.leverageIdx >= len(specLeverages) {
		return 1
	}
	return specLeverages[ts.leverageIdx]
}

func (ts *tradeScreen) open(m *Model) {
	kind := ts.posKind()
	prem, lev := ts.premium(), ts.leverage()
	if m.econ.OpenPosition(kind, prem, lev, economy.SpecExpirySeconds) {
		desc := fmt.Sprintf("%.0fx %s", lev, specWord(kind))
		m.setFlash(fmt.Sprintf(content.Text.Spec.OpenedFmt, desc, economy.FormatNum(prem)))
	} else {
		m.setFlash(content.Text.Spec.CantAfford)
	}
}

func (ts *tradeScreen) close(m *Model) {
	if res, ok := m.econ.ClosePosition(ts.positions.Offset); ok {
		m.setFlash(fmt.Sprintf(content.Text.Spec.ClosedFmt, res.Pos.Desc(), economy.FormatNum(res.Payout)))
	} else {
		m.setFlash(content.Text.Spec.NothingToClose)
	}
}

func (ts *tradeScreen) renderTicket(m *Model, vk layout.Frame) string {
	kind := ts.posKind()
	thesis := content.Text.Spec.CallThesis
	if kind == economy.PosPut {
		thesis = content.Text.Spec.PutThesis
	}
	dir := panels.Toggle(content.Text.Spec.CallToggle, content.Text.Spec.PutToggle, kind != economy.PosPut)

	prem, lev := ts.premium(), ts.leverage()
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
		vk.Row(content.Text.Spec.LeverageLabel, theme.AccentSty.Render(fmt.Sprintf("%.0fx", lev))),
		vk.Row(content.Text.Spec.NotionalLabel, theme.ValSty.Render(economy.FormatNum(notional)+" 🪙")),
		vk.Row(content.Text.Spec.LiqPriceLabel, liq),
		vk.Row(content.Text.Spec.BufferLabel, buffer),
		vk.Row(content.Text.Spec.ExpiryLabel, theme.ValSty.Render(fmt.Sprintf(content.Text.Spec.ExpiryFmt, economy.SpecExpirySeconds))),
		vk.Row(content.Text.Spec.RiskLabel, warn),
	)
}

func (ts *tradeScreen) renderPositions(m *Model, vk layout.Frame) string {
	s := m.econ.Get()
	if len(s.Positions) == 0 {
		return vk.Panel(content.Text.Spec.PositionsPanel, theme.DimSty.Render(content.Text.Spec.PositionsEmpty))
	}
	price := s.EggPrice()
	var lines []string
	for i, p := range s.Positions {
		marker := layout.Cursor(i == ts.positions.Offset)
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
	return vk.ScrollPanel(content.Text.Spec.PositionsPanel, lines, m.panelRows(positionRows), ts.positions.Offset)
}

func (ts *tradeScreen) renderPnL(m *Model, vk layout.Frame) string {
	s := m.econ.Get()
	price := s.EggPrice()
	data := make([]panels.Datum, len(s.Positions))
	for i, p := range s.Positions {
		desc := fmt.Sprintf("%.0fx %s", p.Leverage, specWord(p.Kind))
		data[i] = panels.Datum{Label: desc, Value: p.PnL(price)}
	}
	return panels.BarScroll(vk, content.Text.Spec.PnlPanel, data, panels.MeterWidth(vk.Width, 40), economy.FormatNum, content.Text.Spec.PositionsEmpty, m.panelRows(pnlRows), ts.positions.Offset)
}

// --- Agents (formerly agentsScreen) ---

func (ts *tradeScreen) currentAgent(m *Model) (economy.Agent, bool) {
	agents := m.econ.Get().Agents
	if ts.rosterCursor < 0 || ts.rosterCursor >= len(agents) {
		return economy.Agent{}, false
	}
	return agents[ts.rosterCursor], true
}

func (ts *tradeScreen) cycleSize(m *Model, delta int) {
	a, ok := ts.currentAgent(m)
	if !ok {
		return
	}
	opts := sizeOptions(a)
	if len(opts) == 0 {
		return
	}
	idx := nearestIndex(opts, a.Size)
	idx = panels.StepIndex(idx, delta, len(opts))
	m.econ.SetAgentSize(ts.rosterCursor, opts[idx])
}

func (ts *tradeScreen) nudgeThreshold(m *Model, delta int) {
	a, ok := ts.currentAgent(m)
	if !ok {
		return
	}
	next := a.Threshold + float64(delta)*thresholdStep(a.Metric)
	if a.Metric == economy.MetricTrend {
		next = clampTrendThreshold(next)
	} else if next < 0 {
		next = 0
	}
	m.econ.SetAgentThreshold(ts.rosterCursor, next)
}

func (ts *tradeScreen) renderRoster(m *Model, vk layout.Frame, agents []economy.Agent) string {
	if len(agents) == 0 {
		return vk.Panel(content.Text.Agents.Panel, theme.DimSty.Render(content.Text.Agents.Empty))
	}
	var lines []string
	selStart, selEnd := -1, -1
	for i, a := range agents {
		selected := i == ts.rosterCursor
		rc := content.Text.Agents.Roster[a.Key]

		nameSty := theme.ValSty
		if selected {
			nameSty = theme.TitleSty
		}
		name := layout.Cursor(selected) + nameSty.Render(rc.Name)

		status := theme.DimSty.Render(content.Text.Agents.OffWord)
		if a.Enabled {
			status = theme.CanSty.Render(content.Text.Agents.OnWord)
		}

		if selected {
			selStart = len(lines)
		}
		lines = append(lines, vk.Spread(name, status))
		lines = append(lines, theme.DimSty.Width(vk.Width-5).MarginLeft(5).Render(agentRule(a)))
		if selected {
			lines = append(lines, theme.DimSty.Italic(true).Width(vk.Width-5).MarginLeft(5).Render(rc.Blurb))
			selEnd = len(lines) - 1
		}
	}
	rows := m.panelRows(agentRows)
	if selStart >= 0 {
		ts.roster.Reveal(selEnd, len(lines), rows)
		ts.roster.Reveal(selStart, len(lines), rows)
	}
	return vk.ScrollPanel(content.Text.Agents.Panel, lines, rows, ts.roster.Offset)
}

// --- shared render helpers ---

func renderPriceChart(m *Model, vk layout.Frame) string {
	cs := m.candles
	if len(cs) == 0 {
		return vk.Panel(content.Text.Trade.PriceChartPanel, theme.DimSty.Render(content.Text.Trade.PriceChartGathering))
	}
	width := max(vk.Width-7, 1)
	if len(cs) > width {
		cs = cs[len(cs)-width:]
	}

	last := cs[len(cs)-1]
	cur, first := last.close, cs[0].open
	trend := theme.DimSty.Render(content.Text.Trade.TrendFlat)
	switch {
	case cur > first*1.0001:
		trend = theme.CanSty.Render(fmt.Sprintf(content.Text.Trade.TrendUpFmt, (cur/first-1)*100))
	case cur < first*0.9999:
		trend = theme.CantSty.Render(fmt.Sprintf(content.Text.Trade.TrendDownFmt, (1-cur/first)*100))
	}
	footer := vk.Spread(
		theme.AccentSty.Render(content.Text.Trade.NowPrefix+economy.FormatNum(cur))+theme.DimSty.Render(" tokens/egg"),
		trend)

	beats := (len(cs)-1)*candleSamples + m.candleBeats
	return panels.Candle(vk, fmt.Sprintf(content.Text.Trade.PriceChartTitleFmt, beats), toOHLC(cs), width, priceChartHeight, economy.FormatNum, footer)
}

func renderFlow(m *Model, vk layout.Frame) string {
	s := m.econ.Get()
	data := []panels.Datum{
		{Label: content.Text.Trade.FlowLaying, Value: s.EggsPerSecond()},
		{Label: content.Text.Trade.FlowSelling, Value: m.sellRate},
		{Label: content.Text.Trade.FlowDemand, Value: s.Demand()},
	}
	return panels.Bar(vk, content.Text.Trade.FlowPanel, data, panels.MeterWidth(vk.Width, 40), economy.FormatNum, "")
}

func toOHLC(cs []candle) []panels.OHLC {
	out := make([]panels.OHLC, len(cs))
	for i, c := range cs {
		out[i] = panels.OHLC{Open: c.open, High: c.high, Low: c.low, Close: c.close}
	}
	return out
}

func renderTransactions(m *Model, vk layout.Frame, sv layout.ScrollState) string {
	s := m.econ.Get()
	var prefix []string

	if s.Demand() > 0 {
		left := theme.AccentSty.Render(content.Text.Trade.QueueConsumersLabel) + theme.DimSty.Render(content.Text.Trade.QueueConsumersSuffix)
		right := theme.CanSty.Render("+" + economy.FormatNum(m.sellRate) + " /sec")
		prefix = append(prefix, vk.Spread(left, right))
	}

	lines := make([]string, 0, len(s.Transactions))
	for i, o := range s.Transactions {
		marker := layout.Cursor(i == sv.Offset)
		label := fmt.Sprintf("%s %s 🥚", tradeVerb(o.Kind), economy.FormatNum(o.Amount))
		frac := 0.0
		if o.Amount > 0 {
			frac = o.Filled / o.Amount
		}
		bar := panels.Meter(frac, 14)
		pct := theme.DimSty.Render(fmt.Sprintf("%3.0f%%", frac*100))
		lines = append(lines, vk.Spread(marker+theme.ValSty.Render(label), bar+" "+pct))
	}

	if len(lines) == 0 {
		return vk.Panel(content.Text.Trade.QueuePanel, append(prefix, theme.DimSty.Render(content.Text.Trade.QueueQuiet))...)
	}
	return vk.ScrollPanelWithPrefix(content.Text.Trade.QueuePanel, prefix, lines, m.panelRows(queueRows), sv.Offset)
}

func tradeVerb(k economy.TxKind) string {
	if k == economy.TxSellEggs {
		return content.Text.Trade.VerbSell
	}
	return content.Text.Trade.VerbBuy
}

func tradeTrendLabel(s economy.State) string {
	strength := s.TrendStrength()
	switch {
	case strength < 0.08:
		return theme.DimSty.Render(content.Text.Trade.TrendSideways)
	case s.PriceTrend > 0:
		return theme.CanSty.Render(fmt.Sprintf(content.Text.Trade.TrendBullFmt, strength*100))
	default:
		return theme.CantSty.Render(fmt.Sprintf(content.Text.Trade.TrendBearFmt, strength*100))
	}
}

func tradeTrendStrength(s economy.State) string {
	return panels.Meter(s.TrendStrength(), 12) + " " + theme.DimSty.Render(fmt.Sprintf("%.0f%%", s.TrendStrength()*100))
}

func tradeCompletedMsg(o economy.Transaction) string {
	if o.Kind == economy.TxSellEggs {
		return fmt.Sprintf(content.Text.Trade.CompletedSellFmt, economy.FormatNum(o.Amount))
	}
	return fmt.Sprintf(content.Text.Trade.CompletedBuyFmt, economy.FormatNum(o.Amount))
}

const feedHistory = 50

var (
	capexRows    = layout.TierRows{Short: 4, Medium: 6, Tall: 14}
	ledgerRows   = layout.TierRows{Short: 3, Medium: 6, Tall: 12}
	pnlRows      = layout.TierRows{Short: 3, Medium: 6, Tall: 12}
	positionRows = layout.TierRows{Short: 3, Medium: 6, Tall: 12}
	queueRows    = layout.TierRows{Short: 3, Medium: 6, Tall: 12}
	agentRows    = layout.TierRows{Short: 4, Medium: 6, Tall: 14}
	feedRows     = layout.TierRows{Short: 3, Medium: 5, Tall: 8}
)

func renderLedger(m *Model, vk layout.Frame, sv layout.ScrollState) string {
	led := m.econ.Get().Ledger
	rows := make([]panels.LedgerRow, 0, len(led))
	for _, tx := range slices.Backward(led) {
		rows = append(rows, panels.LedgerRow{Label: ledgerDesc(tx), Delta: tx.Tokens})
	}
	return panels.Ledger(vk, content.Text.Trade.LedgerPanel, rows, "🪙", economy.FormatNum, m.panelRows(ledgerRows), sv.Offset, content.Text.Trade.LedgerEmpty)
}

func ledgerDesc(tx economy.Transaction) string {
	switch tx.Kind {
	case economy.TxBuyEggs:
		return fmt.Sprintf(content.Text.Trade.LedgerBuyEggsFmt, economy.FormatNum(tx.Amount))
	case economy.TxSellEggs:
		return fmt.Sprintf(content.Text.Trade.LedgerSellEggsFmt, economy.FormatNum(tx.Amount))
	case economy.TxBuyProducer:
		if icon, ok := producerIcon(tx.Label); ok {
			return icon + " Bought " + tx.Label
		}
		return content.Text.Trade.LedgerBuyProducerPrefix + tx.Label
	case economy.TxSellProducer:
		if icon, ok := producerIcon(tx.Label); ok {
			return icon + " Decommissioned " + tx.Label
		}
		return content.Text.Trade.LedgerSellProducerPrefix + tx.Label
	case economy.TxUpgrade:
		if icon, ok := upgradeIcon(tx.Label); ok {
			return icon + " Upgraded " + tx.Label
		}
		return content.Text.Trade.LedgerUpgradePrefix + tx.Label
	case economy.TxOptionOpen:
		return fmt.Sprintf(content.Text.Trade.LedgerOptionOpenFmt, tx.Label)
	case economy.TxOptionSettle:
		return fmt.Sprintf(content.Text.Trade.LedgerOptionSettleFmt, tx.Label)
	default:
		return tx.Label
	}
}

func producerIcon(label string) (string, bool) {
	for _, p := range economy.Producers {
		if p.Name == label {
			return p.Icon, true
		}
	}
	return "", false
}

func upgradeIcon(label string) (string, bool) {
	for _, u := range economy.Upgrades {
		if u.Name == label {
			return u.Icon, true
		}
	}
	return "", false
}
