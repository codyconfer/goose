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

func TestSpecDeskLockedBelowLevel(t *testing.T) {
	m := New(economy.NewMachine(), events.NewMachine(), 0)
	m = send(m, key("t"))
	m = send(m, key("d"))
	if _, ok := m.screen.(*specScreen); ok {
		t.Fatal("derivatives desk should stay locked below the unlock level")
	}
	if !strings.Contains(m.flash, "Derivatives Desk") {
		t.Fatalf("expected a locked flash, got %q", m.flash)
	}
}

func TestSpecDeskOpensWhenUnlocked(t *testing.T) {
	m := New(economy.FromState(leveledState()), events.NewMachine(), 0)
	m = send(m, key("t"))
	m = send(m, key("d"))
	ss, ok := m.screen.(*specScreen)
	if !ok {
		t.Fatalf("pressing d opened %T, want *specScreen", m.screen)
	}
	if ss.prev == nil {
		t.Fatal("spec desk forgot the screen it came from")
	}
	m = send(m, key("esc"))
	if _, ok := m.screen.(*tradeScreen); !ok {
		t.Fatalf("esc returned to %T, want *tradeScreen", m.screen)
	}
}

func TestSpecDeskOpensACall(t *testing.T) {
	m := New(economy.FromState(leveledState()), events.NewMachine(), 0)
	m.screen = &specScreen{prev: &gameScreen{}, kind: economy.PosCall}
	m = send(m, key("enter"))
	if len(m.econ.Get().Positions) != 1 {
		t.Fatalf("open positions=%d, want 1", len(m.econ.Get().Positions))
	}
	if m.econ.Get().Positions[0].Kind != economy.PosCall {
		t.Fatalf("opened %v, want a call", m.econ.Get().Positions[0].Kind)
	}
}

func TestSpecDeskTogglesToPut(t *testing.T) {
	m := New(economy.FromState(leveledState()), events.NewMachine(), 0)
	m.screen = &specScreen{prev: &gameScreen{}, kind: economy.PosCall}
	m = send(m, key("right"))
	m = send(m, key("enter"))
	pos := m.econ.Get().Positions
	if len(pos) != 1 || pos[0].Kind != economy.PosPut {
		t.Fatalf("expected one put, got %+v", pos)
	}
}

func TestSpecDeskLeverageAdjusts(t *testing.T) {
	m := New(economy.FromState(leveledState()), events.NewMachine(), 0)
	ss := &specScreen{prev: &gameScreen{}, kind: economy.PosCall}
	m.screen = ss
	m = send(m, key("]"))
	m = send(m, key("enter"))
	pos := m.econ.Get().Positions
	if len(pos) != 1 || pos[0].Leverage != economy.SpecLeverages[1] {
		t.Fatalf("leverage=%v, want %v", pos, economy.SpecLeverages[1])
	}
}

func TestSpecDeskCloseAll(t *testing.T) {
	s := leveledState()
	econ := economy.FromState(s)
	econ.OpenPosition(economy.PosCall, 50, 2, 60)
	econ.OpenPosition(economy.PosPut, 50, 2, 60)
	m := New(econ, events.NewMachine(), 0)
	m.screen = &specScreen{prev: &gameScreen{}, kind: economy.PosCall}
	m = send(m, key("c"))
	if len(m.econ.Get().Positions) != 0 {
		t.Fatalf("close-all left %d positions", len(m.econ.Get().Positions))
	}
}

func TestSpecDeskViewRenders(t *testing.T) {
	econ := economy.FromState(leveledState())
	econ.OpenPosition(economy.PosCall, 50, 5, 60)
	m := New(econ, events.NewMachine(), 0)
	m.screen = &specScreen{prev: &gameScreen{}, kind: economy.PosCall}
	v := m.View()
	for _, want := range []string{"DERIVATIVES DESK", "WRITE A CONTRACT", "OPEN POSITIONS", "Call"} {
		if !strings.Contains(v, want) {
			t.Errorf("spec desk view missing %q", want)
		}
	}
}

func TestSpecDeskViewDoesNotMutateDefaultKind(t *testing.T) {
	m := New(economy.FromState(leveledState()), events.NewMachine(), 0)
	ss := &specScreen{prev: &gameScreen{}}
	m.screen = ss
	_ = m.View()
	if ss.kind != "" {
		t.Fatalf("view mutated default kind to %q", ss.kind)
	}
}

func TestSpecDeskSimulates(t *testing.T) {
	if !(&specScreen{}).simulates() {
		t.Error("derivatives desk should keep the heartbeat running")
	}
}
