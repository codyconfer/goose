package panels

import (
	"strings"
	"testing"
)

func TestScrollStateClamp(t *testing.T) {
	cases := []struct {
		name        string
		start       int
		total, rows int
		want        int
	}{
		{"below zero", -5, 20, 8, 0},
		{"past end", 100, 20, 8, 12},
		{"fits window", 3, 5, 8, 0},
		{"in range", 4, 20, 8, 4},
		{"exact last page", 12, 20, 8, 12},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := ScrollState{Offset: tc.start}
			s.clamp(tc.total, tc.rows)
			if s.Offset != tc.want {
				t.Fatalf("clamp(%d,%d) from %d = %d, want %d", tc.total, tc.rows, tc.start, s.Offset, tc.want)
			}
		})
	}
}

func TestScrollStateScrollStopsAtEnds(t *testing.T) {
	s := ScrollState{}

	s.Scroll(-8, 20, 8)
	if s.Offset != 0 {
		t.Fatalf("scroll up from top = %d, want 0", s.Offset)
	}

	for i := 0; i < 5; i++ {
		s.Scroll(8, 20, 8)
	}
	if s.Offset != 12 {
		t.Fatalf("scroll to end = %d, want 12", s.Offset)
	}
}

func TestScrollWindowNoFooterWhenFits(t *testing.T) {
	lines := []string{"a", "b", "c"}
	win, footer, ok := scrollWindow(lines, 8, 0)
	if ok || footer != "" {
		t.Fatalf("expected no footer when list fits, got ok=%v footer=%q", ok, footer)
	}
	if len(win) != 3 {
		t.Fatalf("window len = %d, want 3", len(win))
	}
}

func TestScrollWindowClampsAndReportsPosition(t *testing.T) {
	lines := make([]string, 42)
	for i := range lines {
		lines[i] = string(rune('a' + i%26))
	}

	win, footer, ok := scrollWindow(lines, 8, 999)
	if !ok {
		t.Fatal("expected footer for oversized list")
	}
	if len(win) != 8 {
		t.Fatalf("window len = %d, want 8", len(win))
	}
	if want := "↕ 35–42 of 42"; footer != want {
		t.Fatalf("footer = %q, want %q", footer, want)
	}
}

func TestScrollPanelIncludesFooterLine(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "row"
	}
	out := ScrollPanel("History", lines, 8, 4)
	if !strings.Contains(out, "5–12 of 20") {
		t.Fatalf("panel missing position footer:\n%s", out)
	}
}
