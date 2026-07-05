package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const resetSeq = "\x1b[0m"

func Screen(body string, width, height int) string {
	bg := Cur().Bg
	if bg == "" {
		return body
	}
	if seq := bgSeq(bg); seq != "" {
		body = strings.ReplaceAll(body, resetSeq, resetSeq+seq)
	}
	return lipgloss.NewStyle().Background(bg).Width(width).Height(height).Render(body)
}

func bgSeq(c lipgloss.Color) string {
	s := lipgloss.NewStyle().Background(c).Render("\x00")
	if i := strings.IndexByte(s, 0); i > 0 {
		return s[:i]
	}
	return ""
}
