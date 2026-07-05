package game

import (
	"math"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/panels"

	"github.com/codyconfer/goose/internal/content"
)

const (
	vizBands       = 12
	vizHeight      = 6
	vizBase        = 0.28
	vizSwing       = 0.30
	vizSellRefRate = 40.0
	vizPhaseStep   = 0.55
	vizAttack      = 0.50
	vizDecay       = 0.22
	vizPeakFall    = 0.55
)

func (m *Model) updateViz(dt float64) {
	if len(m.vizBins) == 0 {
		return
	}
	s := m.econ.Get()

	boost := clamp01(m.sellRate/vizSellRefRate)*0.45 + m.pulse*0.45
	frozen := s.Frozen()

	phase := float64(m.upBeats) * vizPhaseStep
	for i := range m.vizBins {
		wave := math.Sin(phase*(1+0.3*float64(i)) + float64(i)*0.9)
		target := clamp01(vizBase + vizSwing*wave + boost*(0.5+0.5*wave))
		if frozen {
			target = clamp01(0.05 + 0.05*wave)
		}

		lerp := vizDecay
		if target > m.vizBins[i] {
			lerp = vizAttack
		}
		m.vizBins[i] += (target - m.vizBins[i]) * lerp

		if m.vizBins[i] > m.vizPeaks[i] {
			m.vizPeaks[i] = m.vizBins[i]
		} else if m.vizPeaks[i] = m.vizPeaks[i] - vizPeakFall*dt; m.vizPeaks[i] < m.vizBins[i] {
			m.vizPeaks[i] = m.vizBins[i]
		}
	}
}

func renderSpectrum(m *Model, vk layout.Frame) string {
	return panels.Spectrum(vk, content.Text.Viz.Panel, m.vizBins, vizHeight, content.Text.Viz.Empty,
		panels.SpectrumOpts{Peaks: m.vizPeaks, BarWide: 2, BarGap: 1})
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
