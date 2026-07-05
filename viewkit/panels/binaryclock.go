package panels

import (
	"strings"
	"time"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

const (
	binOn  = "●"
	binOff = "○"
)

var binWeights = [4]int{8, 4, 2, 1}

func BinaryClock(f layout.Frame, title string, t time.Time) string {
	h, m, s := t.Hour(), t.Minute(), t.Second()
	digits := [6]int{h / 10, h % 10, m / 10, m % 10, s / 10, s % 10}

	acc, dim := theme.Cur().Accent, theme.Cur().Dim

	rows := make([]string, 0, len(binWeights)+1)
	for _, w := range binWeights {
		var b strings.Builder
		for c, d := range digits {
			b.WriteString(colGap(c))

			if w <= maxForColumn(c) && d&w != 0 {
				b.WriteString(acc.Render(binOn))
			} else {
				b.WriteString(dim.Render(binOff))
			}
		}
		rows = append(rows, b.String())
	}

	var foot strings.Builder
	for c, d := range digits {
		foot.WriteString(colGap(c))
		foot.WriteString(intGlyph(d))
	}
	rows = append(rows, dim.Render(foot.String()))

	return f.Panel(title, rows...)
}

func maxForColumn(col int) int {
	switch col {
	case 0:
		return 2
	case 2, 4:
		return 7
	default:
		return 15
	}
}

func intGlyph(d int) string {
	return string(rune('0' + d))
}

func colGap(col int) string {
	switch {
	case col == 0:
		return ""
	case col%2 == 0:
		return "  "
	default:
		return " "
	}
}
