package economy

import "testing"

func TestApplyOfflineCreditsAndCaps(t *testing.T) {
	m := NewMachine()
	m.s.Owned["server"] = 10
	rate := m.s.TokensPerSecond()

	earned := m.ApplyOffline(60)
	if got := earned; got != rate*60 {
		t.Fatalf("offline earned=%v, want %v", got, rate*60)
	}

	m2 := NewMachine()
	m2.s.Owned["server"] = 10
	capped := m2.ApplyOffline(10 * maxOfflineSeconds)
	if capped != rate*float64(maxOfflineSeconds) {
		t.Fatalf("capped offline=%v, want %v", capped, float64(maxOfflineSeconds)*rate)
	}
}

func TestApplyOfflineNoopForNonPositive(t *testing.T) {
	m := NewMachine()
	m.s.Owned["server"] = 5
	before := m.s.Tokens
	if got := m.ApplyOffline(0); got != 0 || m.s.Tokens != before {
		t.Fatalf("offline for 0s changed state: earned=%v tokens=%v", got, m.s.Tokens)
	}
}

func TestApplyOfflineAdvancesTrendDrivenPrice(t *testing.T) {
	m := NewMachine()
	m.s.Owned["server"] = 1
	m.s.Consumers = 500
	m.s.PriceTrend = 0.08

	before := m.s.PriceFactor
	m.ApplyOffline(10)
	if m.s.PriceFactor <= before {
		t.Fatalf("offline price did not advance with bullish trend: before=%v after=%v", before, m.s.PriceFactor)
	}
}

func TestApplyOfflineProcessesQueuedTrades(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	m.s.Consumers = 100
	if !m.ScheduleTrade(TxBuyEggs, 20) {
		t.Fatal("setup: failed to queue offline trade")
	}

	m.ApplyOffline(10)
	if len(m.s.Transactions) != 0 {
		t.Fatalf("offline should have worked the queue, %d orders remain", len(m.s.Transactions))
	}
	if m.s.EggsBought < 20-1e-6 {
		t.Fatalf("offline buy order did not fill: eggs bought=%v", m.s.EggsBought)
	}
}
