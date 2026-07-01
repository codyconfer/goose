package panels

import (
	"strings"

	"github.com/codyconfer/goose/internal/game/viewkit/theme"
)

type OHLC struct{ Open, High, Low, Close float64 }

func Candle(title string, candles []OHLC, width, height int, fmtVal func(float64) string, footer ...string) string {
	return DefaultFrame().Candle(title, candles, width, height, fmtVal, footer...)
}

func (f Frame) Candle(title string, candles []OHLC, width, height int, fmtVal func(float64) string, footer ...string) string {
	if max := f.bodyWidth() - 7; max > 0 && width > max {
		width = max
	}
	lines := candlePlot(candles, width, height, fmtVal)
	lines = append(lines, footer...)
	return f.Panel(title, lines...)
}

func candlePlot(candles []OHLC, width, height int, fmtVal func(float64) string) []string {
	if len(candles) == 0 || width < 1 || height < 1 {
		return nil
	}
	if len(candles) > width {
		candles = candles[len(candles)-width:]
	}

	lo, hi := candles[0].Low, candles[0].High
	for _, c := range candles {
		if c.Low < lo {
			lo = c.Low
		}
		if c.High > hi {
			hi = c.High
		}
	}

	body := candleRows(candles, width, height, lo, hi)
	lines := make([]string, 0, height+1)
	for i, row := range body {
		lines = append(lines, chartGutter(chartLabel(i, height-1, lo, hi, fmtVal))+row)
	}
	lines = append(lines, chartBaseline(width))
	return lines
}

func candleRows(candles []OHLC, width, height int, lo, hi float64) []string {
	span := chartSpan(lo, hi)
	levels := height * 2
	lvl := func(p float64) int {
		return chartLevel(p, lo, span, levels)
	}

	pad := width - len(candles)
	if pad < 0 {
		pad = 0
	}
	rows := make([]string, height)
	for row := 0; row < height; row++ {
		cell := height - 1 - row
		upper, lower := 2*cell+1, 2*cell
		var b strings.Builder
		if pad > 0 {
			b.WriteString(strings.Repeat(" ", pad))
		}
		for _, c := range candles {
			hiL, loL := lvl(c.High), lvl(c.Low)
			top, bot := lvl(c.Open), lvl(c.Close)
			if bot > top {
				top, bot = bot, top
			}
			sty := theme.CanSty
			if c.Close < c.Open {
				sty = theme.CantSty
			}
			bodyU := upper >= bot && upper <= top
			bodyL := lower >= bot && lower <= top
			switch {
			case bodyU && bodyL:
				b.WriteString(sty.Render("█"))
			case bodyU:
				b.WriteString(sty.Render("▀"))
			case bodyL:
				b.WriteString(sty.Render("▄"))
			case (upper >= loL && upper <= hiL) || (lower >= loL && lower <= hiL):
				b.WriteString(sty.Render("│"))
			default:
				b.WriteByte(' ')
			}
		}
		rows[row] = b.String()
	}
	return rows
}
