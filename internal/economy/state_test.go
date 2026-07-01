package economy

import (
	"math"
	"testing"
)

func TestEarnAndProduce(t *testing.T) {
	m := NewMachine()
	m.Tap()
	if m.s.Tokens != 1 || m.s.TotalEarned != 1 {
		t.Fatalf("after one press: tokens=%v total=%v, want 1/1", m.s.Tokens, m.s.TotalEarned)
	}

	m.s.Tokens = 1000
	nest := Producers[0]
	if !m.BuyProducer(nest) {
		t.Fatal("expected to afford a nest")
	}
	if got := m.s.TokensPerSecond(); got != nest.TokenRate {
		t.Fatalf("tps=%v, want %v", got, nest.TokenRate)
	}

	before := m.s.Tokens
	m.Produce(1)
	if got := m.s.Tokens - before; math.Abs(got-nest.TokenRate) > 1e-9 {
		t.Fatalf("produce gain=%v, want %v", got, nest.TokenRate)
	}
}

func TestCostScaling(t *testing.T) {
	p := Producers[0]
	if p.CostOf(0) != 15 {
		t.Fatalf("first nest cost=%v, want 15", p.CostOf(0))
	}
	if p.CostOf(1) <= p.CostOf(0) {
		t.Fatal("cost should grow with ownership")
	}
}

func TestClickUpgradeDoubles(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1_000_000
	start := m.s.PerClick
	click, _ := UpgradeByKey(UpgradeClick)
	if !m.BuyUpgrade(click) {
		t.Fatal("expected to afford golden touch")
	}
	if m.s.PerClick != start*2 {
		t.Fatalf("per-click=%v, want %v", m.s.PerClick, start*2)
	}
	if m.s.UpgradeLevel(UpgradeClick) != 1 {
		t.Fatalf("click level=%d, want 1", m.s.UpgradeLevel(UpgradeClick))
	}
}

func TestCannotOverspend(t *testing.T) {
	m := NewMachine()
	if m.BuyProducer(Producers[0]) {
		t.Fatal("bought a nest with no tokens")
	}
}

func TestProduceMakesEggs(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	nest := Producers[0]
	m.BuyProducer(nest)
	m.Produce(1)
	if got := m.s.Eggs; math.Abs(got-nest.EggRate) > 1e-9 {
		t.Fatalf("eggs after 1s=%v, want %v", got, nest.EggRate)
	}
}

func TestConsumersBuyEggs(t *testing.T) {
	m := NewMachine()
	m.s.Eggs = 100
	m.s.Consumers = 10

	sold, earned := m.RunMarket(1)
	if math.Abs(sold-5) > 1e-9 {
		t.Fatalf("sold=%v, want 5", sold)
	}
	if math.Abs(earned-5*BasePrice) > 1e-9 {
		t.Fatalf("earned=%v, want %v", earned, 5*BasePrice)
	}
	if math.Abs(m.s.Eggs-95) > 1e-9 || math.Abs(m.s.Tokens-5*BasePrice) > 1e-9 {
		t.Fatalf("post-market eggs=%v tokens=%v, want 95/%v", m.s.Eggs, m.s.Tokens, 5*BasePrice)
	}
}

func TestConsumerBuyingCappedByStock(t *testing.T) {
	m := NewMachine()
	m.s.Eggs = 3
	m.s.Consumers = 100
	sold, _ := m.RunMarket(1)
	if math.Abs(sold-3) > 1e-9 || m.s.Eggs != 0 {
		t.Fatalf("sold=%v eggs=%v, want 3/0", sold, m.s.Eggs)
	}
}

func TestCrierRaisesSellPrice(t *testing.T) {
	m := NewMachine()
	if m.s.PriceMult() != 1 {
		t.Fatalf("base haggle mult=%v, want 1", m.s.PriceMult())
	}
	base := m.s.SellPrice()
	m.s.Tokens = 1_000_000
	crier, _ := UpgradeByKey(UpgradeCrier)
	if !m.BuyUpgrade(crier) {
		t.Fatal("expected to afford a town crier")
	}
	if math.Abs(m.s.PriceMult()-1.3) > 1e-9 {
		t.Fatalf("haggle mult after crier=%v, want 1.3", m.s.PriceMult())
	}

	if m.s.SellPrice() <= base {
		t.Fatalf("crier sell price %v not above base %v", m.s.SellPrice(), base)
	}
}

func TestMarketCapCountsLaidAndBoughtEggs(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	nest := Producers[0]
	m.BuyProducer(nest)

	m.Produce(10)
	wantLaid := nest.EggRate * 10
	if math.Abs(m.s.EggsLaid-wantLaid) > 1e-9 {
		t.Fatalf("eggs laid=%v, want %v", m.s.EggsLaid, wantLaid)
	}
	if math.Abs(m.s.MarketCap()-wantLaid) > 1e-9 {
		t.Fatalf("cap after laying=%v, want %v", m.s.MarketCap(), wantLaid)
	}

	m.s.Tokens = 1000
	bought, _ := m.fillTrade(TxBuyEggs, 5, m.s.BuyPrice())
	if math.Abs(m.s.MarketCap()-(wantLaid+bought)) > 1e-9 {
		t.Fatalf("cap after buying=%v, want %v", m.s.MarketCap(), wantLaid+bought)
	}

	capBefore := m.s.MarketCap()
	m.s.Eggs = m.s.MarketCap()
	m.s.Consumers = 10
	m.RunMarket(1)
	if m.s.MarketCap() < capBefore {
		t.Fatalf("cap shrank from %v to %v after selling", capBefore, m.s.MarketCap())
	}
}

func TestNormalizeSeedsMarketCapForOldSaves(t *testing.T) {

	s := &State{PeakEggs: 1000, EggsBought: 200}
	Normalize(s)
	if want := 800.0; math.Abs(s.EggsLaid-want) > 1e-9 {
		t.Fatalf("seeded eggs laid=%v, want %v", s.EggsLaid, want)
	}
	if want := 1000.0; math.Abs(s.MarketCap()-want) > 1e-9 {
		t.Fatalf("seeded cap=%v, want %v", s.MarketCap(), want)
	}
}

func TestConsumersDriftTowardSupply(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1e9
	for i := 0; i < 5; i++ {
		m.BuyProducer(Producers[1])
	}

	for i := 0; i < 600; i++ {
		m.UpdateConsumers(0.1)
	}
	if m.s.Consumers <= 0 {
		t.Fatal("consumers never showed up despite egg supply")
	}
}
