package panels

import (
	"strconv"
	"strings"
	"time"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

const (
	binOn  = "●"
	binOff = "○"
)

const binBytes = 4

func BinaryClock(f layout.Frame, title string, t time.Time) string {
	ts := t.Unix()
	if ts < 0 {
		ts = 0
	}

	acc, dim := theme.Cur().Accent, theme.Cur().Dim

	rows := make([]string, 0, binBytes+1)
	for i := binBytes - 1; i >= 0; i-- {
		b := byte(ts >> (uint(i) * 8))
		var line strings.Builder
		for bit := 7; bit >= 0; bit-- {
			if bit != 7 {
				line.WriteString(" ")
			}
			if bit == 3 {
				line.WriteString(" ")
			}
			if b&(1<<uint(bit)) != 0 {
				line.WriteString(acc.Render(binOn))
			} else {
				line.WriteString(dim.Render(binOff))
			}
		}
		rows = append(rows, line.String())
	}

	rows = append(rows, dim.Render(strconv.FormatInt(ts, 10)))

	return f.Panel(title, rows...)
}
