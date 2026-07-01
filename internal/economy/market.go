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

func (m *Machine) UpdatePrice(dt float64, r *rand.Rand) {
	if m.s.PriceFactor <= 0 {
		m.s.PriceFactor = 1
	}
	if dt <= 0 {
		return
	}
	target := m.s.priceTarget()
	decay := math.Exp(-priceAdjustRate * dt)
	trendDecay := math.Exp(-priceTrendDecay * dt)

	var trendShock, priceShock float64
	if r != nil {
		trendVariance := priceTrendVol * priceTrendVol * (1 - trendDecay*trendDecay)
		if trendVariance > 0 {
			trendShock = r.NormFloat64() * math.Sqrt(trendVariance)
		}
		variance := priceVolatility * priceVolatility * (1 - decay*decay) / (2 * priceAdjustRate)
		if variance > 0 {
			priceShock = r.NormFloat64() * math.Sqrt(variance)
		}
	}

	trend := m.s.PriceTrend
	trendDrift := trend * dt
	if priceTrendDecay > 0 {
		trendDrift = trend * (1 - trendDecay) / priceTrendDecay
	}
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
