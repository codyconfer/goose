package economy

import (
	"math"
	"testing"
)

func TestCommandsApplyEachEffect(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 100
	m.s.Consumers = 5
	m.s.Owned["server"] = 5
	m.gainEggs(50)

	m.apply([]Command{
		Earn(10),
		Spend(20),
		GrantProducer("gpu", 3),
		GrowCrowd(2),
		AddConsumers(5),
		Trade(TxSellEggs, 10, BasePrice),
		SeizeBest(),
	})

	if math.Abs(m.s.Tokens-(90+10*BasePrice)) > 1e-9 {
		t.Fatalf("tokens=%v, want %v", m.s.Tokens, 90+10*BasePrice)
	}
	if m.s.Owned["gpu"] != 3 {
		t.Fatalf("gpu=%d, want 3", m.s.Owned["gpu"])
	}
	if m.s.Consumers != 15 {
		t.Fatalf("consumers=%v, want 15", m.s.Consumers)
	}
	if math.Abs(m.s.Eggs-40) > 1e-9 {
		t.Fatalf("eggs after sell=%v, want 40", m.s.Eggs)
	}
	if m.s.Owned["server"] != 4 {
		t.Fatalf("server after seize=%d, want 4", m.s.Owned["server"])
	}
}

func TestApplyWindfallRecordsNetSwing(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 100

	m.ApplyWindfall("🍀 Lucky", []Command{Earn(30), Spend(10)})
	if math.Abs(m.s.Tokens-120) > 1e-9 {
		t.Fatalf("tokens=%v, want 120", m.s.Tokens)
	}
	if len(m.s.Ledger) != 1 {
		t.Fatalf("ledger len=%d, want 1", len(m.s.Ledger))
	}
	if e := m.s.Ledger[0]; e.Kind != TxWindfall || e.Label != "🍀 Lucky" || math.Abs(e.Tokens-20) > 1e-9 {
		t.Fatalf("bad windfall entry: %+v", e)
	}
}

func TestApplyWindfallSkipsZeroNet(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 100
	m.s.Owned["server"] = 1

	m.ApplyWindfall("🦈 Predatory", []Command{Earn(50), Spend(50), SeizeBest()})
	if m.s.Owned["server"] != 0 {
		t.Fatalf("server not seized: %d", m.s.Owned["server"])
	}
	if len(m.s.Ledger) != 0 {
		t.Fatalf("zero-net windfall recorded: %+v", m.s.Ledger)
	}
}
