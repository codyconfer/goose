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
		if got := FlexColCount(c.width, 40, 3); got != c.want {
			t.Errorf("FlexColCount(%d, 40, 3) = %d, want %d", c.width, got, c.want)
		}
	}
}

func TestFlexColumnsDefaults(t *testing.T) {
	if got := FlexColCount(120, 0, 0); got != 3 {
		t.Fatalf("FlexColumns with zero opts = %d, want 3 (defaults 40/4, width still fits 3)", got)
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
		Layout: FlexColumns{MinWidth: 40, MaxCols: 3},
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
		Layout: FlexColumns{MinWidth: 40, MaxCols: 3},
		Panes: []Pane{
			{Name: "status", Render: func(f Frame) string { return f.CellBox("STATUS", "tokens 1.2M", "Lv.7") }},
			{Name: "market", Render: func(f Frame) string { return f.CellBox("MARKET", "price 12.4", "trend up") }},
			{Name: "feed", Render: func(f Frame) string { return f.CellBox("FEED", "honk", "sold 50") }},
		},
	}
	for _, w := range []int{126, 100, 60} {
		t.Logf("\n--- width %d (%d cols) ---\n%s", w, FlexColCount(w, 40, 3), scr.Render(NewFrame(w), TierTall, 0))
	}
}

func TestFlexGridNeverExceedsPaneCount(t *testing.T) {
	scr := Screen{
		Layout: FlexColumns{MinWidth: 40, MaxCols: 3},
		Panes:  []Pane{flexBoxPane("solo")},
	}
	out := scr.Render(NewFrame(200), TierTall, 0)
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
		Layout: FlexColumns{MinWidth: 40, MaxCols: 3},
		Panes: []Pane{
			flexBoxPane("alpha"),
			flexBoxPane("beta"),
		},
	}
	out := scr.Render(NewFrame(50), TierTall, 0)
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Fatalf("single-column flex should stack all panes:\n%s", out)
	}
	if topBorderCount(out) != 1 {
		t.Fatalf("expected 1 column at width 50:\n%s", out)
	}
}

func TestFlexRowsExpandsLoneLastRow(t *testing.T) {
	scr := Screen{
		Layout: FlexRows{MinWidth: 40, MaxCols: 2},
		Panes:  []Pane{flexBoxPane("a"), flexBoxPane("b"), flexBoxPane("c")},
	}
	const w = 120
	out := scr.Render(NewFrame(w), TierTall, 0)
	lines := strings.Split(out, "\n")

	// Every composited row must be gap-free full width (blank separator
	// lines between rows come from Stack and are skipped).
	for _, ln := range lines {
		if ln == "" {
			continue
		}
		if got := len([]rune(stripANSI(ln))); got != w {
			t.Fatalf("row width %d, want %d:\n%q", got, w, ln)
		}
	}

	// First row holds two boxes side by side; the ragged last row holds one
	// box that expands to fill the full width.
	firstTop := stripANSI(lines[0])
	if strings.Count(firstTop, "╭") != 2 {
		t.Fatalf("first row should hold two boxes, got: %q", firstTop)
	}
	var lastTop string
	for _, ln := range lines {
		if s := stripANSI(ln); strings.Contains(s, "╭") {
			lastTop = s
		}
	}
	if strings.Count(lastTop, "╭") != 1 {
		t.Fatalf("last row should hold exactly one box, got: %q", lastTop)
	}
	if !strings.HasPrefix(lastTop, "╭") || !strings.HasSuffix(lastTop, "╮") {
		t.Fatalf("lone last-row pane should span full width: %q", lastTop)
	}
}

func TestFlexRowsReflowDemo(t *testing.T) {
	scr := Screen{
		Layout: FlexRows{MinWidth: 40, MaxCols: 3},
		Panes: []Pane{
			{Name: "status", Render: func(f Frame) string { return f.CellBox("STATUS", "tokens 1.2M", "Lv.7") }},
			{Name: "market", Render: func(f Frame) string { return f.CellBox("MARKET", "price 12.4") }},
			{Name: "feed", Render: func(f Frame) string { return f.CellBox("FEED", "honk", "sold 50") }},
			{Name: "orders", Render: func(f Frame) string { return f.CellBox("ORDERS", "buy 10") }},
		},
	}
	for _, w := range []int{126, 100, 60} {
		t.Logf("\n--- width %d (%d across) ---\n%s", w, FlexColCount(w, 40, 3), scr.Render(NewFrame(w), TierTall, 0))
	}
}

func TestFlexDefaultAllowsFourColumns(t *testing.T) {
	if got := FlexColCount(160, 0, 0); got != 4 {
		t.Fatalf("FlexColCount(160,0,0) = %d, want 4 with default max 4", got)
	}
	if got := FlexColCount(400, 0, 0); got != 4 {
		t.Fatalf("FlexColCount(400,0,0) = %d, want 4 (capped)", got)
	}
}
