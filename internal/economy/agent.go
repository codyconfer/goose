package economy

type AgentMetric string

const (
	MetricTrend  AgentMetric = "trend"
	MetricPrice  AgentMetric = "price"
	MetricTokens AgentMetric = "tokens"
	MetricEggs   AgentMetric = "eggs"
)

type AgentCmp string

const (
	CmpAbove AgentCmp = "above"
	CmpBelow AgentCmp = "below"
)

type AgentAction string

const (
	ActBuyEggs  AgentAction = "buy-eggs"
	ActSellEggs AgentAction = "sell-eggs"
	ActOpenCall AgentAction = "open-call"
	ActOpenPut  AgentAction = "open-put"
)

type Agent struct {
	Key       string      `json:"key"`
	Enabled   bool        `json:"enabled"`
	Metric    AgentMetric `json:"metric"`
	Cmp       AgentCmp    `json:"cmp"`
	Threshold float64     `json:"threshold"`
	Action    AgentAction `json:"action"`
	Size      float64     `json:"size"`
}

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

func (s State) SignedTrendStrength() float64 {
	st := s.TrendStrength()
	if s.PriceTrend < 0 {
		return -st
	}
	return st
}

func defaultAgents() []Agent {
	return []Agent{
		{Key: "momentum", Metric: MetricTrend, Cmp: CmpAbove, Threshold: 0.3, Action: ActBuyEggs, Size: 100},
		{Key: "dipbuyer", Metric: MetricPrice, Cmp: CmpBelow, Threshold: 2.0, Action: ActBuyEggs, Size: 100},
		{Key: "profittaker", Metric: MetricPrice, Cmp: CmpAbove, Threshold: 3.0, Action: ActSellEggs, Size: 100},
		{Key: "degencall", Metric: MetricTrend, Cmp: CmpAbove, Threshold: 0.3, Action: ActOpenCall, Size: 50},
		{Key: "degenput", Metric: MetricTrend, Cmp: CmpBelow, Threshold: -0.3, Action: ActOpenPut, Size: 50},
	}
}

type AgentEvent struct {
	Key    string
	Action AgentAction
	Size   float64
}

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
		if a.TradesOptions() {
			if m.s.Level() < SpecUnlockLevel {
				continue
			}
			if len(m.s.Positions) >= agentQueueMax {
				continue
			}
		} else if len(m.s.Transactions) >= agentQueueMax {
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

func (m *Machine) ToggleAgent(i int) {
	if i < 0 || i >= len(m.s.Agents) {
		return
	}
	m.s.Agents[i].Enabled = !m.s.Agents[i].Enabled
}

func (m *Machine) SetAgentSize(i int, size float64) {
	if i < 0 || i >= len(m.s.Agents) || size <= 0 {
		return
	}
	m.s.Agents[i].Size = size
}

func (m *Machine) SetAgentThreshold(i int, t float64) {
	if i < 0 || i >= len(m.s.Agents) {
		return
	}
	m.s.Agents[i].Threshold = t
}
