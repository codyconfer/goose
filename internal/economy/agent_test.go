package economy

import "testing"

func TestBaselineYield(t *testing.T) {
	m := NewMachine()
	m.BaselineYield()
	if m.s.Tokens != baselineTokens {
		t.Fatalf("tokens=%v, want %v", m.s.Tokens, baselineTokens)
	}
	if m.s.Eggs != baselineEggs {
		t.Fatalf("eggs=%v, want %v", m.s.Eggs, baselineEggs)
	}
	if m.s.TotalEarned != baselineTokens {
		t.Fatalf("total earned=%v, want %v", m.s.TotalEarned, baselineTokens)
	}
}

func TestBaselineYieldFrozenNoop(t *testing.T) {
	m := NewMachine()
	m.s.FreezeSeconds = 10
	m.BaselineYield()
	if m.s.Tokens != 0 || m.s.Eggs != 0 {
		t.Fatalf("frozen baseline changed state: tokens=%v eggs=%v", m.s.Tokens, m.s.Eggs)
	}
}

func TestRunAgentsFiresAndSchedules(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	m.s.Agents = []Agent{
		{Key: "test", Enabled: true, Metric: MetricTokens, Cmp: CmpAbove, Threshold: 0, Action: ActBuyEggs, Size: 10},
	}
	events := m.RunAgents()
	if len(events) != 1 {
		t.Fatalf("events=%d, want 1", len(events))
	}
	if len(m.s.Transactions) != 1 || m.s.Transactions[0].Kind != TxBuyEggs {
		t.Fatalf("transactions=%+v, want one buy-eggs order", m.s.Transactions)
	}
}

func TestRunAgentsDisabledNoop(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	m.s.Agents = []Agent{
		{Key: "test", Enabled: false, Metric: MetricTokens, Cmp: CmpAbove, Threshold: 0, Action: ActBuyEggs, Size: 10},
	}
	if events := m.RunAgents(); len(events) != 0 || len(m.s.Transactions) != 0 {
		t.Fatalf("disabled agent acted: events=%d txns=%d", len(events), len(m.s.Transactions))
	}
}

func TestRunAgentsConditionNotMet(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	m.s.Agents = []Agent{
		{Key: "test", Enabled: true, Metric: MetricTokens, Cmp: CmpBelow, Threshold: 500, Action: ActBuyEggs, Size: 10},
	}
	if events := m.RunAgents(); len(events) != 0 {
		t.Fatalf("agent fired when condition unmet: events=%d", len(events))
	}
}

func TestRunAgentsFrozenNoop(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 1000
	m.s.FreezeSeconds = 10
	m.s.Agents = []Agent{
		{Key: "test", Enabled: true, Metric: MetricTokens, Cmp: CmpAbove, Threshold: 0, Action: ActBuyEggs, Size: 10},
	}
	if events := m.RunAgents(); events != nil || len(m.s.Transactions) != 0 {
		t.Fatalf("frozen agents acted: events=%v txns=%d", events, len(m.s.Transactions))
	}
}

func TestRunAgentsOptionsGatedByLevel(t *testing.T) {
	m := NewMachine()
	m.s.Tokens = 100000
	if m.s.Level() >= SpecUnlockLevel {
		t.Skipf("fresh machine already at spec unlock level %d", m.s.Level())
	}
	m.s.Agents = []Agent{
		{Key: "degen", Enabled: true, Metric: MetricTokens, Cmp: CmpAbove, Threshold: 0, Action: ActOpenCall, Size: 50},
	}
	if events := m.RunAgents(); len(events) != 0 || len(m.s.Positions) != 0 {
		t.Fatalf("option agent opened below unlock level: events=%d positions=%d", len(events), len(m.s.Positions))
	}
}
