package economy

import "testing"

func TestBuyProducerRecordsLedger(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	p := Producers[0]
	if !m.BuyProducer(p) {
		t.Fatal("expected to afford a producer")
	}
	if len(m.s.Ledger) != 1 {
		t.Fatalf("ledger len=%d, want 1", len(m.s.Ledger))
	}
	e := m.s.Ledger[0]
	if e.Kind != TxBuyProducer || e.Label != p.Name || e.Tokens >= 0 {
		t.Fatalf("bad ledger entry: %+v", e)
	}
}

func TestDecommissionProducerRefundsAndRecords(t *testing.T) {
	m := NewMachine()
	p := Producers[0]
	m.s.Owned[p.Key] = 2
	before := m.s.Tokens
	refund := p.RefundOf(2)
	if !m.DecommissionProducer(p) {
		t.Fatal("expected to decommission an owned producer")
	}
	if m.s.Owned[p.Key] != 1 {
		t.Fatalf("owned=%d, want 1", m.s.Owned[p.Key])
	}
	if m.s.Tokens != before+refund {
		t.Fatalf("tokens=%v, want %v", m.s.Tokens, before+refund)
	}
	last := m.s.Ledger[len(m.s.Ledger)-1]
	if last.Kind != TxSellProducer || last.Tokens != refund {
		t.Fatalf("bad ledger entry: %+v", last)
	}
}

func TestDecommissionProducerNoneOwned(t *testing.T) {
	m := NewMachine()
	if m.DecommissionProducer(Producers[0]) {
		t.Fatal("decommissioned with none owned")
	}
	if len(m.s.Ledger) != 0 {
		t.Fatal("recorded a no-op decommission")
	}
}

func TestUpgradeRecordsLedger(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	u := Upgrades[0]
	if !m.BuyUpgrade(u) {
		t.Fatal("expected to afford the upgrade")
	}
	last := m.s.Ledger[len(m.s.Ledger)-1]
	if last.Kind != TxUpgrade || last.Label != u.Name || last.Tokens >= 0 {
		t.Fatalf("bad ledger entry: %+v", last)
	}
}

func TestRecordWindfallSkipsZero(t *testing.T) {
	m := NewMachine()
	m.recordWindfall("🍀 Lucky", 50)
	m.recordWindfall("nothing", 0)
	if len(m.s.Ledger) != 1 {
		t.Fatalf("ledger len=%d, want 1", len(m.s.Ledger))
	}
	if e := m.s.Ledger[0]; e.Kind != TxWindfall || e.Tokens != 50 {
		t.Fatalf("bad windfall entry: %+v", e)
	}
}

func TestLedgerCapsToMax(t *testing.T) {
	m := NewMachine()
	for i := 0; i < ledgerMax+10; i++ {
		m.recordWindfall("x", 1)
	}
	if len(m.s.Ledger) != ledgerMax {
		t.Fatalf("ledger len=%d, want %d", len(m.s.Ledger), ledgerMax)
	}
}

func TestCompletedEggOrderRecordsLedger(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 10_000
	m.s.Consumers = 1000
	m.ScheduleTrade(TxBuyEggs, 20)
	for i := 0; i < 5 && len(m.s.Transactions) > 0; i++ {
		m.ProcessTransactions(100)
	}
	if len(m.s.Transactions) != 0 {
		t.Fatal("order did not complete")
	}
	last := m.s.Ledger[len(m.s.Ledger)-1]
	if last.Kind != TxBuyEggs || last.Amount != 20 || last.Tokens >= 0 {
		t.Fatalf("bad ledger entry for completed buy: %+v", last)
	}
}
