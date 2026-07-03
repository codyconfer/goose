package layout

import (
	"strings"
	"testing"
)

func TestFlexColumnsBreakpoints(t *testing.T) {
	cases := []struct {
		width int
		want  int
	}{
		{240, 3}, {121, 3}, {120, 3},
		{119, 2}, {100, 2}, {80, 2},
		{79, 1}, {60, 1}, {40, 1}, {39, 1}, {1, 1},
	}
	for _, c := range cases {
		if got := FlexColumns(c.width, 40, 3); got != c.want {
			t.Errorf("FlexColumns(%d, 40, 3) = %d, want %d", c.width, got, c.want)
		}
	}
}

func TestFlexColumnsDefaults(t *testing.T) {
	if got := FlexColumns(120, 0, 0); got != 3 {
		t.Fatalf("FlexColumns with zero opts = %d, want 3 (defaults 40/3)", got)
	}
}

func flexBoxPane(name string) Pane {
	return Pane{
		Name: name,
		Render: func(f Frame) string {
			return f.CellBox(name, name)
		},
	}
}

func topBorderCount(out string) int {
	first := out
	if i := strings.IndexByte(out, '\n'); i >= 0 {
		first = out[:i]
	}
	return strings.Count(first, "╭")
}

func TestFlexGridReactiveColumnCount(t *testing.T) {
	scr := Screen{
		Layout: FlexGrid{MinWidth: 40, MaxCols: 3},
		Panes: []Pane{
			flexBoxPane("one"),
			flexBoxPane("two"),
			flexBoxPane("three"),
		},
	}
	cases := []struct {
		width    int
		wantCols int
	}{
		{120, 3},
		{100, 2},
		{60, 1},
	}
	for _, c := range cases {
		out := scr.Render(NewFrame(c.width), TierTall, 0)
		if got := topBorderCount(out); got != c.wantCols {
			t.Fatalf("width %d: rendered %d columns, want %d:\n%s", c.width, got, c.wantCols, out)
		}
		for _, line := range strings.Split(out, "\n") {
			if w := len([]rune(stripANSI(line))); w != c.width {
				t.Fatalf("width %d: row width %d, want gap-free %d:\n%q", c.width, w, c.width, line)
			}
		}
	}
}

func TestFlexGridReflowDemo(t *testing.T) {
	scr := Screen{
		Layout: FlexGrid{MinWidth: 40, MaxCols: 3},
		Panes: []Pane{
			{Name: "status", Render: func(f Frame) string { return f.CellBox("STATUS", "tokens 1.2M", "Lv.7") }},
			{Name: "market", Render: func(f Frame) string { return f.CellBox("MARKET", "price 12.4", "trend up") }},
			{Name: "feed", Render: func(f Frame) string { return f.CellBox("FEED", "honk", "sold 50") }},
		},
	}
	for _, w := range []int{126, 100, 60} {
		t.Logf("\n--- width %d (%d cols) ---\n%s", w, FlexColumns(w, 40, 3), scr.Render(NewFrame(w), TierTall, 0))
	}
}

func TestFlexGridNeverExceedsPaneCount(t *testing.T) {
	scr := Screen{
		Layout: FlexGrid{MinWidth: 40, MaxCols: 3},
		Panes:  []Pane{flexBoxPane("solo")},
	}
	out := scr.Render(NewFrame(200), TierTall, 0) // width wants 3 cols, but only 1 pane
	if topBorderCount(out) != 1 {
		t.Fatalf("single pane must occupy 1 column even on a wide screen:\n%s", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if w := len([]rune(stripANSI(line))); w != 200 {
			t.Fatalf("row width %d, want 200 (full-width single pane):\n%q", w, line)
		}
	}
}

func TestFlexGridStacksAllPanesWhenSingleColumn(t *testing.T) {
	scr := Screen{
		Layout: FlexGrid{MinWidth: 40, MaxCols: 3},
		Panes: []Pane{
			flexBoxPane("alpha"),
			flexBoxPane("beta"),
		},
	}
	out := scr.Render(NewFrame(50), TierTall, 0) // 1 column
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Fatalf("single-column flex should stack all panes:\n%s", out)
	}
	if topBorderCount(out) != 1 {
		t.Fatalf("expected 1 column at width 50:\n%s", out)
	}
}
