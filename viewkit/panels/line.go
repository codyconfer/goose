package panels

import (
	"strings"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

func Line(f layout.Frame, title string, series []float64, width, height int, fmtVal func(float64) string, footer ...string) string {
	if max := f.BodyWidth() - 7; max > 0 && width > max {
		width = max
	}
	lines := linePlot(series, width, height, fmtVal)
	lines = append(lines, footer...)
	return f.Panel(title, lines...)
}

func linePlot(series []float64, width, height int, fmtVal func(float64) string) []string {
	if len(series) == 0 || width < 1 || height < 1 {
		return nil
	}
	if len(series) > width {
		series = series[len(series)-width:]
	}

	lo, hi := seriesBounds(series)
	span := chartSpan(lo, hi)
	rowOf := func(v float64) int {
		return height - 1 - chartLevel(v, lo, span, height)
	}

	pad := width - len(series)
	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = []rune(strings.Repeat(" ", width))
	}
	for x, v := range series {
		grid[rowOf(v)][pad+x] = '•'
	}

	out := make([]string, 0, height+1)
	for i, row := range grid {
		out = append(out, chartGutter(chartLabel(i, height-1, lo, hi, fmtVal))+theme.Cur().Can.Render(string(row)))
	}
	out = append(out, chartBaseline(width))
	return out
}
