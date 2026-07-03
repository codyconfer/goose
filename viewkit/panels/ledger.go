package panels

import (
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

type LedgerRow struct {
	Label string
	Delta float64
}

func Ledger(f layout.Frame, title string, rows []LedgerRow, unit string, fmtNum func(float64) string, visible, offset int, empty string) string {
	if len(rows) == 0 {
		return f.Panel(title, theme.Cur().Dim.Render(empty))
	}
	lines := make([]string, len(rows))
	for i, r := range rows {
		lines[i] = f.Spread(theme.Cur().Val.Render(r.Label), delta(r.Delta, unit, fmtNum))
	}
	return f.ScrollPanel(title, lines, visible, offset)
}

func delta(v float64, unit string, fmtNum func(float64) string) string {
	switch {
	case v > 0:
		return theme.Cur().Can.Render("+" + fmtNum(v) + " " + unit)
	case v < 0:
		return theme.Cur().Cant.Render(fmtNum(v) + " " + unit)
	default:
		return theme.Cur().Dim.Render("—")
	}
}
