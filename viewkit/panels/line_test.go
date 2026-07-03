package panels

import (
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/layout"
)

func TestLinePlotEmpty(t *testing.T) {
	if got := linePlot(nil, 10, 5, fnum); got != nil {
		t.Fatalf("linePlot(nil) = %v, want nil", got)
	}
}

func TestLinePlotShape(t *testing.T) {
	series := []float64{1, 3, 2, 5, 4}
	const height = 4
	lines := linePlot(series, 10, height, fnum)
	if len(lines) != height+1 {
		t.Fatalf("linePlot returned %d lines, want %d (height + axis)", len(lines), height+1)
	}
	if !strings.Contains(lines[0], "5") {
		t.Errorf("top row missing high label 5: %q", lines[0])
	}
	if !strings.Contains(lines[height-1], "1") {
		t.Errorf("bottom row missing low label 1: %q", lines[height-1])
	}
	joined := stripANSI(strings.Join(lines, "\n"))
	if !strings.Contains(joined, "•") {
		t.Error("line plot missing point markers")
	}
	if !strings.Contains(lines[len(lines)-1], "└") {
		t.Errorf("last line missing x-axis rule: %q", lines[len(lines)-1])
	}
}

func TestLineWrapsWithFooter(t *testing.T) {
	out := Line(layout.DefaultFrame(), "EGGS ON HAND", []float64{1, 2, 3}, 10, 4, fnum, "now: 3")
	if !strings.Contains(out, "EGGS ON HAND") || !strings.Contains(out, "now: 3") {
		t.Errorf("line panel missing title or footer:\n%s", out)
	}
}
