package layout

import (
	"strings"
	"testing"
)

func fixedPane(name string, interactive bool, pos *GridPos) Pane {
	return Pane{
		Name:        name,
		Interactive: interactive,
		Pos:         pos,
		Render: func(f Frame) string {
			body := name
			if f.Focused {
				body = name + "*"
			}
			if f.Height < 1 {
				return body
			}
			lines := make([]string, f.Height)
			for i := range lines {
				lines[i] = body
			}
			return FitBlock(strings.Join(lines, "\n"), f.Width, f.Height)
		},
	}
}

func lineAt(out string, y int) string {
	lines := strings.Split(out, "\n")
	if y < 0 || y >= len(lines) {
		return ""
	}
	return lines[y]
}

func TestGridTilesTwoColumnsGapFree(t *testing.T) {
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			fixedPane("left", false, &GridPos{Col: 0, Row: 0}),
			fixedPane("right", false, &GridPos{Col: 1, Row: 0}),
		},
	}
	out := scr.Render(Frame{Width: 81, Height: 4}, TierTall, 0)
	rows := strings.Split(out, "\n")
	if len(rows) != 4 {
		t.Fatalf("grid height = %d rows, want 4:\n%s", len(rows), out)
	}
	for _, r := range rows {
		if w := len([]rune(r)); w != 81 {
			t.Fatalf("row width = %d, want 81 (gap-free tiling):\n%q", w, r)
		}
	}
	if !strings.Contains(rows[0], "left") || !strings.Contains(rows[0], "right") {
		t.Fatalf("both columns should occupy row 0:\n%s", out)
	}
	if strings.Index(rows[0], "left") >= strings.Index(rows[0], "right") {
		t.Fatalf("left column should precede right:\n%s", out)
	}
}

func TestGridEdgeTilingHasNoOddWidthGap(t *testing.T) {
	cells := []gridCell{{x: 0, w: 1}, {x: 1, w: 1}, {x: 2, w: 1}}
	total := 0
	prevEnd := 0
	for _, c := range cells {
		r := pixelRect(c, 3, 1, 81, 1)
		if r.x != prevEnd {
			t.Fatalf("column %d starts at %d, want %d (no gap/overlap)", c.x, r.x, prevEnd)
		}
		prevEnd = r.x + r.w
		total += r.w
	}
	if total != 81 {
		t.Fatalf("columns sum to %d, want 81", total)
	}
}

func TestGridRowSpanStacksVertically(t *testing.T) {
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			fixedPane("tall", false, &GridPos{Col: 0, Row: 0, RowSpan: 2}),
			fixedPane("top", false, &GridPos{Col: 1, Row: 0}),
			fixedPane("bot", false, &GridPos{Col: 1, Row: 1}),
		},
	}
	out := scr.Render(Frame{Width: 40, Height: 2}, TierTall, 0)
	if !strings.Contains(lineAt(out, 0), "tall") || !strings.Contains(lineAt(out, 1), "tall") {
		t.Fatalf("row-spanning pane should appear on both rows:\n%s", out)
	}
	if !strings.Contains(lineAt(out, 0), "top") {
		t.Fatalf("top pane missing from row 0:\n%s", out)
	}
	if !strings.Contains(lineAt(out, 1), "bot") {
		t.Fatalf("bot pane missing from row 1:\n%s", out)
	}
	if strings.Contains(lineAt(out, 0), "bot") {
		t.Fatalf("bot pane leaked into row 0:\n%s", out)
	}
}

func TestGridAutoFlowFillsFreeCells(t *testing.T) {
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			fixedPane("a", false, nil),
			fixedPane("b", false, nil),
			fixedPane("c", false, nil),
		},
	}
	out := scr.Render(Frame{Width: 40, Height: 2}, TierTall, 0)
	if !strings.Contains(lineAt(out, 0), "a") || !strings.Contains(lineAt(out, 0), "b") {
		t.Fatalf("a and b should auto-flow into row 0:\n%s", out)
	}
	if !strings.Contains(lineAt(out, 1), "c") {
		t.Fatalf("c should auto-flow into row 1:\n%s", out)
	}
}

func TestGridColSpanClampsToGrid(t *testing.T) {
	c := cellFor(GridPos{Col: 1, ColSpan: 5}, 3)
	if c.x != 1 || c.w != 2 {
		t.Fatalf("cellFor clamp = {x:%d w:%d}, want {x:1 w:2}", c.x, c.w)
	}
}

func TestGridFocusesRingSelection(t *testing.T) {
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			fixedPane("a", true, &GridPos{Col: 0}),
			fixedPane("b", true, &GridPos{Col: 1}),
		},
	}
	out := scr.Render(Frame{Width: 40, Height: 1}, TierTall, 1)
	if !strings.Contains(out, "b*") {
		t.Fatalf("focused pane b should render focused:\n%s", out)
	}
	if strings.Contains(out, "a*") {
		t.Fatalf("unfocused pane a rendered focused:\n%s", out)
	}
}

