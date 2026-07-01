package economy

import (
	"math"
	"testing"
)

func TestScheduleTradeValidates(t *testing.T) {
	m := NewMachine()
	if m.ScheduleTrade(TxBuyEggs, 0) {
		t.Error("scheduled a zero-amount order")
	}
	if m.ScheduleTrade("nonsense", 10) {
		t.Error("scheduled an unknown-kind order")
	}
	if !m.ScheduleTrade(TxBuyEggs, 100) {
		t.Fatal("a valid buy order should schedule")
	}
	if len(m.s.Transactions) != 1 {
		t.Fatalf("queue len=%d, want 1", len(m.s.Transactions))
	}
}

func TestProcessBuyOrderFillsOverBeats(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	m.s.Consumers = 4
	m.ScheduleTrade(TxBuyEggs, 100)

	rep := m.ProcessTransactions(1)
	if len(rep.Completed) != 0 {
		t.Fatalf("first beat completed=%d, want 0", len(rep.Completed))
	}
	if math.Abs(rep.BoughtEggs-5) > 1e-9 {
		t.Fatalf("first beat bought %v eggs, want 5", rep.BoughtEggs)
	}
	if math.Abs(m.s.Tokens-(1000-5*BasePrice)) > 1e-9 || math.Abs(m.s.Eggs-5) > 1e-9 {
		t.Fatalf("after beat tokens=%v eggs=%v", m.s.Tokens, m.s.Eggs)
	}
	if math.Abs(m.s.EggsBought-5) > 1e-9 {
		t.Fatalf("eggs bought=%v, want 5", m.s.EggsBought)
	}

	rep = m.ProcessTransactions(100)
	if len(rep.Completed) != 1 {
		t.Fatalf("drain beat completed=%d, want 1", len(rep.Completed))
	}
	if len(m.s.Transactions) != 0 {
		t.Fatalf("completed order still queued: %d left", len(m.s.Transactions))
	}
	if math.Abs(m.s.Eggs-100) > 1e-9 {
		t.Fatalf("final eggs=%v, want 100", m.s.Eggs)
	}
}

func TestBuyOrderStallsWhenBroke(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 0
	m.s.Consumers = 100
	m.ScheduleTrade(TxBuyEggs, 50)
	rep := m.ProcessTransactions(1)
	if rep.BoughtEggs != 0 {
		t.Fatalf("a broke buy order bought %v eggs, want 0", rep.BoughtEggs)
	}
	if len(m.s.Transactions) != 1 {
		t.Fatal("a stalled order should stay queued")
	}
}

func TestProcessSellOrderConvertsEggsToTokens(t *testing.T) {
	m := NewMachine()
	m.gainEggs(100)
	m.s.Consumers = 100
	m.ScheduleTrade(TxSellEggs, 40)

	rep := m.ProcessTransactions(100)
	if len(rep.Completed) != 1 || math.Abs(rep.SoldEggs-40) > 1e-9 {
		t.Fatalf("sell completed=%d sold=%v, want 1/40", len(rep.Completed), rep.SoldEggs)
	}
	if math.Abs(m.s.Eggs-60) > 1e-9 {
		t.Fatalf("eggs after sell=%v, want 60", m.s.Eggs)
	}
	if math.Abs(m.s.Tokens-40*BasePrice) > 1e-9 {
		t.Fatalf("tokens after sell=%v, want %v", m.s.Tokens, 40*BasePrice)
	}
}

func TestSellingEggsNeverDropsLevel(t *testing.T) {
	m := NewMachine()
	m.gainEggs(800)
	if m.s.Level() != 3 {
		t.Fatalf("level=%d, want 3", m.s.Level())
	}
	m.s.Consumers = 1000
	m.ScheduleTrade(TxSellEggs, 800)
	for i := 0; i < 5 && len(m.s.Transactions) > 0; i++ {
		m.ProcessTransactions(100)
	}
	if m.s.Eggs > 1e-6 {
		t.Fatalf("expected hoard drained, eggs=%v", m.s.Eggs)
	}
	if m.s.Level() != 3 {
		t.Fatalf("level after dumping eggs=%d, want 3 (high-water mark)", m.s.Level())
	}
}

func TestProcessQueueRunsEveryOrder(t *testing.T) {

	m := NewMachine()
	m.s.Tokens = 1000
	m.gainEggs(100)
	m.s.Consumers = 100
	m.ScheduleTrade(TxBuyEggs, 10)
	m.ScheduleTrade(TxSellEggs, 10)
	rep := m.ProcessTransactions(100)
	if math.Abs(rep.BoughtEggs-10) > 1e-9 || math.Abs(rep.SoldEggs-10) > 1e-9 {
		t.Fatalf("bought=%v sold=%v, want 10/10", rep.BoughtEggs, rep.SoldEggs)
	}
	if len(rep.Completed) != 2 || len(m.s.Transactions) != 0 {
		t.Fatalf("completed=%d queued=%d, want 2/0", len(rep.Completed), len(m.s.Transactions))
	}
}

func TestCancelAndClearTrade(t *testing.T) {
	m := NewMachine()
	m.ScheduleTrade(TxBuyEggs, 10)
	m.ScheduleTrade(TxSellEggs, 20)
	m.ScheduleTrade(TxBuyEggs, 30)

	if !m.CancelTransaction(1) || len(m.s.Transactions) != 2 {
		t.Fatalf("cancel middle failed, queue=%v", m.s.Transactions)
	}
	if m.s.Transactions[1].Amount != 30 {
		t.Fatalf("wrong order remained: %v", m.s.Transactions[1])
	}
	if m.CancelTransaction(5) {
		t.Error("cancel out of range should report false")
	}
	m.ClearTransactions()
	if len(m.s.Transactions) != 0 {
		t.Fatal("clear left orders behind")
	}
}

func TestEmptyQueueIsNoop(t *testing.T) {
	m := NewMachine()
	rep := m.ProcessTransactions(1)
	if rep.BoughtEggs != 0 || rep.SoldEggs != 0 || len(rep.Completed) != 0 {
		t.Fatalf("empty queue moved something: %+v", rep)
	}
}
