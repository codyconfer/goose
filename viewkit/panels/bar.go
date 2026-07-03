package panels

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

type Datum struct {
	Label string
	Value float64
}

func Bar(f layout.Frame, title string, data []Datum, width int, fmtNum func(float64) string, empty string) string {
	lines, ok := barLines(f, data, width, fmtNum)
	if !ok {
		return f.Panel(title, theme.Cur().Dim.Render(empty))
	}
	return f.Panel(title, lines...)
}

func BarScroll(f layout.Frame, title string, data []Datum, width int, fmtNum func(float64) string, empty string, visible, offset int) string {
	lines, ok := barLines(f, data, width, fmtNum)
	if !ok {
		return f.Panel(title, theme.Cur().Dim.Render(empty))
	}
	return f.ScrollPanel(title, lines, visible, offset)
}

func barLines(f layout.Frame, data []Datum, width int, fmtNum func(float64) string) ([]string, bool) {
	if len(data) == 0 {
		return nil, false
	}
	if width < 1 {
		width = 1
	}
	max, labelW, valueW := 0.0, 0, 0
	for _, d := range data {
		if a := absf(d.Value); a > max {
			max = a
		}
		if w := ansi.StringWidth(d.Label); w > labelW {
			labelW = w
		}
		if w := ansi.StringWidth(fmtNum(d.Value)); w > valueW {
			valueW = w
		}
	}
	if max == 0 {
		max = 1
	}
	available := f.BodyWidth() - labelW - valueW - 2
	if available < 1 {
		available = 1
	}
	if width > available {
		width = available
	}
	lines := make([]string, len(data))
	for i, d := range data {
		n := int(absf(d.Value)/max*float64(width) + 0.5)
		sty := theme.Cur().Can
		if d.Value < 0 {
			sty = theme.Cur().Cant
		}
		label := theme.Cur().Dim.Render(padRight(d.Label, labelW))
		bar := sty.Render(strings.Repeat("█", n))
		lines[i] = f.Spread(label+" "+bar, sty.Render(fmtNum(d.Value)))
	}
	return lines, true
}

func absf(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func padRight(s string, w int) string {
	if gap := w - ansi.StringWidth(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}
