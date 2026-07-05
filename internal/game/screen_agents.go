package game

import (
	"fmt"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
)

func sizeOptions(a economy.Agent) []float64 {
	if a.TradesOptions() {
		return specPremiums
	}
	return tradeSizes
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
