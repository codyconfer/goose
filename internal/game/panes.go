package game

import "github.com/codyconfer/viewkit/layout"

type gamePaneCtx struct {
	m  *Model
	gs *gameScreen
}

type tradePaneCtx struct {
	m  *Model
	ts *tradeScreen
}

var (
	gamePanesReg  = buildGamePanes()
	tradePanesReg = buildTradePanes()
)

func buildScreen[C any](id string, ctx C, reg *layout.Registry[C]) layout.Screen {
	scr, err := layout.BuildScreen(layoutSpec(id), ctx, reg)
	if err != nil {
		scr, _ = layout.BuildScreen(defaultSpec(id), ctx, reg)
	}
	return scr
}

func buildGamePanes() *layout.Registry[gamePaneCtx] {
	r := layout.NewRegistry[gamePaneCtx]()
	r.Pane("capex", "Capital Expenditure", func(c gamePaneCtx) (layout.Pane, bool) {
		return layout.Pane{
			Name:        "capex",
			Title:       "Capital Expenditure",
			Interactive: true,
			Render:      func(f layout.Frame) string { return c.gs.renderCapex(c.m, cellFrame(f)) },
		}, true
	})
	r.Pane("market", "Market", func(c gamePaneCtx) (layout.Pane, bool) {
		s := c.m.econ.Get()
		return layout.Pane{
			Name:    "market",
			Title:   "Market",
			MinTier: layout.TierTall,
			Render:  func(f layout.Frame) string { return c.m.renderMarket(cellFrame(f)) },
		}, s.EggsPerSecond() > 0 || s.Eggs > 0
	})
	r.Pane("orders", "Orders", func(c gamePaneCtx) (layout.Pane, bool) {
		s := c.m.econ.Get()
		return layout.Pane{
			Name:    "orders",
			Title:   "Orders",
			MinTier: layout.TierTall,
			Render:  func(f layout.Frame) string { return renderTransactions(c.m, cellFrame(f), layout.ScrollState{}) },
		}, len(s.Transactions) > 0 || s.Demand() > 0
	})
	r.Pane("feed", "Feed", func(c gamePaneCtx) (layout.Pane, bool) {
		return layout.Pane{
			Name:        "feed",
			Title:       "Feed",
			Interactive: c.m.feedScrollable(),
			Render:      func(f layout.Frame) string { return c.m.renderFeed(cellFrame(f), c.gs.feedScroll.Offset) },
		}, true
	})
	r.Pane("activity", "Activity", func(c gamePaneCtx) (layout.Pane, bool) {
		return layout.Pane{
			Name:    "activity",
			Title:   "Activity",
			MinTier: layout.TierMedium,
			Render:  func(f layout.Frame) string { return c.m.renderActivity(cellFrame(f)) },
		}, true
	})
	return r
}

func buildTradePanes() *layout.Registry[tradePaneCtx] {
	r := layout.NewRegistry[tradePaneCtx]()
	r.Pane("purse", "Purse", func(c tradePaneCtx) (layout.Pane, bool) {
		return layout.Pane{Name: "purse", Render: func(f layout.Frame) string { return c.ts.renderPurse(c.m, cellFrame(f)) }}, true
	})
	r.Pane("chart", "Price Chart", func(c tradePaneCtx) (layout.Pane, bool) {
		return layout.Pane{Name: "chart", MinTier: layout.TierTall, Render: func(f layout.Frame) string { return renderPriceChart(c.m, cellFrame(f)) }}, true
	})
	r.Pane("flow", "Flow", func(c tradePaneCtx) (layout.Pane, bool) {
		return layout.Pane{Name: "flow", MinTier: layout.TierTall, Render: func(f layout.Frame) string { return renderFlow(c.m, cellFrame(f)) }}, true
	})
	r.Pane("builder", "New Order", func(c tradePaneCtx) (layout.Pane, bool) {
		return layout.Pane{Name: "builder", Interactive: true, Render: func(f layout.Frame) string { return c.ts.renderBuilder(c.m, cellFrame(f)) }}, true
	})
	r.Pane("queue", "Queue", func(c tradePaneCtx) (layout.Pane, bool) {
		s := c.m.econ.Get()
		return layout.Pane{Name: "queue", Interactive: len(s.Transactions) > 0, Render: func(f layout.Frame) string { return renderTransactions(c.m, cellFrame(f), c.ts.queue) }}, true
	})
	r.Pane("book", "Order Book", func(c tradePaneCtx) (layout.Pane, bool) {
		return layout.Pane{Name: "book", MinTier: layout.TierTall, Render: func(f layout.Frame) string { return renderBook(c.m, cellFrame(f)) }}, c.ts.specUnlocked(c.m)
	})
	r.Pane("ticket", "New Position", func(c tradePaneCtx) (layout.Pane, bool) {
		return layout.Pane{Name: "ticket", Interactive: true, Render: func(f layout.Frame) string { return c.ts.renderTicket(c.m, cellFrame(f)) }}, c.ts.specUnlocked(c.m)
	})
	r.Pane("positions", "Positions", func(c tradePaneCtx) (layout.Pane, bool) {
		s := c.m.econ.Get()
		return layout.Pane{Name: "positions", Interactive: len(s.Positions) > 0, Render: func(f layout.Frame) string { return c.ts.renderPositions(c.m, cellFrame(f)) }}, c.ts.specUnlocked(c.m)
	})
	r.Pane("pnl", "P&L", func(c tradePaneCtx) (layout.Pane, bool) {
		s := c.m.econ.Get()
		return layout.Pane{Name: "pnl", MinTier: layout.TierTall, Render: func(f layout.Frame) string { return c.ts.renderPnL(c.m, cellFrame(f)) }}, c.ts.specUnlocked(c.m) && len(s.Positions) > 0
	})
	r.Pane("roster", "Roster", func(c tradePaneCtx) (layout.Pane, bool) {
		agents := c.m.econ.Get().Agents
		return layout.Pane{Name: "roster", Interactive: len(agents) > 0, Render: func(f layout.Frame) string { return c.ts.renderRoster(c.m, cellFrame(f), agents) }}, true
	})
	r.Pane("ledger", "Ledger", func(c tradePaneCtx) (layout.Pane, bool) {
		s := c.m.econ.Get()
		return layout.Pane{Name: "ledger", MinTier: layout.TierMedium, Interactive: len(s.Ledger) > c.m.panelRows(ledgerRows), Render: func(f layout.Frame) string { return renderLedger(c.m, cellFrame(f), c.ts.ledger) }}, true
	})
	return r
}
