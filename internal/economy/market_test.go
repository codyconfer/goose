package economy

import (
	"math/rand"
	"testing"
)

func TestSellPriceCarriesCrierPremium(t *testing.T) {
	m := NewMachine()
	if m.s.BuyPrice() != BasePrice {
		t.Fatalf("buy price=%v, want %v", m.s.BuyPrice(), BasePrice)
	}
	if m.s.SellPrice() != BasePrice {
		t.Fatalf("sell price with no crier=%v, want %v", m.s.SellPrice(), BasePrice)
	}
	m.s.UpgradeLevels[UpgradeCrier] = 1
	if m.s.SellPrice() <= m.s.BuyPrice() {
		t.Fatal("a crier should make the sell price exceed the buy price")
	}
}

func TestMarketInterestRisesWithDemand(t *testing.T) {
	m := NewMachine()
	m.s.Owned["server"] = 10
	low := m.s.MarketInterest()
	m.s.Consumers += 50
	if m.s.MarketInterest() <= low {
		t.Fatal("more consumers should raise market interest")
	}
}

func TestPriceTargetRisesWhenBuyersOutpaceSupply(t *testing.T) {
	m := NewMachine()
	m.s.Owned["server"] = 1
	m.s.Consumers = 500
	if m.s.priceTarget() <= 1.0 {
		t.Fatalf("scarce eggs should pull the fundamental up, target=%v", m.s.priceTarget())
	}
}

func TestPriceTargetFallsOnAGlut(t *testing.T) {
	m := NewMachine()
	m.s.Owned["server"] = 5
	m.s.Eggs = 1_000_000
	m.s.Consumers = 1
	if m.s.priceTarget() >= 1.0 {
		t.Fatalf("an egg glut should pull the fundamental down, target=%v", m.s.priceTarget())
	}
}

func TestPriceMeanRevertsTowardFundamental(t *testing.T) {

	m := NewMachine()
	m.s.Owned["server"] = 1
	m.s.Consumers = 500
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 400; i++ {
		m.UpdatePrice(0.3, r)
	}
	if m.s.PriceFactor <= 1.0 {
		t.Fatalf("price should revert toward a high fundamental, factor=%v", m.s.PriceFactor)
	}
}

func TestPriceIsVolatile(t *testing.T) {

	m := NewMachine()
	m.s.Owned["server"] = 3
	m.s.Consumers = m.s.EggsPerSecond() / consumerAppetite * crowdHeadroom
	r := rand.New(rand.NewSource(1))
	min, max := m.s.PriceFactor, m.s.PriceFactor
	for i := 0; i < 300; i++ {
		m.UpdatePrice(0.3, r)
		if m.s.PriceFactor < min {
			min = m.s.PriceFactor
		}
		if m.s.PriceFactor > max {
			max = m.s.PriceFactor
		}
	}
	if max-min < 0.1 {
		t.Fatalf("price barely moved (range %v); expected visible volatility", max-min)
	}
}

func TestPriceTrendCarriesMovesAcrossRerolls(t *testing.T) {
	oldVol, oldTrendVol := priceVolatility, priceTrendVol
	priceVolatility, priceTrendVol = 0, 0
	defer func() {
		priceVolatility, priceTrendVol = oldVol, oldTrendVol
	}()

	m := NewMachine()
	m.s.Owned["server"] = 3
	m.s.Consumers = m.s.EggsPerSecond() / consumerAppetite * crowdHeadroom
	m.s.PriceTrend = 0.04

	start := m.s.PriceFactor
	for i := 0; i < 3; i++ {
		m.UpdatePrice(3, nil)
	}
	if m.s.PriceFactor <= start {
		t.Fatalf("positive trend should carry price above start, factor=%v start=%v", m.s.PriceFactor, start)
	}
	if m.s.PriceTrend <= 0 {
		t.Fatalf("trend should persist across rerolls, trend=%v", m.s.PriceTrend)
	}
}

func TestPriceStaysWithinBounds(t *testing.T) {
	m := NewMachine()
	m.s.Owned["server"] = 1
	m.s.Consumers = 1e9
	r := rand.New(rand.NewSource(2))
	for i := 0; i < 1000; i++ {
		m.UpdatePrice(0.3, r)
		if m.s.PriceFactor > priceCeil+1e-9 || m.s.PriceFactor < priceFloor-1e-9 {
			t.Fatalf("price factor %v escaped [%v, %v]", m.s.PriceFactor, priceFloor, priceCeil)
		}
	}
}
