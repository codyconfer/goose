package economy

import (
	"fmt"
	"math"

	"github.com/codyconfer/goose/internal/content"
)

type Producer struct {
	content.Producer
}

var Producers = func() []Producer {
	ps := make([]Producer, len(content.Producers))
	for i, c := range content.Producers {
		ps[i] = Producer{c}
	}
	return ps
}()

func (p Producer) CostOf(owned int) float64 {
	return math.Ceil(p.BaseCost * math.Pow(costGrowth, float64(owned)))
}

func (s State) Count(key string) int { return s.Owned[key] }

func (m *Machine) grantProducer(key string, n int) { m.s.Owned[key] += n }

func (s State) ProdMult() float64 {
	return 1 + blitzBonusPerLevel*float64(s.UpgradeLevel(UpgradeBlitz))
}

func (s State) TokensPerSecond() float64 {
	var total float64
	for _, p := range Producers {
		total += p.TokenRate * float64(s.Owned[p.Key])
	}
	return total * s.ProdMult() * s.Settings.MarketMult()
}

func (s State) EggsPerSecond() float64 {
	var total float64
	for _, p := range Producers {
		total += p.EggRate * float64(s.Owned[p.Key])
	}
	return total * s.ProdMult() * s.Settings.MarketMult()
}

func (m *Machine) BuyProducer(p Producer) bool {
	cost := p.CostOf(m.s.Owned[p.Key])
	if m.s.Tokens < cost {
		return false
	}
	m.s.Tokens -= cost
	m.s.Owned[p.Key]++
	m.record(Transaction{Kind: TxBuyProducer, Label: p.Name, Amount: 1, Tokens: -cost})
	if p.Singularity {
		m.s.Tokens = Singularity
		m.s.Eggs = Singularity
	}
	return true
}

func (p Producer) RefundOf(owned int) float64 {
	if owned <= 0 {
		return 0
	}
	return math.Ceil(p.CostOf(owned-1) * decommissionRefund)
}

func (m *Machine) DecommissionProducer(p Producer) bool {
	owned := m.s.Owned[p.Key]
	if owned <= 0 {
		return false
	}
	refund := p.RefundOf(owned)
	m.s.Owned[p.Key]--
	m.s.Tokens += refund
	m.record(Transaction{Kind: TxSellProducer, Label: p.Name, Amount: 1, Tokens: refund})
	return true
}

func (s State) BestProducer() (Producer, bool) {
	best := 0
	var found Producer
	for _, p := range Producers {
		if c := s.Owned[p.Key]; c > best {
			best = c
			found = p
		}
	}
	if best == 0 {
		return Producer{}, false
	}
	return found, true
}

func (m *Machine) seizeBest() {
	if p, ok := m.s.BestProducer(); ok {
		m.s.Owned[p.Key]--
	}
}

const (
	Singularity        = math.MaxFloat64
	singularityDisplay = 1e18
)

func FormatNum(n float64) string {
	abs := math.Abs(n)
	if math.IsInf(n, 0) || abs >= singularityDisplay {
		if n < 0 {
			return "-∞"
		}
		return "∞"
	}
	switch {
	case abs >= 1e15:
		return fmt.Sprintf("%.2fQ", n/1e15)
	case abs >= 1e12:
		return fmt.Sprintf("%.2fT", n/1e12)
	case abs >= 1e9:
		return fmt.Sprintf("%.2fB", n/1e9)
	case abs >= 1e6:
		return fmt.Sprintf("%.2fM", n/1e6)
	case abs >= 1e3:
		return fmt.Sprintf("%.2fK", n/1e3)
	case abs >= 100:
		return fmt.Sprintf("%.0f", n)
	default:
		return fmt.Sprintf("%.1f", n)
	}
}
