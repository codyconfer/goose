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

func (p Position) returnFrac(price float64) float64 {
	if p.Strike <= 0 {
		return 0
	}
	move := (price - p.Strike) / p.Strike
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

func (p Position) wiped(price float64) bool {
	return p.Leverage > 1 && p.Premium*(1+p.returnFrac(price)) <= 0
}

func (p Position) WipeMove() float64 {
	if p.Leverage <= 1 {
		return 0
	}
	return 1 / p.Leverage
}

func (m *Machine) OpenPosition(kind PosKind, premium, leverage, expiry float64) bool {
	if !kind.valid() || premium <= 0 || leverage < 1 || expiry <= 0 {
		return false
	}
	price := m.s.EggPrice()
	if price <= 0 || premium > m.s.Tokens {
		return false
	}
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

func (m *Machine) settle(p Position, price float64, marginCall bool) PosResult {
	payout := p.Value(price)
	m.s.Tokens += payout
	m.record(Transaction{Kind: TxOptionSettle, Label: p.Desc(), Amount: p.Leverage, Tokens: payout})
	return PosResult{Pos: p, Payout: payout, PnL: payout - p.Premium, MarginCall: marginCall}
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
		case p.wiped(price):
			results = append(results, m.settle(p, price, true))
		case p.Expiry <= 0:
			results = append(results, m.settle(p, price, false))
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
	res := m.settle(p, m.s.EggPrice(), false)
	m.s.Positions = append(m.s.Positions[:i], m.s.Positions[i+1:]...)
	return res, true
}

func (m *Machine) CloseAllPositions() int {
	price := m.s.EggPrice()
	n := len(m.s.Positions)
	for _, p := range m.s.Positions {
		m.settle(p, price, false)
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
		payout := p.Value(price) * (1 - penalty)
		m.s.Tokens += payout
		m.record(Transaction{Kind: TxOptionSettle, Label: p.Desc(), Amount: p.Leverage, Tokens: payout})
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
			total += p.Premium
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

func (p Position) Desc() string {
	word := "Call"
	if p.Kind == PosPut {
		word = "Put"
	}
	return FormatNum(p.Leverage) + "x " + word + " @ " + FormatNum(p.Strike)
}
