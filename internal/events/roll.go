package events

import (
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/outcome"
	"github.com/codyconfer/goose/internal/world"
)

func (m *Machine) Roll(catalog []world.Event, econ economy.State, r *rand.Rand) (outcome.Outcome, bool) {
	for _, i := range r.Perm(len(catalog)) {
		ev := catalog[i]
		if !ev.Trigger.Repeatable() && m.s.HasFired(ev.Key) {
			continue
		}
		if !ev.Eligible(econ) {
			continue
		}
		if ev.Trigger.Fires(econ, r) {
			if !ev.Trigger.Repeatable() {
				m.markFired(ev.Key)
			}
			return ev.Apply(econ, r), true
		}
	}
	return outcome.Outcome{}, false
}

func (m *Machine) Reconcile(catalog []world.Event, econ economy.State) {
	for _, ev := range catalog {
		if !ev.Trigger.Repeatable() && !m.s.HasFired(ev.Key) && ev.Eligible(econ) && ev.Trigger.Fires(econ, nil) {
			m.markFired(ev.Key)
		}
	}
}
