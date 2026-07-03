package game

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/panels"
	"github.com/codyconfer/viewkit/theme"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
)

type agentsScreen struct {
	prev   screen
	cursor int
	scroll layout.ScrollState
}

func (as *agentsScreen) simulates() bool { return true }

func (as *agentsScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
	agents := m.econ.Get().Agents
	action, ok := agentsKeymap().Action(msg.String())
	if !ok {
		return nil
	}
	switch action {
	case keys.Quit:
		m.quitting = true
		_ = m.save()
		return tea.Quit
	case keys.Cancel:
		m.screen = as.prev
	case keys.Up:
		as.cursor = panels.StepIndex(as.cursor, -1, len(agents))
	case keys.Down:
		as.cursor = panels.StepIndex(as.cursor, 1, len(agents))
	case keys.Confirm:
		m.econ.ToggleAgent(as.cursor)
	case keys.Left:
		as.cycleSize(m, -1)
	case keys.Right:
		as.cycleSize(m, 1)
	case keys.Dec:
		as.nudgeThreshold(m, -1)
	case keys.Inc:
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

func (as *agentsScreen) panes(m *Model) []layout.Pane {
	agents := m.econ.Get().Agents
	return []layout.Pane{
		{
			Name:        "roster",
			Interactive: len(agents) > 0,
			Render:      func(f layout.Frame) string { return as.renderRoster(m, cellFrame(f), agents) },
		},
	}
}

func (as *agentsScreen) view(m *Model) string {
	vk := m.frame()
	km := agentsKeymap()
	hints := [][2]string{
		km.Hint(keys.Up),
		km.Hint(keys.Confirm),
		km.Hint(keys.Left),
		km.Hint(keys.Inc),
		km.Hint(keys.Cancel),
	}
	body := layout.Screen{Layout: layout.FlexGrid{}, Panes: as.panes(m)}.Render(m.bodyFrame(), m.heightTier(), 0)
	return layout.Stack(
		vk.Header(content.Text.Agents.DeskTitle, content.Text.Agents.Subtitle),
		body,
		panels.Flash(vk.Fit(m.flash)),
		vk.HintLine(hints...),
	)
}

func (as *agentsScreen) renderRoster(m *Model, vk layout.Frame, agents []economy.Agent) string {
	if len(agents) == 0 {
		return vk.Panel(content.Text.Agents.Panel, theme.DimSty.Render(content.Text.Agents.Empty))
	}
	var lines []string
	selStart, selEnd := -1, -1
	for i, a := range agents {
		selected := i == as.cursor
		rc := content.Text.Agents.Roster[a.Key]

		nameSty := theme.ValSty
		if selected {
			nameSty = theme.TitleSty
		}
		name := layout.Cursor(selected) + nameSty.Render(rc.Name)

		status := theme.DimSty.Render(content.Text.Agents.OffWord)
		if a.Enabled {
			status = theme.CanSty.Render(content.Text.Agents.OnWord)
		}

		if selected {
			selStart = len(lines)
		}
		lines = append(lines, vk.Spread(name, status))
		lines = append(lines, theme.DimSty.Width(vk.Width-5).MarginLeft(5).Render(agentRule(a)))
		if selected {
			lines = append(lines, theme.DimSty.Italic(true).Width(vk.Width-5).MarginLeft(5).Render(rc.Blurb))
			selEnd = len(lines) - 1
		}
	}
	rows := m.panelRows(agentRows)
	if selStart >= 0 {
		as.scroll.Reveal(selEnd, len(lines), rows)
		as.scroll.Reveal(selStart, len(lines), rows)
	}
	return vk.ScrollPanel(content.Text.Agents.Panel, lines, rows, as.scroll.Offset)
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
