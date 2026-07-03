package game

import (
	"slices"
	"testing"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
)

func TestFocusNamesFiltersInteractive(t *testing.T) {
	got := focusNames(
		focusable{"a", true},
		focusable{"b", false},
		focusable{"c", true},
	)
	if !slices.Equal(got, []string{"a", "c"}) {
		t.Fatalf("focusNames = %v, want [a c]", got)
	}
}

func TestFocusStepWrapsAndResolveClamps(t *testing.T) {
	names := []string{"a", "c"}
	if got := focusStep(names, 0, 1); got != 1 {
		t.Errorf("step forward = %d, want 1", got)
	}
	if got := focusStep(names, 1, 1); got != 0 {
		t.Errorf("step past the end should wrap to 0, got %d", got)
	}
	if got := focusStep(names, 0, -1); got != 1 {
		t.Errorf("step before the start should wrap to last, got %d", got)
	}
	if got := focusResolve(names, 99); got != "c" {
		t.Errorf("resolve clamps out-of-range idx, got %q", got)
	}
	if got := focusResolve(nil, 0); got != "" {
		t.Errorf("resolve of an empty ring should be empty, got %q", got)
	}
}

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
	if ts.focusedPanel(&m) != "builder" {
		t.Fatalf("tab should wrap back to builder, got %q", ts.focusedPanel(&m))
	}
}
