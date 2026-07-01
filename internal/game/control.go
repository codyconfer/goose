package game

import (
	"fmt"

	"github.com/codyconfer/goose/internal/control"
	"github.com/codyconfer/goose/internal/notify"
)

type ControlMsg control.Message

func (m *Model) applyControl(msg ControlMsg) {
	if !m.screen.simulates() {
		return
	}
	label := msg.Label
	if label == "" {
		label = "🎛️ External command"
	}
	if len(msg.Econ) > 0 {
		m.econ.ApplyWindfall(label, msg.Econ)
	}
	if len(msg.Events) > 0 {
		m.events.Apply(msg.Events)
	}
	m.notifs.Push(notify.Notification{
		Title:   label,
		Message: fmt.Sprintf("Applied %d economy and %d events command(s) from another terminal.", len(msg.Econ), len(msg.Events)),
		Tone:    notify.ToneNeutral,
	}, notifBeats)
}
