package economy

import (
	"math"
	"math/rand"
)

func (s State) MarketCap() float64 {
	return s.EggsLaid + s.EggsBought
}

func (s State) PriceMult() float64 {
	return 1 + crierBonusPerLevel*float64(s.UpgradeLevel(UpgradeCrier))
}

func (m *Machine) UpdateConsumers(dt float64) {
	target := m.s.EggsPerSecond() / consumerAppetite * crowdHeadroom
	m.s.Consumers += (target - m.s.Consumers) * crowdAdjustRate * dt
	if m.s.Consumers < 0 {
		m.s.Consumers = 0
	}
}

func (m *Machine) growCrowd(factor float64) {
	if m.s.Consumers *= factor; m.s.Consumers < 0 {
		m.s.Consumers = 0
	}
}

func (m *Machine) addConsumers(n float64) {
	if m.s.Consumers += n; m.s.Consumers < 0 {
		m.s.Consumers = 0
	}
}

func (s State) Demand() float64 {
	return s.Consumers * consumerAppetite
}

func (s State) EggPrice() float64 { return BasePrice * s.PriceFactor }

func (s State) BuyPrice() float64 { return s.EggPrice() }

func (s State) SellPrice() float64 { return s.EggPrice() * s.PriceMult() }

func (s State) MarketInterest() float64 {
	avail := s.EggsPerSecond() + s.Eggs*priceHoardOnOffer
	if avail < priceMinSupply {
		avail = priceMinSupply
	}
	return s.Demand() / avail
}

func (s State) priceTarget() float64 {
	target := s.MarketInterest() / crowdHeadroom
	return clampPriceFactor(target)
}

func clampPriceFactor(v float64) float64 {
	if v < priceFloor {
		return priceFloor
	}
	if v > priceCeil {
		return priceCeil
	}
	return v
}

func clampPriceTrend(v float64) float64 {
	if priceTrendMax <= 0 {
		return 0
	}
	if v < -priceTrendMax {
		return -priceTrendMax
	}
	if v > priceTrendMax {
		return priceTrendMax
	}
	return v
}

func clampUnit(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (s State) TrendStrength() float64 {
	if priceTrendMax <= 0 {
		return 0
	}
	return clampUnit(math.Abs(s.PriceTrend) / priceTrendMax)
}

func (s State) trendImpactScale() float64 {
	scale := math.Max(25, s.TokensPerSecond()*60)
	scale = math.Max(scale, s.Tokens*0.25)
	scale = math.Max(scale, s.TotalEarned*0.03)
	scale = math.Max(scale, s.LeveragedExposure()*0.2)
	return scale
}

func trendImpulse(size, scale, coeff float64) float64 {
	if size <= 0 || scale <= 0 || coeff <= 0 {
		return 0
	}
	maxImpulse := priceTrendMax * 0.85
	if maxImpulse <= 0 {
		return 0
	}
	impulse := coeff * math.Log1p(size/scale)
	if impulse > maxImpulse {
		return maxImpulse
	}
	return impulse
}

func (m *Machine) pushTrend(delta float64) {
	if delta == 0 {
		return
	}
	m.s.PriceTrend = clampPriceTrend(m.s.PriceTrend + delta)
}

func (m *Machine) trendFromCashflow(delta float64) {
	if delta == 0 {
		return
	}
	impulse := trendImpulse(math.Abs(delta), m.s.trendImpactScale(), priceCashflowTrend)
	if delta < 0 {
		impulse = -impulse
	}
	m.pushTrend(impulse)
}

func (m *Machine) trendFromTrade(kind TxKind, tokens float64) {
	if tokens <= 0 {
		return
	}
	scale := math.Max(m.s.trendImpactScale(), m.s.Demand()*m.s.EggPrice()*8)
	impulse := trendImpulse(tokens, scale, priceTradeTrend)
	if kind == TxSellEggs {
		impulse = -impulse
	}
	m.pushTrend(impulse)
}

func (m *Machine) trendFromDerivative(p Position, price, pnl, extra float64) {
	move := p.UnderlyingMove(price)
	if move == 0 || p.Premium <= 0 {
		return
	}
	size := math.Abs(pnl) + extra*p.Premium
	scale := math.Max(m.s.trendImpactScale(), p.Notional()*0.2)
	impulse := trendImpulse(size, scale, priceDerivativeTrend)
	if move < 0 {
		impulse = -impulse
	}
	m.pushTrend(impulse)
}

func (m *Machine) UpdatePrice(dt float64, r *rand.Rand) {
	if m.s.PriceFactor <= 0 {
		m.s.PriceFactor = 1
	}
	if dt <= 0 {
		return
	}
	target := m.s.priceTarget()
	trendStrength := m.s.TrendStrength()
	adjustRate := priceAdjustRate / (1 + trendStrength*priceTrendReversionDrag)
	if adjustRate <= 0 {
		adjustRate = 0.01
	}
	decay := math.Exp(-adjustRate * dt)
	trendDecay := math.Exp(-priceTrendDecay * dt)

	var trendShock, priceShock float64
	if r != nil {
		trendVariance := priceTrendVol * priceTrendVol * (1 - trendDecay*trendDecay)
		if trendVariance > 0 {
			trendShock = r.NormFloat64() * math.Sqrt(trendVariance)
		}
		volatility := priceVolatility * (1 + trendStrength*priceTrendTurbulence)
		variance := volatility * volatility * (1 - decay*decay) / (2 * adjustRate)
		if variance > 0 {
			priceShock = r.NormFloat64() * math.Sqrt(variance)
		}
	}

	trend := m.s.PriceTrend
	trendDrift := trend * dt
	if priceTrendDecay > 0 {
		trendDrift = trend * (1 - trendDecay) / priceTrendDecay
	}
	trendDrift *= 1 + trendStrength*0.35
	m.s.PriceTrend = clampPriceTrend(trend*trendDecay + trendShock)
	next := target + (m.s.PriceFactor-target)*decay + trendDrift + priceShock
	clamped := clampPriceFactor(next)
	if clamped != next {
		if (clamped == priceFloor && m.s.PriceTrend < 0) || (clamped == priceCeil && m.s.PriceTrend > 0) {
			m.s.PriceTrend *= 0.35
		}
	}
	m.s.PriceFactor = clamped
}
