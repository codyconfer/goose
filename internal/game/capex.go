package game

import (
	"fmt"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
)

type capexItem interface {
	name() string
	icon() string
	desc() string
	owned(economy.State) string
	cost(economy.State) float64
	unlocked(economy.State) bool
	buy(*economy.Machine) bool
	boughtMsg(economy.State) string
	deniedMsg() string
	sell(*economy.Machine) bool
	soldMsg(economy.State) string
	cantSellMsg() string
	lockedTeaser(economy.State) (string, bool)
}

func capexItems() []capexItem {
	items := make([]capexItem, 0, len(economy.Upgrades)+len(economy.Producers))
	for i := range economy.Upgrades {
		items = append(items, upgradeItem{&economy.Upgrades[i]})
	}
	for i := range economy.Producers {
		items = append(items, producerItem{&economy.Producers[i]})
	}
	return items
}

type producerItem struct{ p *economy.Producer }

func (it producerItem) name() string                 { return it.p.Name }
func (it producerItem) icon() string                 { return it.p.Icon }
func (it producerItem) desc() string                 { return it.p.Desc }
func (it producerItem) owned(s economy.State) string { return fmt.Sprintf("x%d", s.Count(it.p.Key)) }
func (it producerItem) cost(s economy.State) float64 { return it.p.CostOf(s.Count(it.p.Key)) }

func (it producerItem) unlocked(s economy.State) bool {
	return s.Count(it.p.Key) > 0 || s.Level() >= it.p.UnlockLevel
}
func (it producerItem) buy(m *economy.Machine) bool { return m.BuyProducer(*it.p) }
func (it producerItem) boughtMsg(s economy.State) string {
	return fmt.Sprintf(content.Text.Capex.ProducerBoughtFmt, it.p.Icon, it.p.Name)
}
func (it producerItem) deniedMsg() string {
	return fmt.Sprintf(content.Text.Capex.ProducerDeniedFmt, it.p.Name)
}

func (it producerItem) sell(m *economy.Machine) bool { return m.DecommissionProducer(*it.p) }
func (it producerItem) soldMsg(s economy.State) string {

	refund := it.p.RefundOf(s.Count(it.p.Key) + 1)
	return fmt.Sprintf(content.Text.Capex.ProducerSoldFmt, it.p.Name, economy.FormatNum(refund))
}
func (it producerItem) cantSellMsg() string {
	return fmt.Sprintf(content.Text.Capex.ProducerCantSellFmt, it.p.Name)
}

func (it producerItem) lockedTeaser(s economy.State) (string, bool) {
	if it.unlocked(s) {
		return "", false
	}
	return fmt.Sprintf(content.Text.Capex.LockedTeaserFmt, it.p.Name, it.p.UnlockLevel), true
}

type upgradeItem struct{ u *economy.Upgrade }

func (it upgradeItem) name() string { return it.u.Name }
func (it upgradeItem) icon() string { return it.u.Icon }
func (it upgradeItem) desc() string { return it.u.Desc }
func (it upgradeItem) owned(s economy.State) string {
	return fmt.Sprintf("Lv.%d", s.UpgradeLevel(it.u.Key))
}
func (it upgradeItem) cost(s economy.State) float64  { return it.u.CostOf(s) }
func (it upgradeItem) unlocked(s economy.State) bool { return it.u.IsUnlocked(s) }
func (it upgradeItem) buy(m *economy.Machine) bool   { return m.BuyUpgrade(*it.u) }

func (it upgradeItem) boughtMsg(s economy.State) string {
	if it.u.BoughtMsg != nil {
		return it.u.BoughtMsg(s)
	}
	return fmt.Sprintf(content.Text.Capex.UpgradeBoughtFmt, it.u.Icon, it.u.Name)
}
func (it upgradeItem) deniedMsg() string {
	return fmt.Sprintf(content.Text.Capex.UpgradeDeniedFmt, it.u.Name)
}

func (it upgradeItem) sell(m *economy.Machine) bool   { return false }
func (it upgradeItem) soldMsg(s economy.State) string { return "" }
func (it upgradeItem) cantSellMsg() string            { return content.Text.Capex.UpgradeCantSell }

func (it upgradeItem) lockedTeaser(s economy.State) (string, bool) { return "", false }
