package game

import (
	"testing"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
)

func TestTradeFocusRingScrollsFocusedPanel(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 100000
	econ := economy.FromState(s)
	for i := 1; i <= 12; i++ {
		econ.ScheduleTrade(economy.TxBuyEggs, float64(i*10))
	}
	m := New(econ, events.NewMachine(), 0)
	ts := &tradeScreen{prev: &gameScreen{}, kind: economy.TxBuyEggs}
	m.screen = ts

	if ts.focusedPanel(&m) != "builder" {
		t.Fatalf("default focus = %q, want builder", ts.focusedPanel(&m))
	}
	m = send(m, key("down"))
	ts = m.screen.(*tradeScreen)
	if ts.queue.Offset != 0 {
		t.Fatalf("scrolling the builder must not move the queue offset (%d)", ts.queue.Offset)
	}

	m = send(m, key("tab"))
	ts = m.screen.(*tradeScreen)
	if ts.focusedPanel(&m) != "queue" {
		t.Fatalf("after tab focus = %q, want queue", ts.focusedPanel(&m))
	}
	m = send(m, key("down"))
	ts = m.screen.(*tradeScreen)
	if ts.queue.Offset != 1 {
		t.Fatalf("focused queue offset = %d, want 1 after one down", ts.queue.Offset)
	}

	m = send(m, key("tab"))
	ts = m.screen.(*tradeScreen)
	if ts.focusedPanel(&m) != "roster" {
		t.Fatalf("after third tab focus = %q, want roster", ts.focusedPanel(&m))
	}

	m = send(m, key("tab"))
	ts = m.screen.(*tradeScreen)
	if ts.focusedPanel(&m) != "builder" {
		t.Fatalf("tab should wrap back to builder, got %q", ts.focusedPanel(&m))
	}
}
