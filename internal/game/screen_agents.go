package game

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/game/viewkit/panels"
	"github.com/codyconfer/goose/internal/game/viewkit/theme"
)

type agentsScreen struct {
	prev   screen
	cursor int
	scroll panels.ScrollState
}

func (as *agentsScreen) simulates() bool { return true }

func (as *agentsScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	agents := m.econ.Get().Agents
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		_ = m.save()
		return tea.Quit
	case "esc", "a", "q":
		m.screen = as.prev
	case "up", "k":
		as.cursor = panels.StepIndex(as.cursor, -1, len(agents))
	case "down", "j":
		as.cursor = panels.StepIndex(as.cursor, 1, len(agents))
	case "enter", " ", "spacebar":
		m.econ.ToggleAgent(as.cursor)
	case "left", "h":
		as.cycleSize(m, -1)
	case "right", "l":
		as.cycleSize(m, 1)
	case "[", "-", "_":
		as.nudgeThreshold(m, -1)
	case "]", "+", "=":
		as.nudgeThreshold(m, 1)
	}
	return nil
}

func (as *agentsScreen) current(m *Model) (economy.Agent, bool) {
	agents := m.econ.Get().Agents
	if as.cursor < 0 || as.cursor >= len(agents) {
		return economy.Agent{}, false
	}
	return agents[as.cursor], true
}

// sizeOptions is the ladder the agent's size cycles through: egg quantities for
// spot trades, premium tiers for options.
func sizeOptions(a economy.Agent) []float64 {
	if a.TradesOptions() {
		return specPremiums
	}
	return tradeSizes
}

func (as *agentsScreen) cycleSize(m *Model, delta int) {
	a, ok := as.current(m)
	if !ok {
		return
	}
	opts := sizeOptions(a)
	if len(opts) == 0 {
		return
	}
	idx := nearestIndex(opts, a.Size)
	idx = panels.StepIndex(idx, delta, len(opts))
	m.econ.SetAgentSize(as.cursor, opts[idx])
}

func (as *agentsScreen) nudgeThreshold(m *Model, delta int) {
	a, ok := as.current(m)
	if !ok {
		return
	}
	next := a.Threshold + float64(delta)*thresholdStep(a.Metric)
	if a.Metric == economy.MetricTrend {
		next = clampTrendThreshold(next)
	} else if next < 0 {
		next = 0
	}
	m.econ.SetAgentThreshold(as.cursor, next)
}

func thresholdStep(metric economy.AgentMetric) float64 {
	switch metric {
	case economy.MetricTrend:
		return 0.05
	case economy.MetricPrice:
		return 0.25
	default:
		return 50
	}
}

func clampTrendThreshold(v float64) float64 {
	if v < -1 {
		return -1
	}
	if v > 1 {
		return 1
	}
	return v
}

func nearestIndex(opts []float64, v float64) int {
	best, bestDist := 0, -1.0
	for i, o := range opts {
		d := o - v
		if d < 0 {
			d = -d
		}
		if bestDist < 0 || d < bestDist {
			best, bestDist = i, d
		}
	}
	return best
}

func (as *agentsScreen) view(m *Model) string {
	vk := m.frame()
	agents := m.econ.Get().Agents

	sections := []string{
		vk.Header(content.Text.Agents.DeskTitle, content.Text.Agents.Subtitle),
		as.renderRoster(m, agents),
		panels.Flash(vk.Fit(m.flash)),
		vk.HintLine(
			[2]string{"↑/↓", "select"},
			[2]string{"enter", "hire/bench"},
			[2]string{"←/→", "size"},
			[2]string{"[ ]", "threshold"},
			[2]string{"ctrl+u/d", "page"},
			[2]string{"esc", "back"},
		),
	}
	return panels.Stack(sections...)
}

func (as *agentsScreen) renderRoster(m *Model, agents []economy.Agent) string {
	vk := m.frame()
	if len(agents) == 0 {
		return vk.Panel(content.Text.Agents.Panel, theme.DimSty.Render(content.Text.Agents.Empty))
	}
	var lines []string
	for i, a := range agents {
		selected := i == as.cursor
		rc := content.Text.Agents.Roster[a.Key]

		nameSty := theme.ValSty
		if selected {
			nameSty = theme.TitleSty
		}
		name := panels.Cursor(selected) + nameSty.Render(rc.Name)

		status := theme.DimSty.Render(content.Text.Agents.OffWord)
		if a.Enabled {
			status = theme.CanSty.Render(content.Text.Agents.OnWord)
		}

		lines = append(lines, vk.Spread(name, status))
		lines = append(lines, theme.DimSty.Width(vk.Width-5).MarginLeft(5).Render(agentRule(a)))
		if selected {
			lines = append(lines, theme.DimSty.Italic(true).Width(vk.Width-5).MarginLeft(5).Render(rc.Blurb))
		}
	}
	if as.cursor >= 0 {
		as.scroll.Reveal(as.cursor, len(lines), agentRows)
	}
	return vk.ScrollPanel(content.Text.Agents.Panel, lines, agentRows, as.scroll.Offset)
}

func agentRule(a economy.Agent) string {
	return fmt.Sprintf(content.Text.Agents.RuleFmt,
		agentMetricLabel(a.Metric),
		agentCmpLabel(a.Cmp),
		agentThresholdLabel(a),
		agentActionLabel(a),
	)
}

func agentMetricLabel(metric economy.AgentMetric) string {
	switch metric {
	case economy.MetricTrend:
		return content.Text.Agents.MetricTrend
	case economy.MetricPrice:
		return content.Text.Agents.MetricPrice
	case economy.MetricTokens:
		return content.Text.Agents.MetricTokens
	case economy.MetricEggs:
		return content.Text.Agents.MetricEggs
	default:
		return string(metric)
	}
}

func agentCmpLabel(cmp economy.AgentCmp) string {
	if cmp == economy.CmpBelow {
		return content.Text.Agents.CmpBelow
	}
	return content.Text.Agents.CmpAbove
}

func agentThresholdLabel(a economy.Agent) string {
	if a.Metric == economy.MetricTrend {
		return fmt.Sprintf("%.0f%%", a.Threshold*100)
	}
	return economy.FormatNum(a.Threshold)
}

func agentActionLabel(a economy.Agent) string {
	size := economy.FormatNum(a.Size)
	switch a.Action {
	case economy.ActBuyEggs:
		return fmt.Sprintf(content.Text.Agents.ActBuyEggs, size)
	case economy.ActSellEggs:
		return fmt.Sprintf(content.Text.Agents.ActSellEggs, size)
	case economy.ActOpenCall:
		return fmt.Sprintf(content.Text.Agents.ActOpenCall, size)
	case economy.ActOpenPut:
		return fmt.Sprintf(content.Text.Agents.ActOpenPut, size)
	default:
		return string(a.Action)
	}
}

func agentFiredMsg(ev economy.AgentEvent) string {
	name := ev.Key
	if rc, ok := content.Text.Agents.Roster[ev.Key]; ok {
		name = rc.Name
	}
	return fmt.Sprintf(content.Text.Agents.FiredFmt, name)
}
