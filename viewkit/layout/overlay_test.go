package layout

import (
	"strings"
	"testing"
)

func TestOverlayCenters(t *testing.T) {
	bg := strings.Repeat(".....\n", 5)
	bg = strings.TrimRight(bg, "\n")
	fg := "#\n#\n#"

	out := Overlay(bg, fg)
	lines := strings.Split(out, "\n")
	if len(lines) != 5 {
		t.Fatalf("want 5 rows, got %d:\n%s", len(lines), out)
	}

	for r, line := range lines {
		if r >= 1 && r <= 3 {
			if line != "..#.." {
				t.Errorf("row %d = %q, want %q", r, line, "..#..")
			}
		} else if line != "....." {
			t.Errorf("row %d = %q, want untouched dots", r, line)
		}
	}
}

func TestOverlayTopLeft(t *testing.T) {
	bg := "....\n....\n...."
	fg := "AB\nCD"
	out := Overlay(bg, fg, OverlayPos{XFrac: 0, YFrac: 0})
	want := "AB..\nCD..\n...."
	if out != want {
		t.Fatalf("top-left overlay:\ngot:\n%s\nwant:\n%s", out, want)
	}
}

func TestOverlayBottomRight(t *testing.T) {
	bg := "....\n....\n...."
	fg := "XY"
	out := Overlay(bg, fg, OverlayPos{XFrac: 1, YFrac: 1})
	want := "....\n....\n..XY"
	if out != want {
		t.Fatalf("bottom-right overlay:\ngot:\n%s\nwant:\n%s", out, want)
	}
}

func TestOverlayRaggedForegroundPadsToBox(t *testing.T) {
	bg := "........\n........\n........"
	fg := "long\nx"
	out := Overlay(bg, fg, OverlayPos{XFrac: 0, YFrac: 0})
	lines := strings.Split(out, "\n")

	if lines[1] != "x   ...." {
		t.Errorf("ragged row = %q, want %q", lines[1], "x   ....")
	}
}

func TestOverlayLargerThanBackgroundGrows(t *testing.T) {
	bg := "."
	fg := "AA\nBB\nCC"
	out := Overlay(bg, fg)
	if out != "AA\nBB\nCC" {
		t.Fatalf("oversized fg should fill:\n%s", out)
	}
}
