package economy

import "maps"

type State struct {
	Tokens           float64        `json:"tokens"`
	TotalEarned      float64        `json:"total_earned"`
	PerClick         float64        `json:"per_click"`
	Owned            map[string]int `json:"owned"`
	UpgradeLevels    map[string]int `json:"upgrades,omitempty"`
	Eggs             float64        `json:"eggs"`
	PeakEggs         float64        `json:"peak_eggs,omitempty"`
	EggsLaid         float64        `json:"eggs_laid"`
	EggsBought       float64        `json:"eggs_bought"`
	Consumers        float64        `json:"consumers"`
	PriceFactor      float64        `json:"price_factor,omitempty"`
	PriceTrend       float64        `json:"price_trend,omitempty"`
	Transactions     []Transaction  `json:"transactions,omitempty"`
	Positions        []Position     `json:"positions,omitempty"`
	Agents           []Agent        `json:"agents,omitempty"`
	Ledger           []Transaction  `json:"ledger,omitempty"`
	PriceCandles     []PriceCandle  `json:"price_candles,omitempty"`
	PriceCandleBeats int            `json:"price_candle_beats,omitempty"`
	Settings         Settings       `json:"settings,omitzero"`
	LastSeen         int64          `json:"last_seen"`

	FreezeSeconds float64 `json:"freeze_seconds,omitempty"`
	FreezeReason  string  `json:"freeze_reason,omitempty"`
}

func NewState() State {
	return State{
		Tokens:        0,
		PerClick:      1,
		PriceFactor:   1,
		Owned:         map[string]int{},
		UpgradeLevels: map[string]int{},
		Agents:        defaultAgents(),
	}
}

func (s State) clone() State {
	out := s
	out.Owned = make(map[string]int, len(s.Owned))
	maps.Copy(out.Owned, s.Owned)
	out.UpgradeLevels = make(map[string]int, len(s.UpgradeLevels))
	maps.Copy(out.UpgradeLevels, s.UpgradeLevels)
	if s.Transactions != nil {
		out.Transactions = append([]Transaction(nil), s.Transactions...)
	}
	if s.Positions != nil {
		out.Positions = append([]Position(nil), s.Positions...)
	}
	if s.Agents != nil {
		out.Agents = append([]Agent(nil), s.Agents...)
	}
	if s.Ledger != nil {
		out.Ledger = append([]Transaction(nil), s.Ledger...)
	}
	if s.PriceCandles != nil {
		out.PriceCandles = append([]PriceCandle(nil), s.PriceCandles...)
	}
	return out
}

func (m *Machine) Get() State { return m.s.clone() }

func Normalize(s *State) {
	if s.Owned == nil {
		s.Owned = map[string]int{}
	}
	if s.UpgradeLevels == nil {
		s.UpgradeLevels = map[string]int{}
	}
	if s.Agents == nil {
		s.Agents = defaultAgents()
	}
	if s.PerClick <= 0 {
		s.PerClick = 1
	}
	if s.PriceFactor <= 0 {
		s.PriceFactor = 1
	}
	s.PriceFactor = clampPriceFactor(s.PriceFactor)
	s.PriceTrend = clampPriceTrend(s.PriceTrend)
	if s.PriceCandleBeats < 0 {
		s.PriceCandleBeats = 0
	}

	if s.Eggs > s.PeakEggs {
		s.PeakEggs = s.Eggs
	}

	if s.EggsLaid == 0 {
		if laid := s.PeakEggs - s.EggsBought; laid > 0 {
			s.EggsLaid = laid
		}
	}
}
