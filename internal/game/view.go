package game

import (
	"fmt"
	"math"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/panels"
	"github.com/codyconfer/viewkit/theme"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
)

func (m Model) View() string {
	if m.quitting {
		return theme.AppFrame.Render(theme.TitleSty.Render(content.Text.App.Quit))
	}
	if !layout.FitsScreenWidth(m.width) {
		return theme.AppFrame.Render(layout.TooNarrow(m.width))
	}
	return theme.AppFrame.Render(layout.ViewportLayout(m.screen.view(&m), layout.ContentRows(m.height), m.pageScroll))
}

func (m Model) frame() layout.Frame {
	return layout.ScreenFrame(m.width)
}

func (m Model) bodyFrame() layout.Frame {
	return layout.NewFrame(m.frame().Width + 4)
}

func (m Model) renderTitleBar() string {
	return m.frame().Header(content.Text.App.Title, content.Text.App.Subtitle)
}

func (m Model) renderStatus() string {
	vk := m.frame()
	s := m.econ.Get()
	lvl := s.Level()

	tokens := vk.Spread(
		theme.AccentSty.Render("🪙 "+economy.FormatNum(s.Tokens))+theme.DimSty.Render(" tokens"),
		theme.DimSty.Render(fmt.Sprintf(content.Text.Status.RateFmt, economy.FormatNum(s.TokensPerSecond()), economy.FormatNum(s.PerClick))),
	)
	eggs := vk.Spread(
		theme.AccentSty.Render("🥚 "+economy.FormatNum(s.Eggs))+theme.DimSty.Render(" eggs"),
		theme.TitleSty.Render(fmt.Sprintf(content.Text.Status.LevelFmt, lvl)),
	)

	var progress string
	if next, ok := s.NextLevelEggs(); ok {
		le := s.LevelEggs()
		prev := s.LevelFloor()
		frac := (le - prev) / (next - prev)
		progress = panels.Meter(frac, panels.MeterWidth(vk.Width, 22)) +
			theme.DimSty.Render("  "+fmt.Sprintf(content.Text.Status.ProgressFmt, economy.FormatNum(le), economy.FormatNum(next), lvl+1))
	} else {
		progress = theme.AccentSty.Render("★ ") + theme.DimSty.Render(content.Text.Status.MaxLevel)
	}

	lines := []string{tokens, eggs, progress}
	return vk.Panel(content.Text.Status.Panel, lines...)
}

func (m Model) renderTapper() string {
	vk := m.frame()
	goose := "🪿"
	caption := theme.DimSty.Render("press ") + theme.KeySty.Render("[enter]") + theme.DimSty.Render(" to generate a token")
	if m.pulse > 0 {
		goose = "🪿💥"
		caption = theme.AccentSty.Render(fmt.Sprintf("+%s 🪙", economy.FormatNum(m.econ.Get().PerClick)))
	}
	card := theme.CardSty.Width(vk.Width + 2).Render(lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Render(goose),
		caption,
	))
	if m.offline > 0 {
		note := theme.DimSty.Render(fmt.Sprintf(content.Text.Tapper.OfflineFmt, economy.FormatNum(m.offline)))
		return lipgloss.JoinVertical(lipgloss.Left, card, note)
	}
	return card
}

func (m Model) renderShutdown() string {
	vk := m.frame()
	s := m.econ.Get()
	secs := int(math.Ceil(s.FreezeSeconds))
	reason := s.FreezeReason
	if reason == "" {
		reason = "reasons that will be explained later, or never"
	}
	body := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render("🚫 BUSINESS SHUT DOWN BY ORDER OF A SUBCOMMITTEE"),
		lipgloss.NewStyle().Width(vk.Width).Render(
			fmt.Sprintf("Stated concern: %s.  Cleared to reopen in ~%ds. No eggs, no sales, no honking until then.", reason, secs),
		),
	)
	return theme.NotifNegativeSty.Render(body)
}

func (m Model) renderActivity() string {
	if m.notifs.Active() {
		return m.renderNotification()
	}
	return theme.NotifIdleSty.Render(lipgloss.NewStyle().Width(m.frame().Width).Render(content.Text.Activity.Idle))
}

func (m Model) renderFeed(offset int, focused bool) string {
	if !m.feed.active() {
		return ""
	}
	vk := m.frame()
	if focused {
		vk = vk.Focus()
	}
	raw := m.feed.lines()

	lines := make([]string, len(raw))
	for i, ln := range raw {
		lines[len(raw)-1-i] = theme.DimSty.Italic(true).Render(vk.Fit(ln))
	}
	return vk.ScrollPanel(content.Text.Feed.Panel, lines, m.panelRows(feedRows), offset)
}

