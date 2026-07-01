package panels

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestHeaderRendersTitleDetailAndRule(t *testing.T) {
	out := stripANSI(Header("GOOSE", "idle riches"))
	if !strings.Contains(out, "GOOSE") || !strings.Contains(out, "idle riches") {
		t.Fatalf("header missing title or detail:\n%s", out)
	}
	if !strings.Contains(out, "─") {
		t.Fatalf("header missing rule:\n%s", out)
	}
}

func TestStackSkipsEmptySections(t *testing.T) {
	got := Stack("one", "", "two")
	if got != "one\n\ntwo" {
		t.Fatalf("Stack = %q, want standard two-section join", got)
	}
}

func TestMeterClampsToWidth(t *testing.T) {
	out := stripANSI(Meter(2, 4))
	if got := strings.Count(out, "█"); got != 4 {
		t.Fatalf("Meter filled cells = %d, want 4", got)
	}
}

func TestFrameSpreadFitsWidth(t *testing.T) {
	f := NewFrame(24)
	out := f.Spread("a very long label that should not spill", "right")
	if got := ansi.StringWidth(out); got > f.Width {
		t.Fatalf("spread width=%d, want <= %d: %q", got, f.Width, stripANSI(out))
	}
	if !strings.Contains(stripANSI(out), "…") {
		t.Fatalf("spread did not truncate overflowing text: %q", stripANSI(out))
	}

	out = f.Spread(strings.Repeat("l", 12), strings.Repeat("r", 12))
	if got := ansi.StringWidth(out); got > f.Width {
		t.Fatalf("exact spread width=%d, want <= %d: %q", got, f.Width, stripANSI(out))
	}
}

func TestFrameHintLineWrapsToWidth(t *testing.T) {
	out := NewFrame(24).HintLine(
		[2]string{"enter", "choose"},
		[2]string{"pgup/pgdn", "history"},
		[2]string{"esc", "back"},
	)
	if !strings.Contains(out, "\n") {
		t.Fatalf("hint line did not wrap:\n%s", stripANSI(out))
	}
	for _, line := range strings.Split(out, "\n") {
		if got := ansi.StringWidth(line); got > 24 {
			t.Fatalf("hint line width=%d, want <= 24: %q", got, stripANSI(line))
		}
	}
}

func TestFrameSelectableTruncatesLabel(t *testing.T) {
	out := NewFrame(24).Selectable("a very long selectable label that should fit", true)
	if got := ansi.StringWidth(out); got > 24 {
		t.Fatalf("selectable width=%d, want <= 24: %q", got, stripANSI(out))
	}
}
