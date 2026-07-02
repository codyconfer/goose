package economy

// AgentMetric is the live market reading an agent watches.
type AgentMetric string

const (
	MetricTrend  AgentMetric = "trend"  // signed trend strength, -1 (bearish) .. +1 (bullish)
	MetricPrice  AgentMetric = "price"  // tokens per egg
	MetricTokens AgentMetric = "tokens" // cash on hand
	MetricEggs   AgentMetric = "eggs"   // egg holdings
)

// AgentCmp is the direction of the trigger comparison.
type AgentCmp string

const (
	CmpAbove AgentCmp = "above"
	CmpBelow AgentCmp = "below"
)

// AgentAction is what the agent does with your money when it fires.
type AgentAction string

const (
	ActBuyEggs  AgentAction = "buy-eggs"
	ActSellEggs AgentAction = "sell-eggs"
	ActOpenCall AgentAction = "open-call"
	ActOpenPut  AgentAction = "open-put"
)

// Agent is a hands-off trading rule: when Metric Cmp Threshold holds on the slow
// beat, it runs Action with Size. The strategy (metric/cmp/action) is fixed per
// agent — its personality — while the player toggles it and tunes Size/Threshold.
type Agent struct {
	Key       string      `json:"key"`
	Enabled   bool        `json:"enabled"`
	Metric    AgentMetric `json:"metric"`
	Cmp       AgentCmp    `json:"cmp"`
	Threshold float64     `json:"threshold"`
	Action    AgentAction `json:"action"`
	Size      float64     `json:"size"` // egg quantity for trades; premium for options
}

// TradesOptions reports whether the agent opens derivatives (vs. spot egg trades).
func (a Agent) TradesOptions() bool {
	return a.Action == ActOpenCall || a.Action == ActOpenPut
}

func (a Agent) metricValue(s State) float64 {
	switch a.Metric {
	case MetricTrend:
		return s.SignedTrendStrength()
	case MetricPrice:
		return s.EggPrice()
	case MetricTokens:
		return s.Tokens
	case MetricEggs:
		return s.Eggs
	default:
		return 0
	}
}

func (a Agent) fires(s State) bool {
	v := a.metricValue(s)
	if a.Cmp == CmpBelow {
		return v < a.Threshold
	}
	return v > a.Threshold
}

// SignedTrendStrength maps the price trend onto -1 (max bearish) .. +1 (max bullish).
func (s State) SignedTrendStrength() float64 {
	st := s.TrendStrength()
	if s.PriceTrend < 0 {
		return -st
	}
	return st
}

// defaultAgents is the roster every game ships with, all disabled until the
// player hands one the company card.
func defaultAgents() []Agent {
	return []Agent{
		{Key: "momentum", Metric: MetricTrend, Cmp: CmpAbove, Threshold: 0.3, Action: ActBuyEggs, Size: 100},
		{Key: "dipbuyer", Metric: MetricPrice, Cmp: CmpBelow, Threshold: 2.0, Action: ActBuyEggs, Size: 100},
		{Key: "profittaker", Metric: MetricPrice, Cmp: CmpAbove, Threshold: 3.0, Action: ActSellEggs, Size: 100},
		{Key: "degencall", Metric: MetricTrend, Cmp: CmpAbove, Threshold: 0.3, Action: ActOpenCall, Size: 50},
		{Key: "degenput", Metric: MetricTrend, Cmp: CmpBelow, Threshold: -0.3, Action: ActOpenPut, Size: 50},
	}
}

// AgentEvent records an agent firing so the UI can flash what it just did.
type AgentEvent struct {
	Key    string
	Action AgentAction
	Size   float64
}

// RunAgents evaluates every enabled agent against the current state and executes
// the ones whose trigger holds. Meant to be called on the slow beat.
func (m *Machine) RunAgents() []AgentEvent {
	if m.s.Frozen() {
		return nil
	}
	var events []AgentEvent
	for i := range m.s.Agents {
		a := m.s.Agents[i]
		if !a.Enabled || !a.fires(m.s) {
			continue
		}
		if a.TradesOptions() && m.s.Level() < SpecUnlockLevel {
			continue
		}
		var ok bool
		switch a.Action {
		case ActBuyEggs:
			ok = m.ScheduleTrade(TxBuyEggs, a.Size)
		case ActSellEggs:
			ok = m.ScheduleTrade(TxSellEggs, a.Size)
		case ActOpenCall:
			ok = m.OpenPosition(PosCall, a.Size, 1, SpecExpirySeconds)
		case ActOpenPut:
			ok = m.OpenPosition(PosPut, a.Size, 1, SpecExpirySeconds)
		}
		if ok {
			events = append(events, AgentEvent{Key: a.Key, Action: a.Action, Size: a.Size})
		}
	}
	return events
}

// ToggleAgent flips whether the agent at index i is trading.
func (m *Machine) ToggleAgent(i int) {
	if i < 0 || i >= len(m.s.Agents) {
		return
	}
	m.s.Agents[i].Enabled = !m.s.Agents[i].Enabled
}

// SetAgentSize sets the trade/premium size for the agent at index i.
func (m *Machine) SetAgentSize(i int, size float64) {
	if i < 0 || i >= len(m.s.Agents) || size <= 0 {
		return
	}
	m.s.Agents[i].Size = size
}

// SetAgentThreshold sets the trigger threshold for the agent at index i.
func (m *Machine) SetAgentThreshold(i int, t float64) {
	if i < 0 || i >= len(m.s.Agents) {
		return
	}
	m.s.Agents[i].Threshold = t
}