func (m Model) feedScrollable() bool {
	return m.feed.size() > m.panelRows(feedRows)
}

func (m Model) renderMarket() string {
	vk := m.frame()
	s := m.econ.Get()
	priceTag := theme.DimSty.Render(content.Text.Market.PriceSteady)
	switch {
	case s.PriceFactor > 1.08:
		priceTag = theme.CanSty.Render(content.Text.Market.PriceDemand)
	case s.PriceFactor < 0.92:
		priceTag = theme.CantSty.Render(content.Text.Market.PriceGlut)
	}
	return vk.Panel(content.Text.Market.Panel,
		vk.Row(content.Text.Market.StockLabel, theme.AccentSty.Render(economy.FormatNum(s.Eggs)+" eggs")),
		vk.Row(content.Text.Market.MarketCapLabel, theme.ValSty.Render(economy.FormatNum(s.MarketCap())+" eggs")),
		vk.Row(content.Text.Market.LayingLabel, theme.CanSty.Render("+"+economy.FormatNum(s.EggsPerSecond())+" /sec")),
		vk.Row(content.Text.Market.SellingLabel, theme.CanSty.Render("+"+economy.FormatNum(m.sellRate)+" /sec")),
		vk.Row(content.Text.Market.ConsumersLabel, theme.ValSty.Render(economy.FormatNum(s.Consumers))),
		vk.Row(content.Text.Market.PriceLabel, theme.ValSty.Render(economy.FormatNum(s.SellPrice())+" tokens/egg ")+priceTag),
	)
}

func (m Model) renderFooter(km *keys.Map, focusVerb string, ringSize int) string {
	hints := [][2]string{
		km.Hint(keys.Confirm),
		km.HintLabeled(keys.Up, focusVerb),
	}
	if ringSize > 1 {
		hints = append(hints, km.Hint(keys.FocusNext))
	}
	hints = append(hints,
		km.Hint(actBuy),
		km.Hint(actSell),
		km.Hint(actMaxBuy),
		km.Hint(actOpenTrade),
		km.Hint(actOpenAgents),
		km.Hint(actOpenLayout),
	)
	if m.econ.Get().Level() >= economy.SpecUnlockLevel {
		hints = append(hints, km.Hint(actMaxCall), km.Hint(actMaxPut))
	}
	hints = append(hints, km.Hint(keys.Quit))
	return m.frame().HintLine(hints...)
}

func (m Model) renderNotification() string {
	n, ok := m.notifs.Current()
	if !ok {
		return ""
	}
	return panels.NotificationCard(m.frame(), n)
}

func (m Model) heightTier() layout.Tier { return layout.TierForHeight(m.height) }

func (m Model) panelRows(r layout.TierRows) int {
	return r.At(layout.TierForHeight(m.height))
}

func (m *Model) handlePageScroll(msg tea.KeyMsg) bool {
	body := m.screen.view(m)
	rows := layout.ScrollableRows(body, layout.ContentRows(m.height))
	if rows <= 0 || !layout.FitsScreenWidth(m.width) {
		return false
	}

	total := layout.CountLines(layout.ScrollableBody(body, rows))
	if total <= rows {
		m.pageScroll = 0
		return false
	}

	page := layout.ViewportContentRows(rows)
	if page < 1 {
		page = 1
	}

	s := layout.ScrollState{Offset: m.pageScroll}
	switch msg.String() {
	case "pgdown":
		s.Scroll(page, total, page)
	case "pgup":
		s.Scroll(-page, total, page)
	default:
		return false
	}
	m.pageScroll = s.Offset
	return true
}

func (m *Model) clampPageScroll() {
	body := m.screen.view(m)
	rows := layout.ScrollableRows(body, layout.ContentRows(m.height))
	if rows <= 0 || !layout.FitsScreenWidth(m.width) {
		m.pageScroll = 0
		return
	}

	total := layout.CountLines(layout.ScrollableBody(body, rows))
	if total <= rows {
		m.pageScroll = 0
		return
	}

	page := layout.ViewportContentRows(rows)
	if page < 1 {
		page = 1
	}

	s := layout.ScrollState{Offset: m.pageScroll}
	s.Scroll(0, total, page)
	m.pageScroll = s.Offset
}
