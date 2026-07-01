package events

import (
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
)

func (m *Machine) Roll(econ economy.State, r *rand.Rand) (Outcome, bool) {
	for _, i := range r.Perm(len(Events)) {
		e := Events[i]
		if !e.Trigger.Repeatable() && m.s.HasFired(e.Key) {
			continue
		}
		if e.CanFire != nil && !e.CanFire(econ) {
			continue
		}
		if e.Trigger.Fires(econ, r) {
			if !e.Trigger.Repeatable() {
				m.markFired(e.Key)
			}
			return e.Apply(econ, r), true
		}
	}
	return Outcome{}, false
}

func (m *Machine) Reconcile(econ economy.State) {
	for _, e := range Events {
		if !e.Trigger.Repeatable() && !m.s.HasFired(e.Key) && e.Trigger.Fires(econ, nil) {
			m.markFired(e.Key)
		}
	}
}
