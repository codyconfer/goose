package game

import (
	"strings"
	"testing"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
)

func leveledState() economy.State {
	s := economy.NewState()
	s.Tokens = 100000
	s.PeakEggs = economy.LevelThresholds[economy.SpecUnlockLevel-1]
	return s
}

func focusDeskPane(t *testing.T, m *Model, ts *tradeScreen, name string) {
	t.Helper()
	ring := ts.focusables(m)
	for i, p := range ring {
		if p == name {
			ts.focus = i
			return
		}
	}
	t.Fatalf("pane %q not focusable; ring=%v", name, ring)
}

func TestTradeDeskHidesDerivativesWhenLocked(t *testing.T) {
	m := New(economy.NewMachine(), events.NewMachine(), 0)
	m.screen = &tradeScreen{prev: &gameScreen{}, kind: economy.TxBuyEggs}
	if v := m.View(); strings.Contains(v, "WRITE A CONTRACT") {
		t.Fatalf("derivatives ticket should be hidden below the unlock level:\n%s", v)
	}
}

func TestTradeDeskShowsDerivativesWhenUnlocked(t *testing.T) {
	m := New(economy.FromState(leveledState()), events.NewMachine(), 0)
	m.screen = &tradeScreen{prev: &gameScreen{}, kind: economy.TxBuyEggs}
	if v := m.View(); !strings.Contains(v, "WRITE A CONTRACT") {
		t.Fatalf("derivatives ticket should be visible once unlocked:\n%s", v)
	}
}

func TestDeskOpensACall(t *testing.T) {
	m := New(economy.FromState(leveledState()), events.NewMachine(), 0)
	ts := &tradeScreen{prev: &gameScreen{}, pos: economy.PosCall}
	m.screen = ts
	focusDeskPane(t, &m, ts, "ticket")
	m = send(m, key("enter"))
	if len(m.econ.Get().Positions) != 1 {
		t.Fatalf("open positions=%d, want 1", len(m.econ.Get().Positions))
	}
	if m.econ.Get().Positions[0].Kind != economy.PosCall {
		t.Fatalf("opened %v, want a call", m.econ.Get().Positions[0].Kind)
	}
}

func TestDeskTogglesToPut(t *testing.T) {
	m := New(economy.FromState(leveledState()), events.NewMachine(), 0)
	ts := &tradeScreen{prev: &gameScreen{}, pos: economy.PosCall}
	m.screen = ts
	focusDeskPane(t, &m, ts, "ticket")
	m = send(m, key("right"))
	m = send(m, key("enter"))
	pos := m.econ.Get().Positions
	if len(pos) != 1 || pos[0].Kind != economy.PosPut {
		t.Fatalf("expected one put, got %+v", pos)
	}
}

func TestDeskLeverageAdjusts(t *testing.T) {
	m := New(economy.FromState(leveledState()), events.NewMachine(), 0)
	ts := &tradeScreen{prev: &gameScreen{}, pos: economy.PosCall}
	m.screen = ts
	focusDeskPane(t, &m, ts, "ticket")
	m = send(m, key("]"))
	m = send(m, key("enter"))
	pos := m.econ.Get().Positions
	if len(pos) != 1 || pos[0].Leverage != economy.SpecLeverages[1] {
		t.Fatalf("leverage=%v, want %v", pos, economy.SpecLeverages[1])
	}
}

func TestDeskCloseAll(t *testing.T) {
	s := leveledState()
	econ := economy.FromState(s)
	econ.OpenPosition(economy.PosCall, 50, 2, 60)
	econ.OpenPosition(economy.PosPut, 50, 2, 60)
	m := New(econ, events.NewMachine(), 0)
	ts := &tradeScreen{prev: &gameScreen{}, pos: economy.PosCall}
	m.screen = ts
	focusDeskPane(t, &m, ts, "positions")
	m = send(m, key("c"))
	if len(m.econ.Get().Positions) != 0 {
		t.Fatalf("close-all left %d positions", len(m.econ.Get().Positions))
	}
}

func TestDeskViewRendersDerivatives(t *testing.T) {
	econ := economy.FromState(leveledState())
	econ.OpenPosition(economy.PosCall, 50, 5, 60)
	m := New(econ, events.NewMachine(), 0)
	m.screen = &tradeScreen{prev: &gameScreen{}, pos: economy.PosCall}
	m.width, m.height = 80, 120
	v := m.View()
	for _, want := range []string{"TRADE DESK", "WRITE A CONTRACT", "OPEN POSITIONS", "Call", "liq. price", "notional"} {
		if !strings.Contains(v, want) {
			t.Errorf("trade desk view missing %q", want)
		}
	}
}

func TestDeskViewDoesNotMutateDefaultKind(t *testing.T) {
	m := New(economy.FromState(leveledState()), events.NewMachine(), 0)
	ts := &tradeScreen{prev: &gameScreen{}}
	m.screen = ts
	_ = m.View()
	if ts.pos != "" {
		t.Fatalf("view mutated default position kind to %q", ts.pos)
	}
}

func TestDeskScrollsPositionsAndClosesVisibleEntry(t *testing.T) {
	s := leveledState()
	s.Tokens = 1000000
	econ := economy.FromState(s)
	for i := 1; i <= 12; i++ {
		if !econ.OpenPosition(economy.PosCall, float64(i*10), 2, 60) {
			t.Fatalf("failed to open position %d", i)
		}
	}

	m := New(econ, events.NewMachine(), 0)
	ts := &tradeScreen{prev: &gameScreen{}, pos: economy.PosCall}
	m.screen = ts
	focusDeskPane(t, &m, ts, "positions")
	for i := 0; i < 12; i++ {
		m = send(m, key("down"))
	}

	if ts.positions.Offset == 0 {
		t.Fatal("focused positions scroll did not advance the offset")
	}
	if got := m.View(); !strings.Contains(got, "7–12 of 12") {
		t.Fatalf("positions footer missing scrolled range:\n%s", got)
	}

	m = send(m, key("x"))
	if len(m.econ.Get().Positions) != 11 {
		t.Fatalf("positions len=%d, want 11 after close", len(m.econ.Get().Positions))
	}
	if got := m.econ.Get().Positions[4].Premium; got != 50 {
		t.Fatalf("visible close removed wrong position, premium at index 4=%v, want 50", got)
	}
}
