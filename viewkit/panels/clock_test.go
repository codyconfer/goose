package panels

import (
	"math/bits"
	"strconv"
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

func TestClockWorldZones(t *testing.T) {
	west := time.FixedZone("PST", -8*3600)
	out := stripANSI(Clock(layout.DefaultFrame(), "CLOCK", clockT, ClockOpts{
		TwentyFour: true,
		Zones: []ClockZone{
			{Label: "LOCAL", Loc: time.UTC},
			{Label: "WEST", Loc: west},
		},
	}))
	for _, want := range []string{"LOCAL", "13:05:09 UTC", "WEST", "05:05:09 PST"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
}

func TestBinaryClockLitBits(t *testing.T) {

	out := stripANSI(BinaryClock(layout.DefaultFrame(), "BINARY", clockT))
	if !strings.Contains(out, binOn) || !strings.Contains(out, binOff) {
		t.Fatalf("expected both lit and unlit bits:\n%s", out)
	}

	if want := strconv.FormatInt(clockT.Unix(), 10); !strings.Contains(out, want) {
		t.Errorf("missing decimal timestamp footer %q:\n%s", want, out)
	}
}

func TestBinaryClockBitCountMatchesPopcount(t *testing.T) {

	out := stripANSI(BinaryClock(layout.DefaultFrame(), "B", clockT))
	lit := strings.Count(out, binOn)
	want := bits.OnesCount32(uint32(clockT.Unix()))
	if lit != want {
		t.Fatalf("lit bits = %d, want popcount %d:\n%s", lit, want, out)
	}
}
