package game

import (
	"time"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/panels"

	"github.com/codyconfer/goose/internal/content"
)

func renderClock(m *Model, vk layout.Frame) string {
	label, loc := clockZone()
	return panels.Clock(vk, content.Text.Clock.Panel, m.clock.last,
		panels.ClockOpts{TwentyFour: true, ShowDate: true, Zones: []panels.ClockZone{
			{Label: "LOCAL", Loc: time.Local},
			{Label: label, Loc: loc},
		}})
}

func renderBinaryClock(m *Model, vk layout.Frame) string {
	return panels.BinaryClock(vk, content.Text.Clock.BinaryPanel, m.clock.last)
}
