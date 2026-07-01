package events

import (
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
)

type Trigger interface {
	Fires(s economy.State, r *rand.Rand) bool
	Repeatable() bool
}

type ChanceTrigger struct{ P float64 }

func (t ChanceTrigger) Fires(s economy.State, r *rand.Rand) bool {
	return r.Float64() < t.P*s.Settings.EventMult()
}
func (t ChanceTrigger) Repeatable() bool { return true }

type LevelTrigger struct{ Level int }

func (t LevelTrigger) Fires(s economy.State, r *rand.Rand) bool { return s.Level() >= t.Level }
func (t LevelTrigger) Repeatable() bool                         { return false }

type MarketCapTrigger struct{ Cap float64 }

func (t MarketCapTrigger) Fires(s economy.State, r *rand.Rand) bool { return s.MarketCap() >= t.Cap }
func (t MarketCapTrigger) Repeatable() bool                         { return false }

type EggPriceTrigger struct {
	High, Low float64
	Chance    float64
}

func (t EggPriceTrigger) Fires(s economy.State, r *rand.Rand) bool {
	p := s.EggPrice()
	outside := (t.High > 0 && p >= t.High) || (t.Low > 0 && p <= t.Low)
	return outside && r.Float64() < t.Chance
}
func (t EggPriceTrigger) Repeatable() bool { return true }

type MarginTrigger struct{ Chance float64 }

func (t MarginTrigger) Fires(s economy.State, r *rand.Rand) bool {
	stress := s.MarginStress()
	if stress <= 0 {
		return false
	}
	if t.Chance >= 1 {
		return true
	}
	if r == nil {
		return false
	}
	chance := t.Chance * (0.45 + 0.55*stress) * (0.75 + 0.25*s.TrendStrength())
	if chance > 1 {
		chance = 1
	}
	return r.Float64() < chance
}
func (t MarginTrigger) Repeatable() bool { return true }
