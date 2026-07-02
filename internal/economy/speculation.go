package economy

import "math"

type PosKind string

const (
	PosCall PosKind = "call"

	PosPut PosKind = "put"
)

func (k PosKind) valid() bool { return k == PosCall || k == PosPut }

type Position struct {
	Kind     PosKind `json:"kind"`
	Strike   float64 `json:"strike"`
	Premium  float64 `json:"premium"`
	Leverage float64 `json:"leverage"`
	Expiry   float64 `json:"expiry"`
}

func (p Position) Notional() float64 { return p.Premium * p.Leverage }

func (p Position) UnderlyingMove(price float64) float64 {
	if p.Strike <= 0 {
		return 0
	}
	return (price - p.Strike) / p.Strike
}

func (p Position) returnFrac(price float64) float64 {
	move := p.UnderlyingMove(price)
	if p.Kind == PosPut {
		move = -move
	}
	return move * p.Leverage
}

func (p Position) Value(price float64) float64 {
	if v := p.Premium * (1 + p.returnFrac(price)); v > 0 {
		return v
	}
	return 0
}

func (p Position) PnL(price float64) float64 { return p.Value(price) - p.Premium }

func (p Position) MarginFrac(price float64) float64 {
	if p.Premium <= 0 {
		return 0
	}
	return p.Value(price) / p.Premium
}

func (p Position) marginCalled(price float64) bool {
	return p.Leverage > 1 && p.MarginFrac(price) <= SpecMaintenanceMargin
}

func (p Position) WipeMove() float64 {
	if p.Leverage <= 1 {
		return 0
	}
	return 1 / p.Leverage
}

func (p Position) MarginCallMove() float64 {
	if p.Leverage <= 1 {
		return 0
	}
	return (1 - SpecMaintenanceMargin) / p.Leverage
}

func (p Position) LiquidationPrice() float64 {
	if p.Strike <= 0 {
		return 0
	}
	move := p.MarginCallMove()
	if move <= 0 {
		return 0
	}
	if p.Kind == PosPut {
		return p.Strike * (1 + move)
	}
	return p.Strike * (1 - move)
}

func (p Position) MarginBuffer(price float64) float64 {
	liq := p.LiquidationPrice()
	if price <= 0 || liq <= 0 || p.Leverage <= 1 {
		return 0
	}
	if p.Kind == PosPut {
		return (liq - price) / price
	}
	return (price - liq) / price
}

func (m *Machine) OpenPosition(kind PosKind, premium, leverage, expiry float64) bool {
	if !kind.valid() || premium <= 0 || leverage < 1 || expiry <= 0 {
		return false
	}
	price := m.s.EggPrice()
	if price <= 0 {
		return false
	}
	// Calls and puts may be opened on credit — the premium can push tokens negative.
	m.s.Tokens -= premium
	p := Position{Kind: kind, Strike: price, Premium: premium, Leverage: leverage, Expiry: expiry}
	m.s.Positions = append(m.s.Positions, p)
	m.record(Transaction{Kind: TxOptionOpen, Label: p.Desc(), Amount: leverage, Tokens: -premium})
	return true
}

type PosResult struct {
	Pos        Position
	Payout     float64
	PnL        float64
	MarginCall bool
}

func (m *Machine) settle(p Position, price float64, marginCall bool, penalty float64) PosResult {
	payout := p.Value(price)
	if marginCall {
		if penalty < 0 {
			penalty = 0
		}
		if penalty > 1 {
			penalty = 1
		}
		payout *= 1 - penalty
	}
	m.s.Tokens += payout
	m.record(Transaction{Kind: TxOptionSettle, Label: p.Desc(), Amount: p.Leverage, Tokens: payout})
	pnl := payout - p.Premium
	extra := 0.0
	if marginCall {
		extra = p.Leverage
	}
	m.trendFromDerivative(p, price, pnl, extra)
	return PosResult{Pos: p, Payout: payout, PnL: pnl, MarginCall: marginCall}
}

func (m *Machine) TickPositions(dt float64) []PosResult {
	if len(m.s.Positions) == 0 || dt <= 0 {
		return nil
	}
	price := m.s.EggPrice()
	var results []PosResult
	kept := m.s.Positions[:0]
	for _, p := range m.s.Positions {
		p.Expiry -= dt
		switch {
		case p.marginCalled(price):
			results = append(results, m.settle(p, price, true, SpecMarginPenalty))
		case p.Expiry <= 0:
			results = append(results, m.settle(p, price, false, 0))
		default:
			kept = append(kept, p)
		}
	}
	m.s.Positions = kept
	return results
}

func (m *Machine) ClosePosition(i int) (PosResult, bool) {
	if i < 0 || i >= len(m.s.Positions) {
		return PosResult{}, false
	}
	p := m.s.Positions[i]
	res := m.settle(p, m.s.EggPrice(), false, 0)
	m.s.Positions = append(m.s.Positions[:i], m.s.Positions[i+1:]...)
	return res, true
}

func (m *Machine) CloseAllPositions() int {
	price := m.s.EggPrice()
	n := len(m.s.Positions)
	for _, p := range m.s.Positions {
		m.settle(p, price, false, 0)
	}
	m.s.Positions = nil
	return n
}

func (m *Machine) marginCallAll(penalty float64) {
	if len(m.s.Positions) == 0 {
		return
	}
	if penalty < 0 {
		penalty = 0
	}
	if penalty > 1 {
		penalty = 1
	}
	price := m.s.EggPrice()
	kept := m.s.Positions[:0]
	for _, p := range m.s.Positions {
		if p.Leverage <= 1 {
			kept = append(kept, p)
			continue
		}
		m.settle(p, price, true, penalty)
	}
	m.s.Positions = kept
}

func MarginPenaltyPct() float64 { return SpecMarginPenalty * 100 }

func (m *Machine) shockPrice(factor float64) {
	if factor <= 0 {
		return
	}
	m.s.PriceFactor = clampPriceFactor(m.s.PriceFactor * factor)
	m.s.PriceTrend = clampPriceTrend(m.s.PriceTrend + math.Log(factor)*priceShockTrend)
}

func (s State) LeveragedExposure() float64 {
	var total float64
	for _, p := range s.Positions {
		if p.Leverage > 1 {
			total += p.Notional()
		}
	}
	return total
}

func (s State) HasLosingLeverage() bool {
	price := s.EggPrice()
	for _, p := range s.Positions {
		if p.Leverage > 1 && p.PnL(price) < 0 {
			return true
		}
	}
	return false
}

func (s State) MarginStress() float64 {
	price := s.EggPrice()
	stress := 0.0
	for _, p := range s.Positions {
		if p.Leverage <= 1 {
			continue
		}
		buffer := p.MarginBuffer(price)
		switch {
		case p.marginCalled(price):
			return 1
		case buffer <= 0:
			if stress < 1 {
				stress = 1
			}
		case buffer < 0.2:
			score := 1 - buffer/0.2
			if score > stress {
				stress = score
			}
		}
	}
	return stress
}

func (p Position) Desc() string {
	word := "Call"
	if p.Kind == PosPut {
		word = "Put"
	}
	return FormatNum(p.Leverage) + "x " + word + " @ " + FormatNum(p.Strike)
}
