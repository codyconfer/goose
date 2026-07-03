package layout

import (
	"strings"
	"testing"
)

func pageClamp(body string, rows, offset int) int {
	cr := ScrollableRows(body, rows)
	total := CountLines(ScrollableBody(body, cr))
	if cr <= 0 || total <= cr {
		return 0
	}
	page := ViewportContentRows(cr)
	if page < 1 {
		page = 1
	}
	s := ScrollState{Offset: offset}
	s.Scroll(0, total, page)
	return s.Offset
}

func TestPageScrollRevealsLastContentLine(t *testing.T) {
	content := make([]string, 8)
	for i := range content {
		content[i] = "c" + string(rune('1'+i))
	}
	body := strings.Join(content, "\n") + "\n\nfooter"
	const rows = 8

	maxOffset := pageClamp(body, rows, 999)
	if maxOffset != 4 {
		t.Fatalf("clamped max offset = %d, want 4", maxOffset)
	}

	out := ViewportLayout(body, rows, maxOffset)
	if !strings.Contains(out, "c8") {
		t.Fatalf("last content line c8 not reachable at max offset:\n%s", out)
	}

	oldMax := ScrollState{Offset: 999}
	oldMax.Scroll(0, len(content), ScrollableRows(body, rows)-1)
	if strings.Contains(ViewportLayout(body, rows, oldMax.Offset), "c8") {
		t.Fatalf("expected old clamp (offset=%d) to hide c8, but it was visible", oldMax.Offset)
	}
}

func TestSplitStickyFooter(t *testing.T) {
	content, footer := SplitStickyFooter("body line\n\nfooter line")
	if content != "body line" || footer != "footer line" {
		t.Fatalf("SplitStickyFooter = (%q, %q), want (%q, %q)", content, footer, "body line", "footer line")
	}

	content, footer = SplitStickyFooter("no footer here")
	if content != "no footer here" || footer != "" {
		t.Fatalf("SplitStickyFooter without separator = (%q, %q), want (%q, %q)", content, footer, "no footer here", "")
	}

	content, footer = SplitStickyFooter("a\n\nb\n\nc")
	if content != "a\n\nb" || footer != "c" {
		t.Fatalf("SplitStickyFooter multi = (%q, %q), want (%q, %q)", content, footer, "a\n\nb", "c")
	}
}

func TestCountLines(t *testing.T) {
	cases := map[string]int{
		"":           1,
		"one":        1,
		"one\ntwo":   2,
		"a\nb\nc":    3,
		"trailing\n": 2,
	}
	for in, want := range cases {
		if got := CountLines(in); got != want {
			t.Errorf("CountLines(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestPadLines(t *testing.T) {
	got := PadLines("a\nb", 4)
	if CountLines(got) != 4 {
		t.Errorf("PadLines to 4 rows produced %d rows:\n%q", CountLines(got), got)
	}
	if !strings.HasPrefix(got, "a\nb") {
		t.Errorf("PadLines must preserve original body: %q", got)
	}

	if got := PadLines("a\nb\nc", 2); got != "a\nb\nc" {
		t.Errorf("PadLines over budget = %q, want unchanged", got)
	}

	if got := PadLines("", 3); CountLines(got) != 3 {
		t.Errorf("PadLines empty to 3 rows = %q (%d rows), want 3", got, CountLines(got))
	}
}

func TestViewportLayoutNoFooterScrollsWholeBody(t *testing.T) {
	body := "l1\nl2\nl3\nl4\nl5"
	got := ViewportLayout(body, 3, 0)
	if !strings.Contains(got, "l1") {
		t.Errorf("expected top of body visible:\n%s", got)
	}
	if strings.Contains(got, "l5") {
		t.Errorf("bottom should be clipped by the 3-row viewport:\n%s", got)
	}
}

func TestViewportLayoutPinsFooter(t *testing.T) {
	body := "c1\nc2\nc3\nc4\nc5\n\nFOOTER"
	got := ViewportLayout(body, 5, 0)
	lines := strings.Split(got, "\n")
	if lines[len(lines)-1] != "FOOTER" {
		t.Errorf("footer must be pinned to the last row:\n%s", got)
	}
	if CountLines(got) != 5 {
		t.Errorf("layout should fill exactly 5 rows, got %d:\n%s", CountLines(got), got)
	}
}

func TestScrollableRowsAndBody(t *testing.T) {
	body := "c1\nc2\nc3\n\nFOOTER"

	if got := ScrollableRows("just content", 6); got != 6 {
		t.Errorf("ScrollableRows without footer = %d, want 6", got)
	}
	if got := ScrollableBody("just content", 6); got != "just content" {
		t.Errorf("ScrollableBody without footer = %q, want whole body", got)
	}

	if got := ScrollableRows(body, 6); got != 4 {
		t.Errorf("ScrollableRows with footer = %d, want 4", got)
	}
	if got := ScrollableBody(body, 6); got != "c1\nc2\nc3" {
		t.Errorf("ScrollableBody with footer = %q, want content region", got)
	}

	if got := ScrollableRows(body, 1); got != 0 {
		t.Errorf("ScrollableRows when footer fills viewport = %d, want 0", got)
	}
	if got := ScrollableBody(body, 1); got != "" {
		t.Errorf("ScrollableBody when footer fills viewport = %q, want empty", got)
	}
}
