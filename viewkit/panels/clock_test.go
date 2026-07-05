package panels

import (
	"strings"
	"testing"
	"time"

	"github.com/codyconfer/viewkit/layout"
)

var clockT = time.Date(2026, time.July, 4, 13, 5, 9, 0, time.UTC)

func TestClockDefault24h(t *testing.T) {
	out := stripANSI(Clock(layout.DefaultFrame(), "CLOCK", clockT))
	if !strings.Contains(out, "13:05:09") {
		t.Fatalf("want 24h HH:MM:SS, got:\n%s", out)
	}
	if !strings.Contains(out, "CLOCK") {
		t.Errorf("missing title:\n%s", out)
	}
}

func TestClock12hWithDate(t *testing.T) {
	out := stripANSI(Clock(layout.DefaultFrame(), "CLOCK", clockT, ClockOpts{ShowDate: true}))
	for _, want := range []string{"1:05:09 PM", "Jul 4 2026"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
}

func TestClockHideSeconds(t *testing.T) {
	out := stripANSI(Clock(layout.DefaultFrame(), "CLOCK", clockT, ClockOpts{TwentyFour: true, HideSeconds: true}))
	if !strings.Contains(out, "13:05") || strings.Contains(out, "13:05:09") {
		t.Fatalf("want HH:MM only, got:\n%s", out)
	}
}

func TestBinaryClockLitBits(t *testing.T) {

	out := stripANSI(BinaryClock(layout.DefaultFrame(), "BINARY", clockT))
	if !strings.Contains(out, binOn) || !strings.Contains(out, binOff) {
		t.Fatalf("expected both lit and unlit bits:\n%s", out)
	}

	if !strings.Contains(out, "1") || !strings.Contains(out, "3") || !strings.Contains(out, "9") {
		t.Errorf("missing numeric footer digits:\n%s", out)
	}
}

func TestBinaryClockBitCountMatchesPopcount(t *testing.T) {

	zero := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	out := stripANSI(BinaryClock(layout.DefaultFrame(), "B", zero))
	if strings.Contains(out, binOn) {
		t.Fatalf("midnight should have no lit bits:\n%s", out)
	}
}
