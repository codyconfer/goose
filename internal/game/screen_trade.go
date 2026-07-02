package game

import (
	"fmt"
	"slices"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/game/viewkit/panels"
	"github.com/codyconfer/goose/internal/game/viewkit/theme"
)

var tradeSizes = content.TradeSizes

const (
	priceChartHeight  = 10
	priceChartHistory = 160
	candleSamples     = 3
)

type candle struct{ open, high, low, close float64 }

type tradeScreen struct {
	prev    *gameScreen
	kind    economy.TxKind
	sizeIdx int
	queue   panels.ScrollState
	ledger  panels.ScrollState
}

func (ts *tradeScreen) simulates() bool { return true }

func (ts *tradeScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		_ = m.save()
		return tea.Quit
	case "esc", "t", "q":
		m.screen = ts.prev
	case "d":
		ts.openSpec(m)
	case "left", "h", "right", "l", "tab":
		ts.toggleKind()
	case "up", "k":
		ts.sizeIdx = panels.StepIndex(ts.sizeIdx, 1, len(tradeSizes)+1)
	case "down", "j":
		ts.sizeIdx = panels.StepIndex(ts.sizeIdx, -1, len(tradeSizes)+1)
	case "enter", " ", "spacebar":
		ts.schedule(m)
	case "c":
		if len(m.econ.Get().Transactions) > 0 {
			m.econ.ClearTransactions()
			m.setFlash(content.Text.Trade.ClearedFlash)
		}
	case "x":
		if m.econ.CancelTransaction(ts.queue.Offset) {
			m.setFlash(content.Text.Trade.CancelledFlash)
		}
	case "pgup":
		ts.queue.Scroll(-queueRows, len(m.econ.Get().Transactions), queueRows)
	case "pgdown":
		ts.queue.Scroll(queueRows, len(m.econ.Get().Transactions), queueRows)
	case ",":
		ts.ledger.Scroll(-ledgerRows, len(m.econ.Get().Ledger), ledgerRows)
	case ".":
		ts.ledger.Scroll(ledgerRows, len(m.econ.Get().Ledger), ledgerRows)
	}
	return nil
}

func (ts *tradeScreen) openSpec(m *Model) {
	if m.econ.Get().Level() < economy.SpecUnlockLevel {
		m.setFlash(fmt.Sprintf(content.Text.Trade.SpecLockedFmt, economy.SpecUnlockLevel))
		return
	}
	m.screen = &specScreen{prev: ts, kind: economy.PosCall}
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
	s := m.econ.Get()
	purse := vk.Panel(content.Text.Trade.PursePanel,
		vk.Spread(theme.EggSty.Render("🪙 "+economy.FormatNum(s.Tokens))+theme.DimSty.Render(" tokens"),
			theme.EggSty.Render("🥚 "+economy.FormatNum(s.Eggs))+theme.DimSty.Render(" eggs")),
		vk.Row(content.Text.Trade.MarketPriceLabel, theme.ValSty.Render(economy.FormatNum(s.EggPrice())+" tokens/egg")),
		vk.Row(content.Text.Trade.ConsumersPayLabel, theme.CanSty.Render(economy.FormatNum(s.SellPrice())+" tokens/egg")),
		vk.Row(content.Text.Trade.TrendLabel, tradeTrendLabel(s)),
		vk.Row(content.Text.Trade.TrendStrengthLabel, tradeTrendStrength(s)),
	)

	hints := []([2]string){
		{"←/→", "buy/sell"},
		{"↑/↓", "amount"},
		{"enter", "queue"},
		{"x", "cancel"},
		{"c", "clear"},
		{"pgup/pgdn", "queue"},
		{",/.", "ledger"},
		{"ctrl+u/d", "page"},
	}
	if s.Level() >= economy.SpecUnlockLevel {
		hints = append(hints, [2]string{"d", "derivatives"})
	}
	hints = append(hints, [2]string{"esc", "back"})
	return panels.StackFit(m.bodyBudget(),
		panels.Section{Content: vk.Header(content.Text.Trade.DeskTitle), Priority: panels.Essential},
		panels.Section{Content: purse, Priority: panels.Essential},
		panels.Section{Content: renderPriceChart(m), Priority: 40},
		panels.Section{Content: renderFlow(m), Priority: 30},
		panels.Section{Content: ts.renderBuilder(m), Priority: panels.Essential},
		panels.Section{Content: renderTransactions(m, ts.queue), Priority: panels.Essential},
		panels.Section{Content: renderLedger(m, ts.ledger), Priority: 20},
		panels.Section{Content: panels.Flash(vk.Fit(m.flash)), Priority: 10},
		panels.Section{Content: vk.HintLine(hints...), Priority: panels.Essential},
	)
}

