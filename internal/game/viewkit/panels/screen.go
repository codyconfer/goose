package panels

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/goose/internal/game/viewkit/theme"
)

func FitsScreenWidth(screenWidth int) bool {
	return screenWidth <= 0 || screenWidth >= theme.MinScreenWidth
}

func ScreenFrame(screenWidth int) Frame {
	if screenWidth <= 0 {
		return DefaultFrame()
	}
	return NewFrame(screenWidth - theme.ScreenPaddingWidth)
}

func TooNarrow(screenWidth int) string {
	current := "unknown"
	if screenWidth > 0 {
		current = fmt.Sprintf("%d", screenWidth)
	}

	width := theme.MinScreenWidth - theme.ScreenPaddingWidth
	if screenWidth > 0 {
		width = max(screenWidth-theme.AppMarginX*2, 1)
	}

	title := theme.TitleSty.Render(ansi.Truncate("TERMINAL TOO NARROW", width, "…"))
	subtitle := theme.DimSty.Render(ansi.Truncate(fmt.Sprintf("Need at least %d columns.", theme.MinScreenWidth), width, "…"))
	body := lipgloss.NewStyle().Width(width).Render(
		fmt.Sprintf("Current width: %s columns. Resize the terminal to at least %d characters wide to use this screen.", current, theme.MinScreenWidth),
	)
	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle, body)
}
