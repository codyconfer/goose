package panels

import "github.com/codyconfer/goose/internal/game/viewkit/theme"

type LedgerRow struct {
	Label string
	Delta float64
}

func Ledger(title string, rows []LedgerRow, unit string, fmtNum func(float64) string, visible, offset int, empty string) string {
	return DefaultFrame().Ledger(title, rows, unit, fmtNum, visible, offset, empty)
}

func (f Frame) Ledger(title string, rows []LedgerRow, unit string, fmtNum func(float64) string, visible, offset int, empty string) string {
	if len(rows) == 0 {
		return f.Panel(title, theme.DimSty.Render(empty))
	}
	lines := make([]string, len(rows))
	for i, r := range rows {
		lines[i] = f.Spread(theme.ValSty.Render(r.Label), delta(r.Delta, unit, fmtNum))
	}
	return f.ScrollPanel(title, lines, visible, offset)
}

func delta(v float64, unit string, fmtNum func(float64) string) string {
	switch {
	case v > 0:
		return theme.CanSty.Render("+" + fmtNum(v) + " " + unit)
	case v < 0:
		return theme.CantSty.Render(fmtNum(v) + " " + unit)
	default:
		return theme.DimSty.Render("—")
	}
}