func (ts *tradeScreen) renderBuilder(m *Model) string {
	vk := m.frame()
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
		vk.Row(content.Text.Trade.AmountLabel, theme.EggSty.Render(ts.sizeLabel(m)+" 🥚")),
		vk.Row(content.Text.Trade.EstimateLabel, note),
	)
}

func (ts *tradeScreen) sizeLabel(m *Model) string {
	if ts.sizeIdx >= 0 && ts.sizeIdx < len(tradeSizes) {
		return economy.FormatNum(tradeSizes[ts.sizeIdx])
	}
	return fmt.Sprintf(content.Text.Trade.MaxFmt, economy.FormatNum(ts.amount(m)))
}

func renderPriceChart(m *Model) string {
	vk := m.frame()
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
		theme.EggSty.Render(content.Text.Trade.NowPrefix+economy.FormatNum(cur))+theme.DimSty.Render(" tokens/egg"),
		trend)

	beats := (len(cs)-1)*candleSamples + m.candleBeats
	return vk.Candle(fmt.Sprintf(content.Text.Trade.PriceChartTitleFmt, beats), toOHLC(cs), width, priceChartHeight, economy.FormatNum, footer)
}

func renderFlow(m *Model) string {
	vk := m.frame()
	s := m.econ.Get()
	data := []panels.Datum{
		{Label: content.Text.Trade.FlowLaying, Value: s.EggsPerSecond()},
		{Label: content.Text.Trade.FlowSelling, Value: m.sellRate},
		{Label: content.Text.Trade.FlowDemand, Value: s.Demand()},
	}
	return vk.Bar(content.Text.Trade.FlowPanel, data, meterWidth(vk.Width, 40), economy.FormatNum, "")
}

func toOHLC(cs []candle) []panels.OHLC {
	out := make([]panels.OHLC, len(cs))
	for i, c := range cs {
		out[i] = panels.OHLC{Open: c.open, High: c.high, Low: c.low, Close: c.close}
	}
	return out
}

func renderTransactions(m *Model, sv panels.ScrollState) string {
	vk := m.frame()
	s := m.econ.Get()
	var prefix []string

	if s.Demand() > 0 {
		left := theme.EggSty.Render(content.Text.Trade.QueueConsumersLabel) + theme.DimSty.Render(content.Text.Trade.QueueConsumersSuffix)
		right := theme.CanSty.Render("+" + economy.FormatNum(m.sellRate) + " /sec")
		prefix = append(prefix, vk.Spread(left, right))
	}

	lines := make([]string, 0, len(s.Transactions))
	for i, o := range s.Transactions {
		marker := panels.Cursor(i == sv.Offset)
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
	return vk.ScrollPanelWithPrefix(content.Text.Trade.QueuePanel, prefix, lines, queueRows, sv.Offset)
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

const (
	capexRows    = 8
	ledgerRows   = 8
	pnlRows      = 8
	positionRows = 8
	queueRows    = 8
	agentRows    = 8
)

func renderLedger(m *Model, sv panels.ScrollState) string {
	vk := m.frame()
	led := m.econ.Get().Ledger
	rows := make([]panels.LedgerRow, 0, len(led))
	for _, tx := range slices.Backward(led) {
		rows = append(rows, panels.LedgerRow{Label: ledgerDesc(tx), Delta: tx.Tokens})
	}
	return vk.Ledger(content.Text.Trade.LedgerPanel, rows, "🪙", economy.FormatNum, ledgerRows, sv.Offset, content.Text.Trade.LedgerEmpty)
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
