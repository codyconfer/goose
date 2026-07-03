package panels

import (
	"fmt"
	"strings"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

func Pie(f layout.Frame, title string, data []Datum, barWidth int, fmtNum func(float64) string, empty string) string {
	total := 0.0
	for _, d := range data {
		if d.Value > 0 {
			total += d.Value
		}
	}
	if total <= 0 {
		return f.Panel(title, theme.Cur().Dim.Render(empty))
	}
	if barWidth < 1 {
		barWidth = 1
	}
	if barWidth > f.BodyWidth() {
		barWidth = f.BodyWidth()
	}

	var bar strings.Builder
	var legend []string
	filled := 0
	for i, d := range data {
		if d.Value <= 0 {
			continue
		}
		frac := d.Value / total
		sty := theme.Cur().Series[i%len(theme.Cur().Series)]
		n := int(frac*float64(barWidth) + 0.5)
		if filled+n > barWidth {
			n = barWidth - filled
		}
		filled += n
		bar.WriteString(sty.Render(strings.Repeat("█", n)))
		legend = append(legend, f.Spread(
			sty.Render("■ ")+theme.Cur().Val.Render(d.Label),
			theme.Cur().Dim.Render(fmt.Sprintf("%s  ·  %.0f%%", fmtNum(d.Value), frac*100))))
	}

	lines := append([]string{bar.String()}, legend...)
	return f.Panel(title, lines...)
}
