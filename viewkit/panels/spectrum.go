package panels

import (
	"strings"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

var vBlocks = [8]string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

type SpectrumOpts struct {
	Peaks   []float64
	BarGap  int
	BarWide int
}

func Spectrum(f layout.Frame, title string, levels []float64, height int, empty string, opts ...SpectrumOpts) string {
	o := SpectrumOpts{BarWide: 1, BarGap: 1}
	if len(opts) > 0 {
		o = opts[0]
		if o.BarWide < 1 {
			o.BarWide = 1
		}
		if o.BarGap < 0 {
			o.BarGap = 0
		}
	}
	if len(levels) == 0 || height < 1 {
		return f.Panel(title, theme.Cur().Dim.Render(empty))
	}

	span := o.BarWide + o.BarGap
	maxBands := (f.BodyWidth() + o.BarGap) / span
	if maxBands < 1 {
		maxBands = 1
	}
	if len(levels) > maxBands {
		levels = levels[:maxBands]
	}
	peaks := o.Peaks
	if len(peaks) > len(levels) {
		peaks = peaks[:len(levels)]
	}

	series := theme.Cur().Series
	dim := theme.Cur().Dim
	rows := make([]string, height)
	for row := range height {
		cell := height - 1 - row
		var b strings.Builder
		for i, lvl := range levels {
			if i > 0 {
				b.WriteString(strings.Repeat(" ", o.BarGap))
			}
			glyph := spectrumCell(lvl, cell, height)
			switch {
			case glyph != "":
				b.WriteString(series[i%len(series)].Render(strings.Repeat(glyph, o.BarWide)))
			case i < len(peaks) && peakCell(peaks[i], height) == cell:
				b.WriteString(dim.Render(strings.Repeat("▔", o.BarWide)))
			default:
				b.WriteString(strings.Repeat(" ", o.BarWide))
			}
		}
		rows[row] = b.String()
	}
	return f.Panel(title, rows...)
}

func spectrumCell(level float64, cell, height int) string {
	if level < 0 {
		level = 0
	}
	if level > 1 {
		level = 1
	}
	eighths := int(level*float64(height)*8 + 0.5)
	n := eighths - cell*8
	switch {
	case n <= 0:
		return ""
	case n >= 8:
		return "█"
	default:
		return vBlocks[n-1]
	}
}

func peakCell(peak float64, height int) int {
	if peak < 0 {
		peak = 0
	}
	if peak > 1 {
		peak = 1
	}
	eighths := int(peak*float64(height)*8 + 0.5)
	if eighths <= 0 {
		return -1
	}
	return (eighths - 1) / 8
}
