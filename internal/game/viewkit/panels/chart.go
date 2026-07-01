package panels

import (
	"fmt"
	"strings"

	"github.com/codyconfer/goose/internal/game/viewkit/theme"
)

const (
	axisLabelWidth = 5
	axisGutter     = axisLabelWidth + 1
)

func seriesBounds(vals []float64) (lo, hi float64) {
	lo, hi = vals[0], vals[0]
	for _, v := range vals {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	return lo, hi
}

func chartSpan(lo, hi float64) float64 {
	span := hi - lo
	if span < 1e-9 {
		span = 1
	}
	return span
}

func chartLevel(v, lo, span float64, cells int) int {
	l := int((v-lo)/span*float64(cells-1) + 0.5)
	if l < 0 {
		l = 0
	}
	if l > cells-1 {
		l = cells - 1
	}
	return l
}

func chartLabel(i, last int, lo, hi float64, fmtVal func(float64) string) string {
	switch i {
	case 0:
		return fmt.Sprintf("%*s", axisLabelWidth, fmtVal(hi))
	case last:
		return fmt.Sprintf("%*s", axisLabelWidth, fmtVal(lo))
	}
	return strings.Repeat(" ", axisLabelWidth)
}

func chartGutter(label string) string {
	return theme.DimSty.Render(label + " │")
}

func chartBaseline(width int) string {
	return theme.DimSty.Render(strings.Repeat(" ", axisGutter) + "└" + strings.Repeat("─", width))
}
