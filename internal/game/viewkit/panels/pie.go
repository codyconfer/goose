package panels

import (
	"fmt"
	"strings"

	"github.com/codyconfer/goose/internal/game/viewkit/theme"
)

func Pie(title string, data []Datum, barWidth int, fmtNum func(float64) string, empty string) string {
	return DefaultFrame().Pie(title, data, barWidth, fmtNum, empty)
}

func (f Frame) Pie(title string, data []Datum, barWidth int, fmtNum func(float64) string, empty string) string {
	total := 0.0
	for _, d := range data {
		if d.Value > 0 {
			total += d.Value
		}
	}
	if total <= 0 {
		return f.Panel(title, theme.DimSty.Render(empty))
	}
	if barWidth < 1 {
		barWidth = 1
	}
	if barWidth > f.bodyWidth() {
		barWidth = f.bodyWidth()
	}

	var bar strings.Builder
	var legend []string
	filled := 0
	for i, d := range data {
		if d.Value <= 0 {
			continue
		}
		frac := d.Value / total
		sty := theme.Series[i%len(theme.Series)]
		n := int(frac*float64(barWidth) + 0.5)
		if filled+n > barWidth {
			n = barWidth - filled
		}
		filled += n
		bar.WriteString(sty.Render(strings.Repeat("█", n)))
		legend = append(legend, f.Spread(
			sty.Render("■ ")+theme.ValSty.Render(d.Label),
			theme.DimSty.Render(fmt.Sprintf("%s  ·  %.0f%%", fmtNum(d.Value), frac*100))))
	}

	lines := append([]string{bar.String()}, legend...)
	return f.Panel(title, lines...)
}
