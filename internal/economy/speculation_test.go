package economy

import "testing"

func (m *Machine) atPrice(price float64) { m.s.PriceFactor = price / BasePrice }

func TestOpenPositionStakesPremium(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 500
	if !m.OpenPosition(PosCall, 200, 5, 30) {
		t.Fatal("open should succeed with enough tokens")
	}
	if m.s.Tokens != 300 {
		t.Fatalf("premium not deducted: tokens=%v, want 300", m.s.Tokens)
	}
	if len(m.s.Positions) != 1 {
		t.Fatalf("positions=%d, want 1", len(m.s.Positions))
	}
	p := m.s.Positions[0]
	if p.Strike != m.s.EggPrice() {
		t.Fatalf("strike=%v, want spot %v", p.Strike, m.s.EggPrice())
	}
	if len(m.s.Ledger) != 1 || m.s.Ledger[0].Kind != TxOptionOpen {
		t.Fatalf("open not recorded to the ledger: %+v", m.s.Ledger)
	}
}

func TestOpenPositionRejectsUnaffordableOrJunk(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 50
	if m.OpenPosition(PosCall, 200, 5, 30) {
		t.Fatal("should not open a position you can't afford")
	}
	if m.OpenPosition("straddle", 10, 5, 30) {
		t.Fatal("should reject an unknown position kind")
	}
	if m.OpenPosition(PosCall, 10, 0.5, 30) {
		t.Fatal("leverage below 1x is nonsensical")
	}
	if m.s.Tokens != 50 {
		t.Fatalf("failed opens should not touch the purse, tokens=%v", m.s.Tokens)
	}
}

func TestCallPaysOffWhenPriceRises(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	m.atPrice(BasePrice)
	m.OpenPosition(PosCall, 100, 3, 30)
	m.atPrice(BasePrice * 1.2)
	res := m.TickPositions(30)
	if len(res) != 1 {
		t.Fatalf("expected one settlement, got %d", len(res))
	}
	if res[0].PnL <= 0 {
		t.Fatalf("a call should profit on a rally, pnl=%v", res[0].PnL)
	}
	if got, want := res[0].Payout, 160.0; got < want-1e-6 || got > want+1e-6 {
		t.Fatalf("payout=%v, want %v (premium 100 * (1 + 0.6))", got, want)
	}
}

func TestPutPaysOffWhenPriceFalls(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	m.atPrice(BasePrice)
	m.OpenPosition(PosPut, 100, 2, 30)
	m.atPrice(BasePrice * 0.8)
	res := m.TickPositions(30)
	if len(res) != 1 || res[0].PnL <= 0 {
		t.Fatalf("a put should profit on a selloff, got %+v", res)
	}
}

func TestLeveragedPositionGetsMarginCalled(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	m.atPrice(BasePrice)
	m.OpenPosition(PosCall, 100, 5, 60)
	m.atPrice(BasePrice * 0.75)
	res := m.TickPositions(1)
	if len(res) != 1 {
		t.Fatalf("expected a margin call settlement, got %d", len(res))
	}
	if !res[0].MarginCall {
		t.Fatal("underwater leveraged position should be margin-called")
	}
	if res[0].Payout != 0 {
		t.Fatalf("a wiped position pays nothing, payout=%v", res[0].Payout)
	}
	if len(m.s.Positions) != 0 {
		t.Fatal("margin-called position should be removed from the book")
	}
}

func TestUnleveragedPositionSurvivesUntilExpiry(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	m.atPrice(BasePrice)
	m.OpenPosition(PosCall, 100, 1, 30)
	m.atPrice(priceFloor * BasePrice)
	if res := m.TickPositions(1); len(res) != 0 {
		t.Fatalf("1x position should not margin-call early, got %+v", res)
	}
	if len(m.s.Positions) != 1 {
		t.Fatal("1x position should still be open")
	}
}

func TestClosePositionEarlyTakesTheMark(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	m.atPrice(BasePrice)
	m.OpenPosition(PosCall, 100, 2, 60)
	m.atPrice(BasePrice * 1.1)
	before := m.s.Tokens
	res, ok := m.ClosePosition(0)
	if !ok || res.PnL <= 0 {
		t.Fatalf("closing a winner early should book a profit, got %+v ok=%v", res, ok)
	}
	if m.s.Tokens <= before {
		t.Fatal("closing should pay the mark into the purse")
	}
	if len(m.s.Positions) != 0 {
		t.Fatal("closed position should leave the book")
	}
	if _, ok := m.ClosePosition(0); ok {
		t.Fatal("closing an empty book should fail")
	}
}

func TestMarginCallAllLiquidatesLeverageWithHaircut(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	m.atPrice(BasePrice)
	m.OpenPosition(PosCall, 100, 1, 60)
	m.OpenPosition(PosCall, 100, 5, 60)
	m.atPrice(BasePrice * 1.1)
	before := m.s.Tokens
	m.marginCallAll(0.25)
	if len(m.s.Positions) != 1 || m.s.Positions[0].Leverage != 1 {
		t.Fatalf("only leveraged positions should be liquidated, book=%+v", m.s.Positions)
	}
	if m.s.Tokens <= before {
		t.Fatal("liquidation should return the haircut mark to the purse")
	}
}

func TestShockPriceStaysWithinBounds(t *testing.T) {
	m := NewMachine()
	m.shockPrice(100)
	if m.s.PriceFactor > priceCeil+1e-9 {
		t.Fatalf("shock escaped the ceiling: %v", m.s.PriceFactor)
	}
	m.shockPrice(0.0001)
	if m.s.PriceFactor < priceFloor-1e-9 {
		t.Fatalf("shock escaped the floor: %v", m.s.PriceFactor)
	}
}

func TestShockPriceSeedsTrend(t *testing.T) {
	m := NewMachine()
	m.shockPrice(1.3)
	if m.s.PriceTrend <= 0 {
		t.Fatalf("melt-up should leave positive follow-through, trend=%v", m.s.PriceTrend)
	}
	m.shockPrice(0.5)
	if m.s.PriceTrend >= 0 {
		t.Fatalf("flash crash should leave negative follow-through, trend=%v", m.s.PriceTrend)
	}
}

func TestHasLosingLeverageDetectsUnderwaterBets(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	m.atPrice(BasePrice)
	m.OpenPosition(PosCall, 100, 5, 60)
	if m.s.HasLosingLeverage() {
		t.Fatal("a fresh at-the-money position is not yet losing")
	}
	m.atPrice(BasePrice * 0.95)
	if !m.s.HasLosingLeverage() {
		t.Fatal("a leveraged call should be losing after the price dips")
	}
}
