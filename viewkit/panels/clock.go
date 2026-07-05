package panels

import (
	"time"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

type ClockOpts struct {
	TwentyFour bool

	HideSeconds bool

	ShowDate bool
}

func Clock(f layout.Frame, title string, t time.Time, opts ...ClockOpts) string {
	o := ClockOpts{TwentyFour: true}
	if len(opts) > 0 {
		o = opts[0]
	}

	layoutStr := "15:04:05"
	switch {
	case o.TwentyFour && o.HideSeconds:
		layoutStr = "15:04"
	case !o.TwentyFour && o.HideSeconds:
		layoutStr = "3:04 PM"
	case !o.TwentyFour:
		layoutStr = "3:04:05 PM"
	}

	lines := []string{theme.Cur().Accent.Render(f.Fit(t.Format(layoutStr)))}
	if o.ShowDate {
		lines = append(lines, theme.Cur().Dim.Render(f.Fit(t.Format("Mon Jan 2 2006"))))
	}
	return f.Panel(title, lines...)
}
