package panels

import (
	"time"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

type ClockZone struct {
	Label string
	Loc   *time.Location
}

type ClockOpts struct {
	TwentyFour bool

	HideSeconds bool

	ShowDate bool

	Zones []ClockZone
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

	if len(o.Zones) > 0 {
		return worldClock(f, title, t, layoutStr, o)
	}

	lines := []string{theme.Cur().Accent.Render(f.Fit(t.Format(layoutStr)))}
	if o.ShowDate {
		lines = append(lines, theme.Cur().Dim.Render(f.Fit(t.Format("Mon Jan 2 2006"))))
	}
	return f.Panel(title, lines...)
}

func worldClock(f layout.Frame, title string, t time.Time, layoutStr string, o ClockOpts) string {
	lines := make([]string, 0, len(o.Zones)+1)
	for _, z := range o.Zones {
		loc := z.Loc
		if loc == nil {
			loc = time.Local
		}
		zt := t.In(loc)
		label := theme.Cur().Dim.Render(z.Label)
		clock := theme.Cur().Accent.Render(zt.Format(layoutStr + " MST"))
		lines = append(lines, f.Spread(label, clock))
	}
	if o.ShowDate {
		loc := o.Zones[0].Loc
		if loc == nil {
			loc = time.Local
		}
		lines = append(lines, theme.Cur().Dim.Render(f.Fit(t.In(loc).Format("Mon Jan 2 2006"))))
	}
	return f.Panel(title, lines...)
}
