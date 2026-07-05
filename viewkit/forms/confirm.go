package forms

import (
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/panels"
	"github.com/codyconfer/viewkit/theme"
)

type Result int

const (
	Pending Result = iota

	Submitted

	Cancelled
)

type Confirm struct {
	Title    string
	Message  string
	YesLabel string
	NoLabel  string
	Yes      bool
}

func (c Confirm) labels() (yes, no string) {
	yes, no = c.YesLabel, c.NoLabel
	if yes == "" {
		yes = "Yes"
	}
	if no == "" {
		no = "No"
	}
	return yes, no
}

func (c *Confirm) Handle(a keys.Action) Result {
	switch a {
	case keys.Left, keys.Dec:
		c.Yes = true
	case keys.Right, keys.Inc:
		c.Yes = false
	case keys.Confirm:
		return Submitted
	case keys.Cancel:
		return Cancelled
	}
	return Pending
}

func (c Confirm) Render(f layout.Frame) string {
	yes, no := c.labels()
	lines := []string{}
	if c.Message != "" {
		lines = append(lines, theme.Cur().Val.Render(f.Fit(c.Message)))
		lines = append(lines, "")
	}
	lines = append(lines, panels.Toggle(yes, no, c.Yes))
	return f.Panel(c.Title, lines...)
}

func (c Confirm) Overlay(bg string, f layout.Frame, pos ...layout.OverlayPos) string {
	return layout.Overlay(bg, c.Render(f), pos...)
}
