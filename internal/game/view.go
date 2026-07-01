package game

import (
	"fmt"
	"math"

	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/game/viewkit/panels"
	"github.com/codyconfer/goose/internal/game/viewkit/theme"
	"github.com/codyconfer/goose/internal/notify"
)

func (m Model) View() string {
	if m.quitting {
		return theme.AppFrame.Render(theme.TitleSty.Render(content.Text.App.Quit))
	}
	return theme.AppFrame.Render(m.screen.view(&m))
}

func (m Model) frame() panels.Frame {
	if m.width <= 0 {
		return panels.DefaultFrame()
	}
	return panels.NewFrame(m.width - 8)
}

func (m Model) renderTitleBar() string {
	return m.frame().Header(content.Text.App.Title, content.Text.App.Subtitle)
}

func (m Model) renderStatus() string {
	vk := m.frame()
	s := m.econ.Get()
	lvl := s.Level()

	tokens := vk.Spread(
		theme.EggSty.Render("🪙 "+economy.FormatNum(s.Tokens))+theme.DimSty.Render(" tokens"),
		theme.DimSty.Render(fmt.Sprintf(content.Text.Status.RateFmt, economy.FormatNum(s.TokensPerSecond()), economy.FormatNum(s.PerClick))),
	)
	eggs := vk.Spread(
		theme.EggSty.Render("🥚 "+economy.FormatNum(s.Eggs))+theme.DimSty.Render(" eggs"),
		theme.TitleSty.Render(fmt.Sprintf(content.Text.Status.LevelFmt, lvl)),
	)

	var progress string
	if next, ok := s.NextLevelEggs(); ok {
		le := s.LevelEggs()
		prev := s.LevelFloor()
		frac := (le - prev) / (next - prev)
		progress = panels.Meter(frac, meterWidth(vk.Width, 22)) +
			theme.DimSty.Render("  "+fmt.Sprintf(content.Text.Status.ProgressFmt, economy.FormatNum(le), economy.FormatNum(next), lvl+1))
	} else {
		progress = theme.EggSty.Render("★ ") + theme.DimSty.Render(content.Text.Status.MaxLevel)
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
		caption = theme.EggSty.Render(fmt.Sprintf("+%s 🪙", economy.FormatNum(m.econ.Get().PerClick)))
	}
	card := theme.TapCardSty.Width(vk.Width + 2).Render(lipgloss.JoinVertical(lipgloss.Center,
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
		vk.Row(content.Text.Market.StockLabel, theme.EggSty.Render(economy.FormatNum(s.Eggs)+" eggs")),
		vk.Row(content.Text.Market.MarketCapLabel, theme.ValSty.Render(economy.FormatNum(s.MarketCap())+" eggs")),
		vk.Row(content.Text.Market.LayingLabel, theme.CanSty.Render("+"+economy.FormatNum(s.EggsPerSecond())+" /sec")),
		vk.Row(content.Text.Market.SellingLabel, theme.CanSty.Render("+"+economy.FormatNum(m.sellRate)+" /sec")),
		vk.Row(content.Text.Market.ConsumersLabel, theme.ValSty.Render(economy.FormatNum(s.Consumers))),
		vk.Row(content.Text.Market.PriceLabel, theme.ValSty.Render(economy.FormatNum(s.SellPrice())+" tokens/egg ")+priceTag),
	)
}

func (m Model) renderFooter() string {
	return m.frame().HintLine(
		[2]string{"enter", "generate"},
		[2]string{"↑/↓", "select"},
		[2]string{"b/→", "buy"},
		[2]string{"s", "sell"},
		[2]string{"t", "trade"},
		[2]string{"q", "quit"},
	)
}

func (m Model) renderNotification() string {
	n, ok := m.notifs.Current()
	if !ok {
		return ""
	}
	return notificationCard(n.Title, n.Message, n.Tone, m.frame().Width)
}

func notificationCard(title, message string, tone notify.Tone, width int) string {
	sty := theme.NotifNeutralSty
	switch tone {
	case notify.TonePositive:
		sty = theme.NotifPositiveSty
	case notify.ToneWarning:
		sty = theme.NotifWarningSty
	case notify.ToneNegative:
		sty = theme.NotifNegativeSty
	}
	body := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render(title),
		lipgloss.NewStyle().Width(width).Render(message),
	)
	return sty.Render(body)
}

func meterWidth(frameWidth, desired int) int {
	if desired < 1 {
		return 1
	}
	max := frameWidth / 3
	if max < 8 {
		max = 8
	}
	if desired > max {
		return max
	}
	return desired
}
