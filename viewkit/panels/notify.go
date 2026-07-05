package panels

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/notify"
	"github.com/codyconfer/viewkit/theme"
)

func toneStyle(tone notify.Tone) lipgloss.Style {
	t := theme.Cur()
	switch tone {
	case notify.TonePositive:
		return t.NotifPositive
	case notify.ToneWarning:
		return t.NotifWarning
	case notify.ToneNegative:
		return t.NotifNegative
	default:
		return t.NotifNeutral
	}
}

func toneGlyph(tone notify.Tone) string {
	switch tone {
	case notify.TonePositive:
		return "✓"
	case notify.ToneWarning:
		return "!"
	case notify.ToneNegative:
		return "✕"
	default:
		return "•"
	}
}

func NotificationToast(f layout.Frame, n notify.Notification) string {
	line := toneGlyph(n.Tone) + " " + n.Title
	if n.Message != "" {
		line += " — " + n.Message
	}
	return toneStyle(n.Tone).Render(f.Fit(line))
}

func NotificationPanel(f layout.Frame, title string, ns []notify.Notification) string {
	if len(ns) == 0 {
		return f.Panel(title, theme.Cur().Dim.Render("no notifications"))
	}
	t := theme.Cur()
	lines := make([]string, 0, len(ns))
	for _, n := range ns {
		sty := toneStyle(n.Tone)
		head := sty.GetForeground()
		marker := lipgloss.NewStyle().Foreground(head).Render(toneGlyph(n.Tone) + " ")
		lines = append(lines, f.Fit(marker+t.NotifTitle.Render(n.Title)))
		if n.Message != "" {
			lines = append(lines, f.Fit(t.Dim.Render("  "+n.Message)))
		}
	}
	return f.Panel(title, lines...)
}

func NotificationOverlay(bg string, f layout.Frame, n notify.Notification, pos ...layout.OverlayPos) string {
	return layout.Overlay(bg, NotificationCard(f, n), pos...)
}
