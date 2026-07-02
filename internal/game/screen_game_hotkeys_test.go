package game

import (
	"strings"
	"testing"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
)

func TestGameScreenQueuesMaxBuyOrder(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 1000

	m := New(economy.FromState(s), events.NewMachine(), 0)
	m = send(m, key("B"))

	orders := m.econ.Get().Transactions
	if len(orders) != 1 {
		t.Fatalf("queue len=%d, want 1", len(orders))
	}
	if got := orders[0]; got.Kind != economy.TxBuyEggs || got.Amount != 400 {
		t.Fatalf("queued %+v, want max buy of 400 eggs", got)
	}
}

func TestGameScreenQueuesMaxSellOrder(t *testing.T) {
	s := economy.NewState()
	s.Eggs = 321

	m := New(economy.FromState(s), events.NewMachine(), 0)
	m = send(m, key("S"))

	orders := m.econ.Get().Transactions
	if len(orders) != 1 {
		t.Fatalf("queue len=%d, want 1", len(orders))
	}
	if got := orders[0]; got.Kind != economy.TxSellEggs || got.Amount != 321 {
		t.Fatalf("queued %+v, want max sell of 321 eggs", got)
	}
}

func TestGameScreenOpensMaxCallPosition(t *testing.T) {
	s := leveledState()
	s.Tokens = 6000

	m := New(economy.FromState(s), events.NewMachine(), 0)
	m = send(m, key("O"))

	pos := m.econ.Get().Positions
	if len(pos) != 1 {
		t.Fatalf("open positions=%d, want 1", len(pos))
	}
	if got := pos[0]; got.Kind != economy.PosCall || got.Premium != 5000 || got.Leverage != 50 {
		t.Fatalf("opened %+v, want 50x call with 5000 premium", got)
	}
}

func TestGameScreenOpensMaxPutPosition(t *testing.T) {
	s := leveledState()
	s.Tokens = 100000

	m := New(economy.FromState(s), events.NewMachine(), 0)
	m = send(m, key("P"))

	pos := m.econ.Get().Positions
	if len(pos) != 1 {
		t.Fatalf("open positions=%d, want 1", len(pos))
	}
	if got := pos[0]; got.Kind != economy.PosPut || got.Premium != 100000 || got.Leverage != 50 {
		t.Fatalf("opened %+v, want 50x put with 100000 premium", got)
	}
}

func TestGameScreenMaxOptionHotkeyStaysLockedBelowSpecLevel(t *testing.T) {
	m := New(economy.NewMachine(), events.NewMachine(), 0)
	m = send(m, key("O"))

	if len(m.econ.Get().Positions) != 0 {
		t.Fatalf("positions=%d, want none while locked", len(m.econ.Get().Positions))
	}
	if !strings.Contains(m.flash, "Derivatives Desk") {
		t.Fatalf("expected derivatives lock flash, got %q", m.flash)
	}
}