func TestGridFallsBackToStackWithoutHeight(t *testing.T) {
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			fixedPane("a", false, &GridPos{Col: 0}),
			fixedPane("b", false, &GridPos{Col: 1}),
		},
	}
	out := scr.Render(NewFrame(40), TierTall, 0)
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Fatalf("fallback should still render both panes:\n%s", out)
	}
}

func demoPanel(title string, body ...string) Pane {
	return Pane{
		Title: title,
		Render: func(f Frame) string {
			inner := NewFrame(f.Width - 4)
			return FitBlock(inner.Panel(title, body...), f.Width, f.Height)
		},
	}
}

func TestGridRendersBorderedDashboard(t *testing.T) {
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			demoPanel("STATUS", "tokens 1.2M", "eggs 340", "Lv.7"),
			demoPanel("MARKET", "price 12.4", "trend up", "demand hi"),
			{
				Pos:   &GridPos{Col: 0, Row: 1, ColSpan: 2},
				Title: "FEED",
				Render: func(f Frame) string {
					inner := NewFrame(f.Width - 4)
					return FitBlock(inner.Panel("FEED", "honk", "honk", "sold 50 eggs"), f.Width, f.Height)
				},
			},
		},
	}
	out := scr.Render(Frame{Width: 64, Height: 14}, TierTall, 0)
	rows := strings.Split(out, "\n")
	if len(rows) != 14 {
		t.Fatalf("dashboard height = %d rows, want 14", len(rows))
	}
	for i, r := range rows {
		if w := ansiWidth(r); w != 64 {
			t.Fatalf("row %d width = %d, want 64:\n%q", i, w, r)
		}
	}
	if !strings.Contains(out, "STATUS") || !strings.Contains(out, "MARKET") || !strings.Contains(out, "FEED") {
		t.Fatalf("all three panels should render:\n%s", out)
	}
	t.Log("\n" + out)
}

func ansiWidth(s string) int {
	return len([]rune(stripANSI(s)))
}

func TestFitBlockClipsAndPads(t *testing.T) {
	got := FitBlock("abcdef\nxy", 4, 3)
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("want 3 lines, got %d: %q", len(lines), got)
	}
	if lines[0] != "abcd" {
		t.Fatalf("line 0 = %q, want abcd (clipped)", lines[0])
	}
	if lines[1] != "xy  " {
		t.Fatalf("line 1 = %q, want 'xy  ' (padded)", lines[1])
	}
	if lines[2] != "    " {
		t.Fatalf("line 2 = %q, want 4 spaces (blank row)", lines[2])
	}
}

func TestGridSlimNarrowsAndDonates(t *testing.T) {
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			func() Pane { p := fixedPane("slim", false, &GridPos{Col: 0, Row: 0}); p.Slim = true; return p }(),
			fixedPane("wide", false, &GridPos{Col: 1, Row: 0}),
		},
	}
	out := scr.Render(Frame{Width: 80, Height: 1}, TierTall, 0)
	row := lineAt(out, 0)
	if w := ansiWidth(row); w != 80 {
		t.Fatalf("slim row width = %d, want gap-free 80:\n%q", w, row)
	}
	wideStart := strings.Index(row, "wide")
	if wideStart < 0 {
		t.Fatalf("wide pane missing:\n%q", row)
	}
	if wideStart >= 40 {
		t.Fatalf("slim pane did not narrow: wide starts at %d, want < 40:\n%q", wideStart, row)
	}
	if wideStart < 20 {
		t.Fatalf("slim pane shrank past the 20-col floor: wide starts at %d:\n%q", wideStart, row)
	}
}

func TestGridSlimFloorAt20(t *testing.T) {
	if got := slimWidth(40); got != 20 {
		t.Fatalf("slimWidth(40) = %d, want 20", got)
	}
	if got := slimWidth(30); got != 20 {
		t.Fatalf("slimWidth(30) = %d, want 20 (floored)", got)
	}
	if got := slimWidth(50); got != 25 {
		t.Fatalf("slimWidth(50) = %d, want 25 (half)", got)
	}
	if got := slimWidth(16); got != 16 {
		t.Fatalf("slimWidth(16) = %d, want 16 (already narrower than floor)", got)
	}
}

func TestGridSoleAutoPaneFillsWidth(t *testing.T) {
	scr := Screen{
		Layout: Grid{Cols: 2},
		Panes: []Pane{
			fixedPane("a", false, &GridPos{Col: 0, Row: 0, ColSpan: 2}),
			fixedPane("b", false, nil),
		},
	}
	out := scr.Render(Frame{Width: 60, Height: 2}, TierTall, 0)
	if w := ansiWidth(lineAt(out, 1)); w != 60 {
		t.Fatalf("sole auto pane row width = %d, want 60:\n%s", w, out)
	}
}
