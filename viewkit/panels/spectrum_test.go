package panels

import (
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/layout"
)

func TestSpectrumEmpty(t *testing.T) {
	out := Spectrum(layout.DefaultFrame(), "EQ", nil, 6, "silent")
	if !strings.Contains(out, "silent") {
		t.Fatalf("empty spectrum missing placeholder:\n%s", out)
	}
	if strings.Contains(stripANSI(out), "█") {
		t.Errorf("empty spectrum drew bars:\n%s", out)
	}
}

func TestSpectrumFullAndZeroBands(t *testing.T) {
	const height = 6
	out := stripANSI(Spectrum(layout.DefaultFrame(), "EQ", []float64{0, 1}, height, ""))

	if got := strings.Count(out, "█"); got != height {
		t.Errorf("full band drew %d full cells, want %d:\n%s", got, height, out)
	}

	for _, g := range vBlocks[:7] {
		if strings.Contains(out, g) {
			t.Errorf("full band drew partial glyph %q:\n%s", g, out)
		}
	}
}

func TestSpectrumClampsAndHeight(t *testing.T) {
	const height = 4
	out := stripANSI(Spectrum(layout.DefaultFrame(), "EQ", []float64{5.0}, height, ""))
	if got := strings.Count(out, "█"); got != height {
		t.Errorf("over-unit level drew %d cells, want clamped to %d:\n%s", got, height, out)
	}
}

func TestSpectrumPeakCap(t *testing.T) {
	out := stripANSI(Spectrum(layout.DefaultFrame(), "EQ", []float64{0}, 4, "",
		SpectrumOpts{Peaks: []float64{1}}))
	if !strings.Contains(out, "▔") {
		t.Errorf("held peak did not draw a cap:\n%s", out)
	}
}

func TestSpectrumTrimsToWidth(t *testing.T) {
	levels := make([]float64, 60)
	levels[len(levels)-1] = 1
	out := stripANSI(Spectrum(layout.NewFrame(24), "EQ", levels, 5, ""))
	if strings.Contains(out, "█") {
		t.Errorf("band beyond the width cutoff was not trimmed:\n%s", out)
	}
}

func TestSpectrumCell(t *testing.T) {
	if got := spectrumCell(0, 0, 4); got != "" {
		t.Errorf("zero level cell = %q, want empty", got)
	}
	if got := spectrumCell(1, 0, 4); got != "█" {
		t.Errorf("full level bottom cell = %q, want █", got)
	}
	if got := spectrumCell(1, 3, 4); got != "█" {
		t.Errorf("full level top cell = %q, want █", got)
	}

	if got := spectrumCell(0.5, 3, 4); got != "" {
		t.Errorf("half level top cell = %q, want empty", got)
	}
}

func TestPeakCell(t *testing.T) {
	if got := peakCell(0, 4); got != -1 {
		t.Errorf("zero peak = %d, want -1", got)
	}
	if got := peakCell(1, 4); got != 3 {
		t.Errorf("full peak = %d, want top cell 3", got)
	}
	if peakCell(1, 4) <= peakCell(0.25, 4) {
		t.Error("peak cell should rise with magnitude")
	}
}
