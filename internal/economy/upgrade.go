package economy

import (
	"fmt"
	"math"

	"github.com/codyconfer/goose/internal/content"
)

const (
	UpgradeClick = "click"
	UpgradeCrier = "crier"
	UpgradeBlitz = "blitz"
)

type Upgrade struct {
	content.Upgrade

	Apply func(m *Machine)

	Unlocked func(s State) bool

	BoughtMsg func(s State) string
}

type upgradeBehavior struct {
	Apply     func(m *Machine)
	Unlocked  func(s State) bool
	BoughtMsg func(s State) string
}

var upgradeBehaviors = map[string]upgradeBehavior{
	UpgradeClick: {
		Apply: func(m *Machine) { m.s.PerClick *= 2 },
		BoughtMsg: func(s State) string {
			return fmt.Sprintf("🧘 Flow state achieved. You now earn %s tokens per tap.", FormatNum(s.PerClick))
		},
	},
	UpgradeCrier: {
		Unlocked: func(s State) bool { return s.EggsPerSecond() > 0 || s.UpgradeLevel(UpgradeCrier) > 0 },
		BoughtMsg: func(s State) string {
			return fmt.Sprintf("🧥 Jet Set Huang zips up the magical jacket! Consumers now pay %.0f%% more per egg.", (s.PriceMult()-1)*100)
		},
	},
	UpgradeBlitz: {
		Unlocked: func(s State) bool { return s.TokensPerSecond() > 0 || s.UpgradeLevel(UpgradeBlitz) > 0 },
		BoughtMsg: func(s State) string {
			return fmt.Sprintf("🌀 Blitzscaling! Every producer now outputs %.0f%% more. Unit economics remain a Q4 problem.", (s.ProdMult()-1)*100)
		},
	},
}

var Upgrades = func() []Upgrade {
	us := make([]Upgrade, len(content.Upgrades))
	for i, c := range content.Upgrades {
		b := upgradeBehaviors[c.Key]
		us[i] = Upgrade{Upgrade: c, Apply: b.Apply, Unlocked: b.Unlocked, BoughtMsg: b.BoughtMsg}
	}
	return us
}()

func UpgradeByKey(key string) (Upgrade, bool) {
	for _, u := range Upgrades {
		if u.Key == key {
			return u, true
		}
	}
	return Upgrade{}, false
}

func (u Upgrade) CostOf(s State) float64 {
	return math.Ceil(u.BaseCost * math.Pow(u.Growth, float64(s.UpgradeLevel(u.Key))))
}

func (u Upgrade) IsUnlocked(s State) bool {
	return u.Unlocked == nil || u.Unlocked(s)
}

func (s State) UpgradeLevel(key string) int { return s.UpgradeLevels[key] }

func (m *Machine) BuyUpgrade(u Upgrade) bool {
	cost := u.CostOf(m.s)
	if m.s.Tokens < cost {
		return false
	}
	m.s.Tokens -= cost
	m.s.UpgradeLevels[u.Key]++
	if u.Apply != nil {
		u.Apply(m)
	}
	m.record(Transaction{Kind: TxUpgrade, Label: u.Name, Amount: 1, Tokens: -cost})
	return true
}
