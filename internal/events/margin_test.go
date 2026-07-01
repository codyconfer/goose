package events

import (
	"math/rand"
	"testing"

	"github.com/codyconfer/goose/internal/economy"
)

func losingLeverage(t *testing.T) *economy.Machine {
	t.Helper()
	s := economy.NewState()
	s.Tokens = 1000
	m := economy.FromState(s)
	if !m.OpenPosition(economy.PosCall, 100, 5, 60) {
		t.Fatal("setup: failed to open leveraged position")
	}
	m2 := economy.FromState(m.Get())

	st := m2.Get()
	st.PriceFactor *= 0.9
	return economy.FromState(st)
}

func TestMarginTriggerNeedsLosingLeverage(t *testing.T) {
	tr := MarginTrigger{Chance: 1}
	if !tr.Repeatable() {
		t.Fatal("margin trigger should be repeatable")
	}
	r := rand.New(rand.NewSource(1))

	flat := economy.NewState()
	if tr.Fires(flat, r) {
		t.Fatal("margin trigger fired with no positions")
	}

	m := losingLeverage(t)
	if !tr.Fires(m.Get(), r) {
		t.Fatal("margin trigger did not fire on an underwater leveraged position")
	}
}

func TestMarginCallEventLiquidatesLeverage(t *testing.T) {
	var margin *Event
	for i := range Events {
		if Events[i].Key == "margin_call" {
			margin = &Events[i]
		}
	}
	if margin == nil {
		t.Fatal("margin_call event is not registered")
	}

	m := losingLeverage(t)
	out := margin.Apply(m.Get(), rand.New(rand.NewSource(1)))
	m.ApplyWindfall(out.Notif.Title, out.Cmds)
	if len(m.Get().Positions) != 0 {
		t.Fatalf("margin call should have liquidated the leveraged book, left %d", len(m.Get().Positions))
	}
}
